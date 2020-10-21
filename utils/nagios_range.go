package utils

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// NagiosRange represents range for value
//
// http://nagios-plugins.org/doc/guidelines.html#THRESHOLDFORMAT
type NagiosRange struct {
	Start    float64
	End      float64
	Inverted bool
}

func ParseRange(s string) (NagiosRange, error) {
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
			return NagiosRange{}, fmt.Errorf("Threshold parse error: %v", err)
		}

		end, err := parse(parts[0])
		if err != nil {
			return NagiosRange{}, fmt.Errorf("Threshold parse error: %v", err)
		}

		return NagiosRange{
			Start:    start,
			End:      end,
			Inverted: inverted,
		}, nil
	}

	f, err := parse(s)
	if err != nil {
		return NagiosRange{}, fmt.Errorf("Threshold parse error: %v", err)
	}

	return NagiosRange{
		Start:    0.0,
		End:      f,
		Inverted: inverted,
	}, nil
}

// Compare returns true if value should be alerted
func (r NagiosRange) Compare(v float64) bool {
	if r.Inverted {
		return v >= r.Start || v <= r.End
	}

	return v < r.Start || v > r.End
}
