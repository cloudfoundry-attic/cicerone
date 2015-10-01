package dsl

import (
	"io"
	"time"
)

//Entries is a list of invidiual Entry(ies)
type Entries []Entry

//First returns the *first* Entry that satisfies the passed in Matcher.
//The second return value tells the caller if an entry was found or not
func (e Entries) First(matcher Matcher) (Entry, bool) {
	for _, entry := range e {
		if matcher.Match(entry) {
			return entry, true
		}
	}
	return Entry{}, false
}

//Filter returns the list of Entries that match the passed-in Matcher
func (e Entries) Filter(matcher Matcher) Entries {
	filtered := Entries{}
	for _, entry := range e {
		if matcher.Match(entry) {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

//ConstructTimeline takes a TimelineDescription and a Zeroth entry and returns a Timeline
//The Zeroth entry is used to compute the starting time the Timeline
func (e Entries) ConstructTimeline(description TimelineDescription, zeroEntry Entry) Timeline {
	timeline := Timeline{
		Description: description,
		ZeroEntry:   zeroEntry,
	}

	timeOffset := time.Duration(0)

	for i, point := range description {
		entry, _ := e.First(point.Matcher)
		if i > 0 {
			previousEntry, _ := e.First(description[i-1].Matcher)
			duration := entry.Timestamp.Sub(previousEntry.Timestamp)
			timeOffset -= time.Duration(float64(duration) * (1 - point.Squash))
			entry.Timestamp = entry.Timestamp.Add(timeOffset)
		}

		timeline.Entries = append(timeline.Entries, entry)
	}

	return timeline
}

//GroupBy groups all Entries by the passed in Getter it returns a GroupedEntries object
//The values returned by the Getter correpond to the Keys in the returned GroupedEntries object
func (e Entries) GroupBy(getter Getter) *GroupedEntries {
	groups := NewGroupedEntries()

	for _, entry := range e {
		key, ok := getter.Get(entry)
		if !ok {
			continue
		}
		groups.Append(key, entry)
	}

	return groups
}

//WriteLagerFormatTo emits lager-formatted entries to the passed in writer
func (e Entries) WriteLagerFormatTo(w io.Writer) error {
	for _, entry := range e {
		err := entry.WriteLagerFormatTo(w)
		if err != nil {
			return err
		}
	}
	return nil
}

func (e Entries) Len() int           { return len(e) }
func (e Entries) Swap(i, j int)      { e[i], e[j] = e[j], e[i] }
func (e Entries) Less(i, j int) bool { return e[i].Timestamp.Before(e[j].Timestamp) }
