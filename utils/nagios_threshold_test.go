package utils

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseThreshold(t *testing.T) {
	inf := math.Inf(1)
	ninf := math.Inf(-1)
	abcerr := fmt.Errorf(`Threshold parse error: strconv.ParseFloat: parsing "abc": invalid syntax`)

	testCases := []struct {
		Input  string
		Output NagiosThreshold
		Err    error
	}{
		{"10", NagiosThreshold{0.0, 10.0, false}, nil},
		{"10:", NagiosThreshold{10.0, inf, false}, nil},
		{"~:10", NagiosThreshold{ninf, 10.0, false}, nil},
		{"10:20", NagiosThreshold{10.0, 20.0, false}, nil},
		{"@10:20", NagiosThreshold{10.0, 20.0, true}, nil},
		{"abc", NagiosThreshold{}, abcerr},
		{"abc:10", NagiosThreshold{}, abcerr},
		{"10:abc", NagiosThreshold{}, abcerr},
	}

	for _, tc := range testCases {
		t.Run(tc.Input, func(t *testing.T) {
			assert := assert.New(t)

			result, err := ParseThreshold(tc.Input)
			if tc.Err != nil {
				assert.EqualError(err, tc.Err.Error(), "parse error")
				return
			}

			assert.NoError(err)
			assert.Equal(tc.Output.Start, result.Start, "start")
			assert.Equal(tc.Output.End, result.End, "end")
			assert.Equal(tc.Output.Inverted, result.Inverted, "inverted")
		})
	}
}

func TestNagiosThresholdCheck(t *testing.T) {
	type ValueResult struct {
		Value  float64
		Result bool
	}

	testCases := []struct {
		Input  string
		Values []ValueResult
	}{
		{"10", []ValueResult{{-1.0, true}, {0.0, false}, {10.0, false}, {20.0, true}}},
		{"10:", []ValueResult{{5.0, true}, {10.0, false}, {20.0, false}}},
		{"~:10", []ValueResult{{-1.0, false}, {5.0, false}, {10.0, false}, {20.0, true}}},
		{"10:20", []ValueResult{{5.0, true}, {10.0, false}, {15.0, false}, {20.0, false}, {30.0, true}}},
		{"@10:20", []ValueResult{{5.0, false}, {10.0, true}, {15.0, true}, {20.0, true}, {30.0, false}}},
	}

	for _, tc := range testCases {
		for _, tv := range tc.Values {
			t.Run(fmt.Sprintf("%s_%0.1f", tc.Input, tv.Value), func(t *testing.T) {
				assert := assert.New(t)

				threshold, err := ParseThreshold(tc.Input)
				assert.NoError(err)

				result := threshold.Check(tv.Value)
				assert.Equal(tv.Result, result)
			})
		}
	}
}
