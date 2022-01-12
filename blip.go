// Package blip provides high-level data structs and const for integrating with Blip.
package blip

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
)

const VERSION = "0.0.0"

var SHA = ""

// Metric types.
const (
	UNKNOWN byte = iota
	COUNTER
	GAUGE
	BOOL
	EVENT
)

// Metrics are metrics collected for one plan level, from one database instance.
type Metrics struct {
	Begin     time.Time                // when collection started
	End       time.Time                // when collection completed
	MonitorId string                   // ID of monitor (MySQL)
	Plan      string                   // plan name
	Level     string                   // level name
	State     string                   // state of monitor
	Values    map[string][]MetricValue // keyed on domain
}

// MetricValue is one metric and its name, type, value, and tags. Tags are optional;
// the other fields are required and always set. This is the lowest-level data struct:
// a Collector reports metric values, which the monitor.Engine organize into Metrics
// by adding the appropriate metadata.
type MetricValue struct {
	// Name is the domain-specific metric name, like threads_running from the
	// status.global collector. Names are lowercase but otherwise not modified
	// (for example, hyphens and underscores are not changed).
	Name string

	// Value is the value of the metric. String values are not supported.
	// Boolean values are reported as 0 and 1.
	Value float64

	// Type is the metric type: COUNTER, COUNTER, and other const.
	Type byte

	// Group is the set of name-value pairs that determine the group to which
	// the metric value belongs. Only certain domains group metrics.
	Group map[string]string

	// Meta is optional key-value pairs that annotate or describe the metric value.
	Meta map[string]string
}

// Sink sends metrics to an external destination.
type Sink interface {
	// Send sends metrics to the sink. It must respect the context timeout, if any.
	Send(context.Context, *Metrics) error
	Status() string
}

type SinkFactory interface {
	Make(name, monitorId string, opts, tags map[string]string) (Sink, error)
}

// Plugins are function callbacks that let you override specific functionality of Blip.
// Every plugin is optional: if specified, it overrides the built-in functionality.
type Plugins struct {
	LoadConfig       func(Config) (Config, error)
	LoadMonitors     func(Config) ([]ConfigMonitor, error)
	LoadLevelPlans   func(ConfigPlans) ([]Plan, error)
	ModifyDB         func(*sql.DB)
	TransformMetrics func(*Metrics) error
}

// Factories are interfaces that let you override certain object creation of Blip.
// Every factory is optional: if specified, it overrides the built-in factory.
type Factories struct {
	AWSConfig  AWSConfigFactory
	DbConn     DbFactory
	HTTPClient HTTPClientFactory
}

// Env is the startup environment: command line args and environment variables.
// This is mostly used for testing to override the defaults.
type Env struct {
	Args []string
	Env  []string
}

type AWS struct {
	Region string
}

type AWSConfigFactory interface {
	Make(AWS) (aws.Config, error)
}

type DbFactory interface {
	Make(ConfigMonitor) (*sql.DB, string, error)
}

type HTTPClientFactory interface {
	Make(cfg ConfigHTTP, usedFor string) (*http.Client, error)
}

// Monitor states used by level plan adjuster (LPA).
const (
	STATE_NONE      = ""
	STATE_OFFLINE   = "offline"
	STATE_STANDBY   = "standby"
	STATE_READ_ONLY = "read-only"
	STATE_ACTIVE    = "active"
)

var (
	Strict    = false
	Debugging = false
	debugLog  = log.New(os.Stderr, "", log.LstdFlags|log.Lmicroseconds)
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
}

func Debug(msg string, v ...interface{}) {
	if !Debugging {
		return
	}
	_, file, line, _ := runtime.Caller(1)
	msg = fmt.Sprintf("DEBUG %s:%d %s", path.Base(file), line, msg)
	debugLog.Printf(msg, v...)
}

// True returns true if b is non-nil and true.
// This is convenience function related to *bool files in config structs,
// which is required for knowing when a bool config is explicitily set
// or not. If set, it's not changed; if not, it's set to the default value.
// That makes a good config experience but a less than ideal code experience
// because !*b will panic if b is nil, hence the need for this func.
func True(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

func Bool(s string) bool {
	v := strings.ToLower(s)
	return v == "true" || v == "yes" || v == "enable" || v == "enabled"
}

func MonitorId(cfg ConfigMonitor) string {
	switch {
	case cfg.MonitorId != "":
		return cfg.MonitorId
	case cfg.Hostname != "":
		return cfg.Hostname
	case cfg.Socket != "":
		return cfg.Socket
	}
	return ""
}

// SetOrDefault returns a if not empty, else it returns b. This is a convenience
// function to define variables with an explicit value or a DEFAULT_* value.
func SetOrDefault(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
