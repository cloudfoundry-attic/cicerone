package dsl

import (
	"fmt"
	"io"
)

type Entries []Entry

func (e Entries) First(matcher Matcher) (Entry, bool) {
	for _, entry := range e {
		if matcher.Match(entry) {
			return entry, true
		}
	}
	return Entry{}, false
}

func (e Entries) Filter(matcher Matcher) Entries {
	filtered := Entries{}
	for _, entry := range e {
		if matcher.Match(entry) {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

//Dedupe entries based on passed in matcher: only the *first* entry matching the matcher survives
func (e Entries) Dedupe(matcher Matcher) Entries {
	deduped := Entries{}
	found := false
	for _, entry := range e {
		if matcher.Match(entry) {
			if found {
				continue
			}
			found = true
		}
		deduped = append(deduped, entry)
	}
	return deduped
}

func (e Entries) ConstructTimeline(description TimelineDescription, zeroEntry Entry) Timeline {
	timeline := Timeline{
		Description: description,
		ZeroEntry:   zeroEntry,
	}

	for _, point := range description {
		entry, _ := e.First(point.Matcher)
		timeline.Entries = append(timeline.Entries, entry)
	}

	return timeline
}

func (e Entries) PairAllWith(firstPairEntry Entry, annotationGetter Getter) EntryPairs {
	pairs := EntryPairs{}
	for _, entry := range e {
		annotation, _ := annotationGetter.Get(entry)
		pairs = append(pairs, EntryPair{
			FirstEntry:  firstPairEntry,
			SecondEntry: entry,
			Annotation:  annotation,
		})
	}
	return pairs
}

//Find the *first* entry that matches the first matcher, and the *last* entry that matches the last matcher and pair them up
func (e Entries) FindFirstPair(first Matcher, last Matcher) (EntryPair, bool) {
	pair := EntryPair{}
	found := false
	for _, entry := range e {
		if first.Match(entry) {
			pair.FirstEntry = entry
			found = true
			break
		}
	}
	if !found {
		return EntryPair{}, false
	}

	found = false
	for i := len(e) - 1; i >= 0; i-- {
		if last.Match(e[i]) {
			pair.SecondEntry = e[i]
			found = true
			break
		}
	}
	if !found {
		return EntryPair{}, false
	}

	return pair, true
}

//Find all pairs that match first then last.  Assumes entries aren't interleaved
func (e Entries) FindAllPairs(first Matcher, last Matcher) EntryPairs {
	pairs := EntryPairs{}
	pair := EntryPair{}
	for _, entry := range e {
		if first.Match(entry) {
			if pair.FirstEntry.IsZero() && pair.SecondEntry.IsZero() {
				pair.FirstEntry = entry
			} else {
				pair = EntryPair{
					FirstEntry: entry,
				}
			}
		} else if second.Match(entry) {
			if !pair.FirstEntry.IsZero() {
				pair.SecondEntry = second
				pairs = append(pairs, pair)
				pair = EntryPair{}
			}
		}
	}
	return pairs
}

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

func (e Entries) WriteLagerFormatTo(w io.Writer) error {
	for _, entry := range e {
		err := entry.WriteLagerFormatTo(w)
		if err != nil {
			return err
		}
	}
	return nil
}

type GroupedEntries struct {
	Keys    []interface{}
	Entries []Entries
	lookup  map[interface{}]int
}

func NewGroupedEntries() *GroupedEntries {
	return &GroupedEntries{
		lookup: map[interface{}]int{},
	}
}

func (g *GroupedEntries) Append(key interface{}, entry Entry) {
	g.AppendEntries(key, Entries{entry})
}

func (g *GroupedEntries) AppendEntries(key interface{}, entries Entries) {
	_, hasKey := g.lookup[key]
	if !hasKey {
		g.Keys = append(g.Keys, key)
		g.Entries = append(g.Entries, Entries{})
		g.lookup[key] = len(g.Keys) - 1
	}
	g.Entries[g.lookup[key]] = append(g.Entries[g.lookup[key]], entries...)
}

func (g *GroupedEntries) EachGroup(f func(interface{}, Entries) error) error {
	for i := 0; i < len(g.Keys); i++ {
		err := f(g.Keys[i], g.Entries[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *GroupedEntries) Filter(matcher Matcher) *GroupedEntries {
	filteredGroups := NewGroupedEntries()
	g.EachGroup(func(key interface{}, entries Entries) error {
		filtered := entries.Filter(matcher)
		if len(filtered) > 0 {
			filteredGroups.AppendEntries(key, filtered)
		}
		return nil
	})
	return filteredGroups
}

//Run entries.Dedupe for each entry
func (g *GroupedEntries) Dedupe(matcher Matcher) *GroupedEntries {
	dedupedGroups := NewGroupedEntries()
	g.EachGroup(func(key interface{}, entries Entries) error {
		deduped := entries.Dedupe(matcher)
		if len(deduped) > 0 {
			dedupedGroups.AppendEntries(key, deduped)
		}
		return nil
	})
	return dedupedGroups
}

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

func (g *GroupedEntries) FindFirstPairs(first Matcher, last Matcher) EntryPairs {
	pairs := EntryPairs{}
	g.EachGroup(func(key interface{}, entries Entries) error {
		pair, found := entries.FindFirstPair(first, last)
		if found {
			pair.Annotation = key
			pairs = append(pairs, pair)
		}
		return nil
	})
	return pairs
}

func (g *GroupedEntries) FindAllPairs(first Matcher, last Matcher) EntryPairs {
	allPairs := EntryPairs{}
	g.EachGroup(func(key interface{}, entries Entries) error {
		pairs := entries.FindAllPairs(first, last)
		for i := range pairs {
			pairs[i].Annotation = key
		}
		allPairs = append(allPairs, pairs...)
		return nil
	})
	return allPairs
}

func (g *GroupedEntries) WriteLagerFormatTo(w io.Writer) error {
	g.EachGroup(func(key interface{}, entries Entries) error {
		fmt.Fprintf(w, "%s\n", key)
		return entries.WriteLagerFormatTo(w)
	})
	return nil
}
