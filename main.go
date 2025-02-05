package main

import (
	"context"
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/sardinasystems/sensu-go-check-common/nagios"
	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-plugin-sdk/sensu"
)

// Config represents the check plugin config.
type Config struct {
	sensu.PluginConfig
	PrometheusURL     string
	Query             string
	CriticalThreshold nagios.Threshold
	WarningThreshold  nagios.Threshold
	EmitPerfdata      bool
	Name              string
	DebugQuery        bool
	NanIsOk           bool
}

var (
	plugin = Config{
		PluginConfig: sensu.PluginConfig{
			Name:     "sensu-go-prometheus-metric-check",
			Short:    "plugin to query prometheus for alerting",
			Keyspace: "sensu.io/plugins/sensu-go-prometheus-metric-check/config",
		},
	}

	options = []sensu.ConfigOption{
		&sensu.PluginConfigOption[string]{
			Path:      "host",
			Env:       "PROMETHEUS_HOST",
			Argument:  "host",
			Shorthand: "H",
			Default:   "http://127.0.0.1:9090",
			Usage:     "Host URL to access Prometheus",
			Value:     &plugin.PrometheusURL,
		},
		&sensu.PluginConfigOption[string]{
			Path:      "query",
			Env:       "PROMETHEUS_QUERY",
			Argument:  "query",
			Shorthand: "q",
			Default:   "",
			Usage:     "PromQL query",
			Value:     &plugin.Query,
		},
		&nagios.ThresholdConfigOption{
			Option: sensu.PluginConfigOption[string]{
				Path:      "warning",
				Env:       "PROMETHEUS_WARNING",
				Argument:  "warning",
				Shorthand: "w",
				Default:   "",
				Usage:     "Warning level",
			},
			Value: &plugin.WarningThreshold,
		},
		&nagios.ThresholdConfigOption{
			Option: sensu.PluginConfigOption[string]{
				Path:      "critical",
				Env:       "PROMETHEUS_CRITICAL",
				Argument:  "critical",
				Shorthand: "c",
				Default:   "",
				Usage:     "Critical level",
			},
			Value: &plugin.CriticalThreshold,
		},
		&sensu.PluginConfigOption[bool]{
			Path:      "emit_perfdata",
			Env:       "PROMETHEUS_EMIT_PERFDATA",
			Argument:  "emit-perfdata",
			Shorthand: "p",
			Default:   false,
			Usage:     "Add perfdata to check output",
			Value:     &plugin.EmitPerfdata,
		},
		&sensu.PluginConfigOption[string]{
			Path:      "name",
			Env:       "PROMETHEUS_NAME",
			Argument:  "name",
			Shorthand: "n",
			Default:   "",
			Usage:     "Name",
			Value:     &plugin.Name,
		},
		&sensu.PluginConfigOption[bool]{
			Path:      "debug_query",
			Env:       "PROMETHEUS_DEBUG_QUERY",
			Argument:  "debug-query",
			Shorthand: "i",
			Default:   false,
			Usage:     "Enable debug output for query",
			Value:     &plugin.DebugQuery,
		},
		&sensu.PluginConfigOption[bool]{
			Path:      "nan_is_ok",
			Env:       "PROMETHEUS_NAN_IS_OK",
			Argument:  "nan-is-ok",
			Shorthand: "O",
			Default:   false,
			Usage:     "NaN result is ok",
			Value:     &plugin.NanIsOk,
		},
	}
)

func main() {
	useStdin := false
	fi, err := os.Stdin.Stat()
	if err != nil {
		fmt.Printf("Error check stdin: %v\n", err)
	}
	// Check the Mode bitmask for Named Pipe to indicate stdin is connected
	if fi.Mode()&os.ModeNamedPipe != 0 {
		useStdin = true
	}

	check := sensu.NewGoCheck(&plugin.PluginConfig, options, checkArgs, executeCheck, useStdin)
	check.Execute()
}

func checkArgs(event *corev2.Event) (int, error) {

	if plugin.Name == "" {
		qn := fmt.Sprintf("query_%s", plugin.Query)
		for _, rs := range []string{"+", "-", "*", "/", "[", "]", "{", "}", "(", ")", "=", ".", ":", ";", "\"", " ", "\t"} {
			qn = strings.ReplaceAll(qn, rs, "_")
		}

		plugin.Name = qn
	}

	return sensu.CheckStateOK, nil
}

func executeCheck(event *corev2.Event) (int, error) {

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

	if !math.IsNaN(value) {
		if crit {
			fmt.Printf("CRITICAL: %s is %f which is out of %s", plugin.Name, value, plugin.CriticalThreshold.String())
			state = sensu.CheckStateCritical
		} else if warn {
			fmt.Printf("WARNING: %s is %f which is out of %s", plugin.Name, value, plugin.WarningThreshold.String())
			state = sensu.CheckStateWarning
		} else {
			fmt.Printf("OK: %s is %f", plugin.Name, value)
		}
	} else {
		if plugin.NanIsOk {
			fmt.Printf("OK: %s is %f", plugin.Name, value)
		} else {
			fmt.Printf("UNKNOWN: %s is %f", plugin.Name, value)
			state = sensu.CheckStateUnknown
		}
	}

	if plugin.EmitPerfdata {
		fmt.Printf(" | %s=%f", plugin.Name, value)
	}

	fmt.Println()
	return state, nil
}
