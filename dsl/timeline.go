package dsl

import (
	"fmt"
	"strings"

	"github.com/pivotal-cf-experimental/veritas/say"
)

type TimelinePoint struct {
	Name    string
	Matcher Matcher
}

type TimelineDescription []TimelinePoint

type Timeline struct {
	Annotation  interface{}
	Description TimelineDescription
	ZeroEntry   Entry
	Entries     Entries
}

func (t Timeline) String() string {
	s := []string{fmt.Sprintf("%s:", say.Green("%s", t.Annotation))}

	runningTimestamp := t.ZeroEntry.Timestamp
	for i := 0; i < len(t.Description); i++ {
		if t.Entries[i].IsZero() {
			s = append(s, say.Red(t.Description[i].Name))
		} else {
			s = append(s, fmt.Sprintf("%s:%s", t.Description[i].Name, t.Entries[i].Timestamp.Sub(runningTimestamp)))
			runningTimestamp = t.Entries[i].Timestamp
		}
	}

	return strings.Join(s, " ")
}

func (t Timeline) EntryPair(index int) (EntryPair, bool) {
	if index >= len(t.Description) {
		return EntryPair{}, false
	}
	if t.Entries[index].IsZero() {
		return EntryPair{}, false
	}
	if index == 0 {
		return EntryPair{
			FirstEntry:  t.ZeroEntry,
			SecondEntry: t.Entries[0],
			Annotation:  t.Annotation,
		}, true
	}
	if t.Entries[index-1].IsZero() {
		return EntryPair{}, false
	}
	return EntryPair{
		FirstEntry:  t.Entries[index-1],
		SecondEntry: t.Entries[index],
		Annotation:  t.Annotation,
	}, true
}

type Timelines []Timeline

func (t Timelines) String() string {
	s := []string{}
	for _, timeline := range t {
		s = append(s, timeline.String())
	}
	return strings.Join(s, "\n")
}

func (t Timelines) Description() TimelineDescription {
	return t[0].Description
}

func (t Timelines) EntryPairs(index int) EntryPairs {
	pairs := EntryPairs{}
	for _, timeline := range t {
		pair, ok := timeline.EntryPair(index)
		if ok {
			pairs = append(pairs, pair)
		}
	}

	return pairs
}

func (t Timelines) DTStatsSlice() DTStatsSlice {
	dtStats := []DTStats{}
	for i, timelinePoint := range t.Description() {
		pairs := EntryPairs{}
		for _, timeline := range t {
			pair, ok := timeline.EntryPair(i)
			if ok {
				pairs = append(pairs, pair)
			}
		}
		stats := pairs.DTStats()
		stats.Name = timelinePoint.Name
		dtStats = append(dtStats, stats)
	}
	return dtStats
}
