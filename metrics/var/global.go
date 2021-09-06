package sysvar

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"

	"github.com/square/blip"
	"github.com/square/blip/collect"
)

const (
	OPT_SOURCE    = "source"
	SOURCE_SELECT = "select"
	SOURCE_PFS    = "pfs"
	SOURCE_SHOW   = "show"
)

var validMetricRegex = regexp.MustCompile("^[a-zA-Z0-9_-]*$")

// Global collects global system variables for the var.global domain.
type Global struct {
	db       *sql.DB
	plans    collect.Plan
	domain   string
	workIn   map[string][]string
	queryIn  map[string]string
	sourceIn map[string]string
}

func NewGlobal(db *sql.DB) *Global {
	return &Global{
		db:       db,
		domain:   "var.global",
		workIn:   map[string][]string{},
		queryIn:  make(map[string]string),
		sourceIn: make(map[string]string),
	}
}

func (c *Global) Domain() string {
	return c.domain
}

func (c *Global) Help() collect.Help {
	return collect.Help{
		Domain:      c.domain,
		Description: "Collect global status variables (sysvars)",
		Options: [][]string{
			{
				OPT_SOURCE,
				"Where to collect sysvars from",
				"auto (auto-determine best source); select (@@GLOBAL.metric_name); pfs (performance_schema.global_variables); show (SHOW GLOBAL STATUS)",
			},
		},
	}
}

// Prepares queries for all levels in the plan that contain the "var.global" domain
func (c *Global) Prepare(plan collect.Plan) error {
LEVEL:
	for levelName, level := range plan.Levels {
		dom, ok := level.Collect[c.domain]
		if !ok {
			// This domain not collected in this level
			continue LEVEL
		}
		err := c.prepareLevel(levelName, dom.Metrics, dom.Options)
		if err != nil {
			// return early with error even if preparing a single level fails
			return err
		}
	}
	return nil
}

func (c *Global) Collect(ctx context.Context, levelName string) ([]blip.MetricValue, error) {
	switch c.sourceIn[levelName] {
	case SOURCE_SELECT:
		return c.collectSELECT(ctx, levelName)
	case SOURCE_PFS:
		return c.collectPFS(ctx, levelName)
	case SOURCE_SHOW:
		return c.collectSHOW(ctx, levelName)
	}

	errorStr := fmt.Sprintf("invalid source in Collect %s", c.sourceIn[levelName])
	panic(errorStr)
}

// //////////////////////////////////////////////////////////////////////////
// Internal methods
// //////////////////////////////////////////////////////////////////////////

// Prepares the query for given level based on it's metrics and source option
func (c *Global) prepareLevel(levelName string, metrics []string, options map[string]string) error {

	// Reset in case because prepareLevel can be called multiple times
	// if the LPA changes the plan
	c.sourceIn[levelName] = ""
	c.queryIn[levelName] = ""
	c.workIn[levelName] = []string{}

	// Validate the metricnames for the level
	err := validateMetricNames(metrics)
	if err != nil {
		return err
	}

	// Save metrics to collect for this level
	c.workIn[levelName] = append(c.workIn[levelName], metrics...)

	// -------------------------------------------------------------------------
	// Manual source
	// -------------------------------------------------------------------------

	// If user specified a method, use only that method, whether it works or not
	if options != nil {
		src := options[OPT_SOURCE]

		if len(src) > 0 && src != "auto" {
			switch src {
			case SOURCE_SELECT:
				return c.prepareSELECT(levelName)
			case SOURCE_PFS:
				return c.preparePFS(levelName)
			case SOURCE_SHOW:
				return c.prepareSHOW(levelName)
			default:
				return fmt.Errorf("invalid source: %s; valid values: auto, select, pfs, show", src)
			}
		}
	}

	// -------------------------------------------------------------------------
	// Auto source (default)
	// -------------------------------------------------------------------------

	if err = c.prepareSELECT(levelName); err == nil {
		return nil
	}

	if err = c.preparePFS(levelName); err == nil {
		return nil
	}

	if err = c.prepareSHOW(levelName); err == nil {
		return nil
	}

	return fmt.Errorf("auto source failed, last error: %s", err)
}

// Validate input metric names to make sure there won't be any
// sql injection attacks.
func validateMetricNames(metricNames []string) error {
	for _, name := range metricNames {
		if !validMetricRegex.MatchString(name) {
			return fmt.Errorf("%s metric isn't a valid metric name", name)
		}
	}
	return nil
}

func (c *Global) prepareSELECT(levelName string) error {
	var globalMetrics = make([]string, len(c.workIn[levelName]))

	for i, str := range c.workIn[levelName] {
		globalMetrics[i] = fmt.Sprintf("@@GLOBAL.%s", str)
	}
	globalMetricString := strings.Join(globalMetrics, ", ")

	c.queryIn[levelName] = fmt.Sprintf("SELECT CONCAT_WS(',', %s) v;", globalMetricString)
	c.sourceIn[levelName] = SOURCE_SELECT

	// Try collecting, discard metrics
	_, err := c.collectSELECT(context.TODO(), levelName)
	return err
}

func (c *Global) collectSELECT(ctx context.Context, levelName string) ([]blip.MetricValue, error) {
	rows, err := c.db.QueryContext(ctx, c.queryIn[levelName])
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	metrics := []blip.MetricValue{}
	for rows.Next() {
		var val string
		if err := rows.Scan(&val); err != nil {
			log.Println(err)
			// Log error and continue to next row to retrieve next metric
			continue
		}

		values := strings.Split(val, ",")
		for idx, name := range c.workIn[levelName] {
			s, err := strconv.ParseFloat(values[idx], 64)
			if err != nil {
				log.Printf("Error parsing the metric: %s value: %s as float %s", name, val, err)
				// Log error and continue to next row to retrieve next metric
				continue
			}
			metrics = append(metrics, blip.MetricValue{
				Name:  name,
				Value: s,
				Type:  blip.GAUGE,
			})
		}
	}

	return metrics, nil
}

func (c *Global) preparePFS(levelName string) error {
	var metricString string
	metricString = strings.Join(c.workIn[levelName], "', '")

	query := fmt.Sprintf("SELECT variable_name, variable_value from performance_schema.global_variables WHERE variable_name in ('%s');",
		metricString,
	)
	c.queryIn[levelName] = query
	c.sourceIn[levelName] = SOURCE_PFS

	// Try collecting, discard metrics
	_, err := c.collectPFS(context.TODO(), levelName)
	return err
}

func (c *Global) collectPFS(ctx context.Context, levelName string) ([]blip.MetricValue, error) {
	return c.collectSHOWorPFS(ctx, levelName)
}

func (c *Global) prepareSHOW(levelName string) error {
	metricString := strings.Join(c.workIn[levelName], "', '")
	query := fmt.Sprintf("SHOW GLOBAL VARIABLES WHERE variable_name in ('%s');", metricString)

	c.queryIn[levelName] = query
	c.sourceIn[levelName] = SOURCE_SHOW

	// Try collecting, discard metrics
	_, err := c.collectPFS(context.TODO(), levelName)
	return err
}

func (c *Global) collectSHOW(ctx context.Context, levelName string) ([]blip.MetricValue, error) {
	return c.collectSHOWorPFS(ctx, levelName)
}

// Since both `show` and `pfs` queries return results in same format (ie; 2 columns, name and value)
// use the same logic for querying and retrieving metrics from the results
func (c *Global) collectSHOWorPFS(ctx context.Context, levelName string) ([]blip.MetricValue, error) {
	rows, err := c.db.QueryContext(ctx, c.queryIn[levelName])
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	metrics := []blip.MetricValue{}
	for rows.Next() {
		m := blip.MetricValue{Type: blip.GAUGE}
		var val string
		if err := rows.Scan(&m.Name, &val); err != nil {
			log.Printf("Error scanning row %s", err)
			// Log error and continue to next row to retrieve next metric
			continue
		}

		s, err := strconv.ParseFloat(val, 64)
		if err != nil {
			log.Printf("Error parsing the metric: %s value: %s as float %s", m.Name, val, err)
			// Log error and continue to next row to retrieve next metric
			continue
		}
		m.Value = s
		metrics = append(metrics, m)
	}

	return metrics, err
}
