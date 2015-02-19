package dsl

import (
	"fmt"
	"strings"
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
	Name      string
	Min       time.Duration
	Max       time.Duration
	Mean      time.Duration
	N         int
	MinWinner EntryPair
	MaxWinner EntryPair
}

func (d DTStats) String() string {
	s := fmt.Sprintf("[%d] %s (%s) < %s < %s (%s)", d.N, d.Min, d.MinWinner.Annotation, d.Mean, d.Max, d.MaxWinner.Annotation)
	if d.Name != "" {
		s = fmt.Sprintf("%s\n\t%s", d.Name, s)
	}
	return s
}

type DTStatsSlice []DTStats

func (d DTStatsSlice) String() string {
	s := []string{}
	for _, dtStat := range d {
		s = append(s, dtStat.String())
	}
	return strings.Join(s, "\n")
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
