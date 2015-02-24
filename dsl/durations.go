package dsl

import (
	"math"
	"time"
)

//Duration wraps []time.Duration and adds a number of convenience methods
type Durations []time.Duration

//Min returns the smallest duration in the list
func (d Durations) Min() time.Duration {
	min := time.Duration(math.MaxInt64)
	for _, duration := range d {
		if duration < min {
			min = duration
		}
	}

	return min
}

//Max returns the largest duration in the list
func (d Durations) Max() time.Duration {
	max := time.Duration(math.MinInt64)
	for _, duration := range d {
		if duration > max {
			max = duration
		}
	}

	return max
}

//CountInRange returns the number of durations between min and max
//It's used to construct histograms
func (d Durations) CountInRange(min time.Duration, max time.Duration) int {
	count := 0
	for _, duration := range d {
		if min <= duration && duration < max {
			count++
		}
	}
	return count
}
