package utils

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseRange(t *testing.T) {
	inf := math.Inf(1)
	ninf := math.Inf(-1)

	testCases := []struct {
		Input  string
		Output NagiosRange
		Err    error
	}{
		{"10", NagiosRange{0.0, 10.0, false}, nil},
		{"10:", NagiosRange{10.0, inf, false}, nil},
		{"~:10", NagiosRange{ninf, 10.0, false}, nil},
		{"10:20", NagiosRange{10.0, 20.0, false}, nil},
		{"@10:20", NagiosRange{10.0, 20.0, true}, nil},
		{"abc", NagiosRange{}, nil},
		{"abc:10", NagiosRange{}, nil},
		{"10:abc", NagiosRange{}, nil},
	}

	for _, tc := range testCases {
		t.Run(tc.Input, func(t *testing.T) {
			assert := assert.New(t)

			result, err := ParseRange(tc.Input)
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
