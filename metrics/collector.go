package metrics

import (
	"fmt"
	"sync"

	"github.com/cashapp/blip"
	"github.com/cashapp/blip/event"
	"github.com/cashapp/blip/metrics/innodb"
	"github.com/cashapp/blip/metrics/size"
	"github.com/cashapp/blip/metrics/status"
	sysvar "github.com/cashapp/blip/metrics/var"
)

// Register registers a factory that makes one or more collector by domain name.
// This is function is one several integration points because it allows users
// to plug in new metric collectors by providing a factory to make them.
// Blip calls this function in an init function to register the built-in metric
// collectors.
//
// See types in the blip package for more details.
func Register(domain string, f blip.CollectorFactory) error {
	r.Lock()
	defer r.Unlock()
	_, ok := r.factory[domain]
	if ok && blip.Strict {
		return fmt.Errorf("%s already registered", domain)
	}
	r.factory[domain] = f
	event.Sendf(event.REGISTER_METRICS, domain)
	return nil
}

// List lists all registered metric collectors. It is used by the server API
// for GET /registered.
func List() []string {
	r.Lock()
	defer r.Unlock()
	names := []string{}
	for k := range r.factory {
		names = append(names, k)
	}
	return names
}

// Make makes a metric collector for the domain using a previously registered factory.
//
// See types in the blip package for more details.
func Make(domain string, args blip.CollectorFactoryArgs) (blip.Collector, error) {
	r.Lock()
	defer r.Unlock()
	f, ok := r.factory[domain]
	if !ok {
		return nil, fmt.Errorf("%s not registeres", domain)
	}
	return f.Make(domain, args)
}

// --------------------------------------------------------------------------

// Register built-in collectors using built-in factories.
func init() {
	for _, mc := range builtinCollectors {
		Register(mc, f)
	}
}

// repo holds registered blip.CollectorFactory. There's a single package
// instance below.
type repo struct {
	*sync.Mutex
	factory map[string]blip.CollectorFactory
}

// Internal package instance of repo that holds all collector factories registered
// by calls to Register, which includes the built-in factories.
var r = &repo{
	Mutex:   &sync.Mutex{},
	factory: map[string]blip.CollectorFactory{},
}

// factory is the built-in factory for creating all built-in collectors.
// There's a single package instance below. It implements blip.CollectorFactory.
type factory struct{}

var _ blip.CollectorFactory = &factory{}

// Internet package instance of factory that makes all built-it collectors.
// This factory is registered in the init func above.
var f = factory{}

// Make makes a metric collector for the domain. This is the built-in factory
// that makes the built-in collectors: status.global, var.global, and so on.
func (f factory) Make(domain string, args blip.CollectorFactoryArgs) (blip.Collector, error) {
	switch domain {
	case "status.global":
		mc := status.NewGlobal(args.DB)
		return mc, nil
	case "var.global":
		mc := sysvar.NewGlobal(args.DB)
		return mc, nil
	case "size.data":
		mc := size.NewData(args.DB)
		return mc, nil
	case "size.binlogs":
		mc := size.NewBinlogs(args.DB)
		return mc, nil
	case "innodb":
		mc := innodb.NewMetrics(args.DB)
		return mc, nil
	}
	return nil, fmt.Errorf("collector for domain %s not registered", domain)
}

// List of built-in collectors. To add one, add its domain name here, and add
// the same domain in the switch statement above (in factory.Make).
var builtinCollectors = []string{
	"status.global",
	"var.global",
	"size.data",
	"size.binlogs",
	"innodb",
}
