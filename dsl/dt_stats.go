package dsl

import (
	"fmt"
	"strings"
	"time"
)

//DTStats bndles up statistics extracted from a collection of EntryPairs
type DTStats struct {
	Name      string
	Min       time.Duration
	Max       time.Duration
	Mean      time.Duration
	N         int
	MinWinner EntryPair
	MaxWinner EntryPair
}

//Printing out a DTStats is useful - it will emit the Annotation associated with the most extreme EntryPair outliers in the sample
func (d DTStats) String() string {
	s := fmt.Sprintf("[%d] %s (%s) < %s < %s (%s)", d.N, d.Min, d.MinWinner.Annotation, d.Mean, d.Max, d.MaxWinner.Annotation)
	if d.Name != "" {
		s = fmt.Sprintf("%s\n\t%s", d.Name, s)
	}
	return s
}

//DTStatsSLice is a collection of DTStats
type DTStatsSlice []DTStats

//String() joins the Strings() of the underlying DTStats
func (d DTStatsSlice) String() string {
	s := []string{}
	for _, dtStat := range d {
		s = append(s, dtStat.String())
	}
	return strings.Join(s, "\n")
}
