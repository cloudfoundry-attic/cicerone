package dsl

import (
	"fmt"
	"io"
)

type Entries []Entry

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
func (e Entries) FindPair(first Matcher, last Matcher) (EntryPair, bool) {
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

func (g *GroupedEntries) FindPairs(first Matcher, last Matcher) EntryPairs {
	pairs := EntryPairs{}
	g.EachGroup(func(key interface{}, entries Entries) error {
		pair, found := entries.FindPair(first, last)
		if found {
			pair.Annotation = key
			pairs = append(pairs, pair)
		}
		return nil
	})
	return pairs
}

func (g *GroupedEntries) WriteLagerFormatTo(w io.Writer) error {
	g.EachGroup(func(key interface{}, entries Entries) error {
		fmt.Fprintf(w, "%s\n", key)
		return entries.WriteLagerFormatTo(w)
	})
	return nil
}
