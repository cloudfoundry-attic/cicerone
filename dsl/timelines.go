package dsl

import (
	"math"
	"sort"
	"strings"
	"time"
)

//Timelines is a slice of Timeline
//It is assumed (though not enforced) that all Timeline entries share identical TimelineDescriptions
type Timelines []Timeline

//String() joins the strings of the constitutent Timeline entries
func (t Timelines) String() string {
	s := []string{}
	for _, timeline := range t {
		s = append(s, timeline.String())
	}
	return strings.Join(s, "\n")
}

//CompleteTimelines returns the subset of Timelines that are complete.
func (t Timelines) CompleteTimelines() Timelines {
	subset := Timelines{}
	for _, timeline := range t {
		if timeline.IsComplete() {
			subset = append(subset, timeline)
		}
	}
	return subset
}

//Len returns the length of the Timelines slice
func (t Timelines) Len() int { return len(t) }

//Swap swaps two timelines in place
func (t Timelines) Swap(i, j int) { t[i], t[j] = t[j], t[i] }

//Description returns the TimelineDescription associated with the Timelines
func (t Timelines) Description() TimelineDescription {
	return t[0].Description
}

//SortByVMForEntryAtindex sorts the Timelines in-place by VM
//
//Since a Timeline can be comprised of events that span multiple VMs one must specify the entry (by index in the TimelineDescription)
//that should be used to fetch the VM.
func (t Timelines) SortByVMForEntryAtIndex(index int) {
	sort.Sort(byVMForEntryAtIndex{t, index})
}

//SortByEntryAtIndex sorts the Timelines in-place by the timestamp of the entry at the specifeid index in the TimelineDescription
func (t Timelines) SortByEntryAtIndex(index int) {
	sort.Sort(byEntryAtIndex{t, index})
}

//SortByEndTime sorts the Timelines in-place by the timestamp returned by timeline.EndsAt()
func (t Timelines) SortByEndTime() {
	sort.Sort(byEndTime{t})
}

//SortByStartTime sorts the Timelines in-place by the timestamp returned by timeline.BeginsAt()
func (t Timelines) SortByStartTime() {
	sort.Sort(byStartTime{t})
}

//EntryPairs fetches the EntryPair at the passed-in index for each Timeline then returns the corresponding list of EntryPairs
//
//See the documentation for timeline.EntryPair for more details
func (t Timelines) EntryPairs(index int) EntryPairs {
	pairs := EntryPairs{}
	for _, timeline := range t {
		pair, ok := timeline.EntryPair(index)
		if ok && !pair.FirstEntry.Timestamp.After(pair.SecondEntry.Timestamp) {
			pairs = append(pairs, pair)
		}
	}

	return pairs
}

//DTStatsSlice returns a collection of DTStats -- one for each TimelinePoint in the TimelineDescription
//
//This allows one to efficiently identify (for example) which TimelinePoint contributes the most to the total elapsed time.
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

//StartsAfter finds the smallest timeline.BeginsAt().Sub(timeline.ZeroEntry.Timestamp) in the entire set of Timelines in the sample
func (t Timelines) StartsAfter() time.Duration {
	min := time.Duration(math.MaxInt64)
	for _, timeline := range t {
		dt := timeline.BeginsAt().Sub(timeline.ZeroEntry.Timestamp)
		if dt < min {
			min = dt
		}
	}

	return min
}

//StartsAfter finds the largest timeline.EndsAt().Sub(timeline.ZeroEntry.Timestamp) in the entire set of Timelines in the sample
func (t Timelines) EndsAfter() time.Duration {
	max := time.Duration(math.MinInt64)
	for _, timeline := range t {
		dt := timeline.EndsAt().Sub(timeline.ZeroEntry.Timestamp)
		if dt > max {
			max = dt
		}
	}

	return max
}

// Sorters (private)

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

type byEntryAtIndex struct {
	Timelines
	index int
}

func (s byEntryAtIndex) Less(i, j int) bool {
	return !s.Timelines[i].Entries[s.index].Timestamp.After(s.Timelines[j].Entries[s.index].Timestamp)
}

type byEndTime struct {
	Timelines
}

func (s byEndTime) Less(i, j int) bool {
	return !s.Timelines[i].EndsAt().After(s.Timelines[j].EndsAt())
}

type byStartTime struct {
	Timelines
}

func (s byStartTime) Less(i, j int) bool {
	return !s.Timelines[i].BeginsAt().After(s.Timelines[j].BeginsAt())
}
