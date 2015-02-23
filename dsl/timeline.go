package dsl

import (
	"fmt"
	"sort"
	"strings"
	"time"

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

func (t Timeline) BeginsAt() time.Time {
	for _, entry := range t.Entries {
		if entry.IsZero() {
			continue
		}
		return entry.Timestamp
	}

	return time.Unix(0, 0)
}

func (t Timeline) EndsAt() time.Time {
	for i := len(t.Entries) - 1; i > 0; i-- {
		entry := t.Entries[i]
		if entry.IsZero() {
			continue
		}
		return entry.Timestamp

	}

	return time.Unix(0, 0)
}

type Timelines []Timeline

func (t Timelines) String() string {
	s := []string{}
	for _, timeline := range t {
		s = append(s, timeline.String())
	}
	return strings.Join(s, "\n")
}

func (t Timelines) Len() int      { return len(t) }
func (t Timelines) Swap(i, j int) { t[i], t[j] = t[j], t[i] }

func (t Timelines) Description() TimelineDescription {
	return t[0].Description
}

type byVMForEntryAtIndex struct {
	Timelines
	index int
}

func (s byVMForEntryAtIndex) Less(i, j int) bool {
	a := s.Timelines[i].Entries[s.index].VM()
	b := s.Timelines[j].Entries[s.index].VM()
	if a == b {
		return !s.Timelines[i].Entries[s.index].Timestamp.After(s.Timelines[j].Entries[s.index].Timestamp)
	}
	return a < b
}

func (t Timelines) SortByVMForEntryAtIndex(index int) {
	sort.Sort(byVMForEntryAtIndex{t, index})
}

type byEntryAtIndex struct {
	Timelines
	index int
}

func (s byEntryAtIndex) Less(i, j int) bool {
	return !s.Timelines[i].Entries[s.index].Timestamp.After(s.Timelines[j].Entries[s.index].Timestamp)
}

func (t Timelines) SortByEntryAtIndex(index int) {
	sort.Sort(byEntryAtIndex{t, index})
}

type byEndTime struct {
	Timelines
}

func (s byEndTime) Less(i, j int) bool {
	return !s.Timelines[i].EndsAt().After(s.Timelines[j].EndsAt())
}

func (t Timelines) SortByEndTime() {
	sort.Sort(byEndTime{t})
}

type byStartTime struct {
	Timelines
}

func (s byStartTime) Less(i, j int) bool {
	return !s.Timelines[i].BeginsAt().After(s.Timelines[j].BeginsAt())
}

func (t Timelines) SortByStartTime() {
	sort.Sort(byStartTime{t})
}

func (t Timelines) EntryPairs(index int) EntryPairs {
	pairs := EntryPairs{}
	for _, timeline := range t {
		pair, ok := timeline.EntryPair(index)
		if ok && pair.FirstEntry.Timestamp.Before(pair.SecondEntry.Timestamp) {
			pairs = append(pairs, pair)
		}
	}

	return pairs
}

func (t Timelines) DTStatsSlice() DTStatsSlice {
	dtStats := []DTStats{}
	for i, timelinePoint := range t.Description() {
		pairs := t.EntryPairs(i)
		stats := pairs.DTStats()
		stats.Name = timelinePoint.Name
		dtStats = append(dtStats, stats)
	}
	return dtStats
}

func (t Timelines) StartsAfter() time.Duration {
	min := time.Hour * 100000
	for _, timeline := range t {
		dt := timeline.BeginsAt().Sub(timeline.ZeroEntry.Timestamp)
		if dt < min {
			min = dt
		}
	}

	return min
}

func (t Timelines) EndsAfter() time.Duration {
	max := -time.Hour * 100000
	for _, timeline := range t {
		dt := timeline.EndsAt().Sub(timeline.ZeroEntry.Timestamp)
		if dt > max {
			max = dt
		}
	}

	return max
}
