package main

import (
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"

	"github.com/sensu-community/sensu-plugin-sdk/sensu"
	"github.com/sensu/sensu-go/types"

	"github.com/sardinasystems/sensu-go-prometheus-metric-check/utils"
)

// Config represents the check plugin config.
type Config struct {
	sensu.PluginConfig
	PrometheusURL    string
	Query            string
	WarningRangeStr  string
	CriticalRangeStr string
	CriticalRange    utils.NagiosRange
	WarningRange     utils.NagiosRange
	EmitPerfdata     bool
	PerfdataName     string
}

var (
	plugin = Config{
		PluginConfig: sensu.PluginConfig{
			Name:     "sensu-go-prometheus-metric-check",
			Short:    "plugin to query prometheus for alerting",
			Keyspace: "sensu.io/plugins/sensu-go-prometheus-metric-check/config",
		},
	}

	options = []*sensu.PluginConfigOption{
		&sensu.PluginConfigOption{
			Path:      "host",
			Env:       "PROMETHEUS_HOST",
			Argument:  "host",
			Shorthand: "H",
			Default:   "http://127.0.0.1:9090",
			Usage:     "Host URL to access Prometheus",
			Value:     &plugin.PrometheusURL,
		},
		&sensu.PluginConfigOption{
			Path:      "query",
			Env:       "PROMETHEUS_QUERY",
			Argument:  "query",
			Shorthand: "q",
			Default:   "",
			Usage:     "PromQL query",
			Value:     &plugin.Query,
		},
		&sensu.PluginConfigOption{
			Path:      "method",
			Env:       "PROMETHEUS_METHOD",
			Argument:  "method",
			Shorthand: "m",
			Default:   "ge",
			Usage:     "PromQL query",
			Value:     &plugin.Query,
		},
		&sensu.PluginConfigOption{
			Path:      "warning",
			Env:       "PROMETHEUS_WARNING",
			Argument:  "warning",
			Shorthand: "w",
			Default:   0.0,
			Usage:     "Warning level",
			Value:     &plugin.WarningRangeStr,
		},
		&sensu.PluginConfigOption{
			Path:      "critical",
			Env:       "PROMETHEUS_CRITICAL",
			Argument:  "critical",
			Shorthand: "c",
			Default:   0.0,
			Usage:     "Critical level",
			Value:     &plugin.CriticalRangeStr,
		},
		&sensu.PluginConfigOption{
			Path:      "emit_perfdata",
			Env:       "PROMETHEUS_EMIT_PERFDATA",
			Argument:  "emit-perfdata",
			Shorthand: "p",
			Default:   true,
			Usage:     "Add perfdata to check output",
			Value:     &plugin.EmitPerfdata,
		},
		&sensu.PluginConfigOption{
			Path:      "name",
			Env:       "PROMETHEUS_NAME",
			Argument:  "name",
			Shorthand: "n",
			Default:   "",
			Usage:     "Perfata name",
			Value:     &plugin.PerfdataName,
		},
	}
)

func main() {
	check := sensu.NewGoCheck(&plugin.PluginConfig, options, checkArgs, executeCheck, false)
	check.Execute()
}

func checkArgs(event *types.Event) (int, error) {
	var err error

	plugin.WarningRange, err = utils.ParseRange(plugin.WarningRangeStr)
	if err != nil {
		return sensu.CheckStateCritical, fmt.Errorf("--warning error: %v", err)
	}

	plugin.CriticalRange, err = utils.ParseRange(plugin.CriticalRangeStr)
	if err != nil {
		return sensu.CheckStateCritical, fmt.Errorf("--critical error: %v", err)
	}

	return sensu.CheckStateOK, nil
}

func executeCheck(event *types.Event) (int, error) {
	return sensu.CheckStateOK, nil
}
