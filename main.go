package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/prometheus/client_golang/api"
	"github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/sensu-community/sensu-plugin-sdk/sensu"
	"github.com/sensu/sensu-go/types"

	"github.com/sardinasystems/sensu-go-prometheus-metric-check/utils"
)

// Config represents the check plugin config.
type Config struct {
	sensu.PluginConfig
	PrometheusURL     string
	Query             string
	WarningStr        string
	CriticalStr       string
	CriticalThreshold utils.NagiosThreshold
	WarningThreshold  utils.NagiosThreshold
	EmitPerfdata      bool
	PerfdataName      string
	DebugQuery        bool
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
		{
			Path:      "host",
			Env:       "PROMETHEUS_HOST",
			Argument:  "host",
			Shorthand: "H",
			Default:   "http://127.0.0.1:9090",
			Usage:     "Host URL to access Prometheus",
			Value:     &plugin.PrometheusURL,
		},
		{
			Path:      "query",
			Env:       "PROMETHEUS_QUERY",
			Argument:  "query",
			Shorthand: "q",
			Default:   "",
			Usage:     "PromQL query",
			Value:     &plugin.Query,
		},
		{
			Path:      "warning",
			Env:       "PROMETHEUS_WARNING",
			Argument:  "warning",
			Shorthand: "w",
			Default:   "",
			Usage:     "Warning level",
			Value:     &plugin.WarningStr,
		},
		{
			Path:      "critical",
			Env:       "PROMETHEUS_CRITICAL",
			Argument:  "critical",
			Shorthand: "c",
			Default:   "",
			Usage:     "Critical level",
			Value:     &plugin.CriticalStr,
		},
		{
			Path:      "emit_perfdata",
			Env:       "PROMETHEUS_EMIT_PERFDATA",
			Argument:  "emit-perfdata",
			Shorthand: "p",
			Default:   true,
			Usage:     "Add perfdata to check output",
			Value:     &plugin.EmitPerfdata,
		},
		{
			Path:      "name",
			Env:       "PROMETHEUS_NAME",
			Argument:  "name",
			Shorthand: "n",
			Default:   "",
			Usage:     "Perfata name",
			Value:     &plugin.PerfdataName,
		},
		{
			Path:      "debug_query",
			Env:       "PROMETHEUS_DEBUG_QUERY",
			Argument:  "debug-query",
			Shorthand: "i",
			Default:   false,
			Usage:     "Enable debug output for query",
			Value:     &plugin.DebugQuery,
		},
	}
)

func main() {
	check := sensu.NewGoCheck(&plugin.PluginConfig, options, checkArgs, executeCheck, false)
	check.Execute()
}

func checkArgs(event *types.Event) (int, error) {
	var err error

	plugin.WarningThreshold, err = utils.ParseThreshold(plugin.WarningStr)
	if err != nil {
		return sensu.CheckStateCritical, fmt.Errorf("--warning error: %v", err)
	}

	plugin.CriticalThreshold, err = utils.ParseThreshold(plugin.CriticalStr)
	if err != nil {
		return sensu.CheckStateCritical, fmt.Errorf("--critical error: %v", err)
	}

	if plugin.PerfdataName == "" {
		qn := fmt.Sprintf("query_%s", plugin.Query)
		for _, rs := range []string{"-", "{", "}", "(", ")", "=", "."} {
			qn = strings.ReplaceAll(qn, rs, "_")
		}

		plugin.PerfdataName = qn
	}

	return sensu.CheckStateOK, nil
}

func executeCheck(event *types.Event) (int, error) {

	apicfg := api.Config{
		Address: plugin.PrometheusURL,
	}

	client, err := api.NewClient(apicfg)
	if err != nil {
		return sensu.CheckStateUnknown, fmt.Errorf("failed to create prometheus api client: %v", err)
	}

	v1api := v1.NewAPI(client)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, warnings, err := v1api.Query(ctx, plugin.Query, time.Now())
	if err != nil {
		return sensu.CheckStateUnknown, fmt.Errorf("query error: %v", err)
	}

	if len(warnings) > 0 {
		for _, w := range warnings {
			fmt.Printf("Query warning: %v\n", w)
		}
	}

	if plugin.DebugQuery {
		fmt.Printf("Query result: %v\n", result)
	}

	var value float64
	switch result.Type() {
	case model.ValScalar:
		scalar := result.(*model.Scalar)
		value = float64(scalar.Value)

	case model.ValVector:
		vector := result.(model.Vector)
		if len(vector) == 0 {
			return sensu.CheckStateUnknown, fmt.Errorf("query returns empty vector: %v", result)
		}
		value = float64(vector[0].Value)

	default:
		return sensu.CheckStateUnknown, fmt.Errorf("unsupported type for qery result: %s, result: %v", result.Type(), result)
	}

	crit := plugin.CriticalThreshold.Check(value)
	warn := plugin.WarningThreshold.Check(value)
	state := sensu.CheckStateOK

	if crit {
		fmt.Printf("CRITICAL: %s=%f which is out of %s", plugin.PerfdataName, value, plugin.CriticalStr)
		state = sensu.CheckStateCritical
	} else if warn {
		fmt.Printf("WARNING: %s=%f which is out of %s", plugin.PerfdataName, value, plugin.WarningStr)
		state = sensu.CheckStateWarning
	} else {
		fmt.Printf("OK: %s=%f", plugin.PerfdataName, value)
	}

	if plugin.EmitPerfdata {
		fmt.Printf(" | %s=%f", plugin.PerfdataName, value)
	}

	fmt.Println()
	return state, nil
}
