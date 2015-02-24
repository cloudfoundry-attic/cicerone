package dsl

import (
	"fmt"
	"io"
)

//GroupedEntries represent an ordered collection of Grouped Entries.
//
//groupedEntries.Keys is a slice of Keys (the value returned by the Getter passed to entries.GroupBy)
//groupedEntries.Entries is a slice of Entries for each Key
//
//GroupedEntries uses parallel arrays to retain the ordering in which Keys are detected - the first Key in the GroupedEntries
//is the first Key that was detected by entried.GroupBy
type GroupedEntries struct {
	Keys    []interface{}
	Entries []Entries
	lookup  map[interface{}]int
}

func newGroupedEntries() *GroupedEntries {
	return &GroupedEntries{
		lookup: map[interface{}]int{},
	}
}

//Append adds an entry to the given Key
func (g *GroupedEntries) Append(key interface{}, entry Entry) {
	g.AppendEntries(key, Entries{entry})
}

//Append entries appends a slice of Entries to the given Key
func (g *GroupedEntries) AppendEntries(key interface{}, entries Entries) {
	_, hasKey := g.lookup[key]
	if !hasKey {
		g.Keys = append(g.Keys, key)
		g.Entries = append(g.Entries, Entries{})
		g.lookup[key] = len(g.Keys) - 1
	}
	g.Entries[g.lookup[key]] = append(g.Entries[g.lookup[key]], entries...)
}

//EachGroup is an iterator (think functional thoughts) that loops over all Keys and Entries in order
//
//	groupedEntries.EachGroup(func(key interface{}, entries Entries) error {
//  	fmt.Printf("%s: %s\n", key, entries)
//		return nil
//	})
//
//will print all entries in the group.  Returning non-nil will cause the iterator to abort.
func (g *GroupedEntries) EachGroup(f func(interface{}, Entries) error) error {
	for i := 0; i < len(g.Keys); i++ {
		err := f(g.Keys[i], g.Entries[i])
		if err != nil {
			return err
		}
	}
	return nil
}

//Filter applies `entries.Filter` to each Entries element in the group.
//Entries that end up empty are removed from the group.
//
//As a consequence:
//
//	entries.Filter(matcher).GroupBy(getter)
//
//and
//
//	entries.GroupBy(getter).Filter(matcher)
//
//are identical.
func (g *GroupedEntries) Filter(matcher Matcher) *GroupedEntries {
	filteredGroups := newGroupedEntries()
	g.EachGroup(func(key interface{}, entries Entries) error {
		filtered := entries.Filter(matcher)
		if len(filtered) > 0 {
			filteredGroups.AppendEntries(key, filtered)
		}
		return nil
	})
	return filteredGroups
}

//ConstructTimelines creates a slice of Timelines by calling entries.ConstructTimeline on each Entries element in the group.
//The Key associated with the Entries element becomes the Annotation associated with the Timeline.
//
//Note that Timelines aren't Key=>Timeline mappings.  Instead GroupedEntries returns a *flat list* of Timelines with the Key parameter associated with the individual Timeline.
func (g *GroupedEntries) ConstructTimelines(description TimelineDescription, zeroEntry Entry) Timelines {
	timelines := Timelines{}

	g.EachGroup(func(key interface{}, entries Entries) error {
		timeline := entries.ConstructTimeline(description, zeroEntry)
		timeline.Annotation = key
		timelines = append(timelines, timeline)
		return nil
	})

	return timelines
}

//WriteLagerFormat emits lager formatted output for all Entries in the group.
func (g *GroupedEntries) WriteLagerFormatTo(w io.Writer) error {
	g.EachGroup(func(key interface{}, entries Entries) error {
		fmt.Fprintf(w, "%s\n", key)
		return entries.WriteLagerFormatTo(w)
	})
	return nil
}
