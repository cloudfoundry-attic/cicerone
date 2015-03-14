package dsl

import (
	"strings"
	"time"
)

//EntryPairs is a slice of EntryPair
type EntryPairs []EntryPair

func (e EntryPairs) String() string {
	s := []string{}
	for _, pair := range e {
		s = append(s, pair.String())
	}
	return strings.Join(s, "\n")
}

//DTStats returns a DTStats rollup summarizing the distribution of time intervals in the slice of EntryPairs
func (e EntryPairs) DTStats() DTStats {
	var minWinner, maxWinner EntryPair
	min := time.Hour * 1000000
	max := -time.Hour * 1000000
	mean := time.Duration(0)
	for _, pair := range e {
		dt := pair.DT()
		mean += dt
		if dt < min {
			min = dt
			minWinner = pair
		}
		if dt > max {
			max = dt
			maxWinner = pair
		}
	}
	mean = mean / time.Duration(len(e))

	return DTStats{
		Min:       min,
		Max:       max,
		Mean:      mean,
		N:         len(e),
		MinWinner: minWinner,
		MaxWinner: maxWinner,
	}
}

//Durations returns a Durations slice of time.Duration composed - it's a collection of the time intervals in the slice of EntryPairs
func (e EntryPairs) Durations() Durations {
	durations := Durations{}
	for _, entry := range e {
		durations = append(durations, entry.DT())
	}
	return durations
}

//FilterByDurationGreaterThan returns a copy of the EntryPairs containing entries that exceed the passed in duration
func (e EntryPairs) FilterByDurationGreaterThan(cutoff time.Duration) EntryPairs {
	filteredPairs := EntryPairs{}
	for _, pair := range e {
		if pair.DT() > cutoff {
			filteredPairs = append(filteredPairs, pair)
		}
	}

	return filteredPairs
}

//FilterByDurationGreaterThan returns a copy of the EntryPairs containing entries that exceed the passed in duration
func (e EntryPairs) FilterByDurationLessThan(cutoff time.Duration) EntryPairs {
	filteredPairs := EntryPairs{}
	for _, pair := range e {
		if pair.DT() < cutoff {
			filteredPairs = append(filteredPairs, pair)
		}
	}

	return filteredPairs
}
