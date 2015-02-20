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

func (e EntryPairs) Durations() Durations {
	durations := Durations{}
	for _, entry := range e {
		durations = append(durations, entry.DT())
	}
	return durations
}

type Durations []time.Duration

func (d Durations) Min() time.Duration {
	min := time.Hour * 1000000
	for _, duration := range d {
		if duration < min {
			min = duration
		}
	}

	return min
}

func (d Durations) Max() time.Duration {
	max := -time.Hour * 1000000
	for _, duration := range d {
		if duration > max {
			max = duration
		}
	}

	return max
}

func (d Durations) CountInRange(min time.Duration, max time.Duration) int {
	count := 0
	for _, duration := range d {
		if min <= duration && duration < max {
			count++
		}
	}
	return count
}
