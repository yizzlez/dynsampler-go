package dynsampler

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateEMA(t *testing.T) {
	e := &EMASampleRate{
		movingAverage: make(map[string]float64),
		Weight:        0.2,
		AgeOutValue:   0.2,
	}

	tests := []struct {
		keyAValue    float64
		keyAExpected float64
		keyBValue    float64
		keyBExpected float64
		keyCValue    float64
		keyCExpected float64
	}{
		{463, 93, 235, 47, 0, 0},
		{176, 109, 458, 129, 0, 0},
		{345, 156, 470, 197, 0, 0},
		{339, 193, 317, 221, 0, 0},
		{197, 194, 165, 210, 0, 0},
		{387, 232, 95, 187, 6960, 1392},
	}

	for _, tt := range tests {
		counts := make(map[string]float64)
		counts["a"] = tt.keyAValue
		counts["b"] = tt.keyBValue
		counts["c"] = tt.keyCValue
		e.updateEMA(counts)
		assert.Equal(t, tt.keyAExpected, math.Round(e.movingAverage["a"]))
		assert.Equal(t, tt.keyBExpected, math.Round(e.movingAverage["b"]))
		assert.Equal(t, tt.keyCExpected, math.Round(e.movingAverage["c"]))
	}
}

func TestEMASampleGetSampleRateStartup(t *testing.T) {
	e := &EMASampleRate{
		GoalSampleRate: 10,
		currentCounts:  map[string]float64{},
	}
	rate := e.GetSampleRate("key")
	assert.Equal(t, rate, 10)
	assert.Equal(t, e.currentCounts["key"], float64(1))
}

func TestEMASampleUpdateMaps(t *testing.T) {
	e := &EMASampleRate{
		GoalSampleRate: 20,
		Weight:         0.2,
		AgeOutValue:    0.2,
	}
	tsts := []struct {
		inputSampleCount         map[string]float64
		expectedSavedSampleRates map[string]int
	}{
		{
			map[string]float64{
				"one":   1,
				"two":   1,
				"three": 2,
				"four":  5,
				"five":  8,
				"six":   15,
				"seven": 45,
				"eight": 612,
				"nine":  2000,
				"ten":   10000,
			},
			map[string]int{
				"one":   1,
				"two":   1,
				"three": 1,
				"four":  1,
				"five":  1,
				"six":   1,
				"seven": 1,
				"eight": 6,
				"nine":  14,
				"ten":   47,
			},
		},
		{
			map[string]float64{
				"one":   1,
				"two":   1,
				"three": 2,
				"four":  5,
				"five":  8,
				"six":   15,
				"seven": 45,
				"eight": 50,
				"nine":  60,
			},
			map[string]int{
				"one":   1,
				"two":   1,
				"three": 2,
				"four":  5,
				"five":  8,
				"six":   11,
				"seven": 24,
				"eight": 26,
				"nine":  30,
			},
		},
		{
			map[string]float64{
				"one":   1,
				"two":   1,
				"three": 2,
				"four":  5,
				"five":  7,
			},
			map[string]int{
				"one":   1,
				"two":   1,
				"three": 2,
				"four":  5,
				"five":  7,
			},
		},
		{
			map[string]float64{
				"one":   1000,
				"two":   1000,
				"three": 2000,
				"four":  5000,
				"five":  7000,
			},
			map[string]int{
				"one":   7,
				"two":   7,
				"three": 13,
				"four":  29,
				"five":  39,
			},
		},
		{
			map[string]float64{
				"one":   6000,
				"two":   6000,
				"three": 6000,
				"four":  6000,
				"five":  6000,
			},
			map[string]int{
				"one":   20,
				"two":   20,
				"three": 20,
				"four":  20,
				"five":  20,
			},
		},
		{
			map[string]float64{
				"one": 12000,
			},
			map[string]int{
				"one": 20,
			},
		},
		{
			map[string]float64{},
			map[string]int{},
		},
	}
	for i, tst := range tsts {
		e.movingAverage = make(map[string]float64)
		e.savedSampleRates = make(map[string]int)

		// Test data is based on `TestAvgSampleUpdateMaps` for AvgSampleRate.
		// To get the same sample rates though, we must reach averages that match
		// the inputs - for the EMA, the way to do this is to just apply the same
		// input values over and over and converge on that average
		for i := 0; i <= 100; i++ {
			input := make(map[string]float64)
			for k, v := range tst.inputSampleCount {
				input[k] = v
			}
			e.currentCounts = input
			e.updateMaps()
		}
		assert.Equal(t, 0, len(e.currentCounts))
		assert.Equal(t, tst.expectedSavedSampleRates, e.savedSampleRates, fmt.Sprintf("test %d failed", i))
	}
}

func TestEMAAgesOutSmallValues(t *testing.T) {
	e := &EMASampleRate{
		GoalSampleRate: 20,
		Weight:         0.2,
		AgeOutValue:    0.2,
	}
	e.movingAverage = make(map[string]float64)
	for i := 0; i < 100; i++ {
		e.currentCounts = map[string]float64{"foo": 500.0}
		e.updateMaps()
	}
	assert.Equal(t, 1, len(e.movingAverage))
	assert.Equal(t, float64(500), math.Round(e.movingAverage["foo"]))
	for i := 0; i < 100; i++ {
		// "observe" no occurrences of foo for many iterations
		e.currentCounts = map[string]float64{"asdf": 1}
		e.updateMaps()
	}
	_, found := e.movingAverage["foo"]
	assert.Equal(t, false, found)
	_, found = e.movingAverage["asdf"]
	assert.Equal(t, true, found)
}
