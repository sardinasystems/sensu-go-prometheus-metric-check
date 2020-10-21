package utils

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// NagiosThreshold represents range for value
//
// http://nagios-plugins.org/doc/guidelines.html#THRESHOLDFORMAT
type NagiosThreshold struct {
	Start    float64
	End      float64
	Inverted bool
}

// ParseThreshold parses Nagios Threshold format
func ParseThreshold(s string) (NagiosThreshold, error) {
	inverted := false
	if strings.HasPrefix(s, "@") {
		inverted = true
		s = s[1:]
	}

	parse := func(ss string) (float64, error) {
		switch ss {
		case "~":
			return math.Inf(-1), nil
		case "":
			return math.Inf(1), nil
		default:
			return strconv.ParseFloat(ss, 64)
		}
	}

	if strings.Contains(s, ":") {
		parts := strings.SplitN(s, ":", 2)

		start, err := parse(parts[0])
		if err != nil {
			return NagiosThreshold{}, fmt.Errorf("Threshold parse error: %v", err)
		}

		end, err := parse(parts[1])
		if err != nil {
			return NagiosThreshold{}, fmt.Errorf("Threshold parse error: %v", err)
		}

		return NagiosThreshold{
			Start:    start,
			End:      end,
			Inverted: inverted,
		}, nil
	}

	f, err := parse(s)
	if err != nil {
		return NagiosThreshold{}, fmt.Errorf("Threshold parse error: %v", err)
	}

	return NagiosThreshold{
		Start:    0.0,
		End:      f,
		Inverted: inverted,
	}, nil
}

// Check returns true if value should be alerted
func (r NagiosThreshold) Check(v float64) bool {
	if r.Inverted {
		return v >= r.Start && v <= r.End
	}

	return v < r.Start || v > r.End
}
