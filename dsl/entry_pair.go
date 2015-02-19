package dsl

import (
	"fmt"
	"time"
)

type EntryPair struct {
	FirstEntry  Entry
	SecondEntry Entry
	Annotation  interface{}
}

func (e EntryPair) DT() time.Duration {
	return e.SecondEntry.Timestamp.Sub(e.FirstEntry.Timestamp)
}

type DTStats struct {
	Min       time.Duration
	Max       time.Duration
	Mean      time.Duration
	N         int
	MinWinner EntryPair
	MaxWinner EntryPair
}

func (d DTStats) String() string {
	return fmt.Sprintf("[%d]: %s (%s) < %s < %s (%s)", d.N, d.Min, d.MinWinner.Annotation, d.Mean, d.Max, d.MaxWinner.Annotation)
}

type EntryPairs []EntryPair

func (e EntryPairs) DTStats() DTStats {
	var minWinner, maxWinner EntryPair
	min := time.Hour * 24
	max := time.Nanosecond
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
