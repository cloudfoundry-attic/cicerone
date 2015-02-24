package dsl

import (
	"fmt"
	"strings"
	"time"

	"github.com/onsi/say"
)

//Timeline encapsulates a slice of Entries and a corresponding TimelineDescription
//A Timeline can have an optional - arbitrary - Annotation
//A Timeline also has a ZeroEntry.  This is used as a zero point to compute the time at which the first entry occurs.
//The Timeline is effectively comprised of two parallel slices: the TimelineDescription and the slice of Entries.
type Timeline struct {
	Annotation  interface{}
	Description TimelineDescription
	ZeroEntry   Entry
	Entries     Entries
}

//String() produces a textual representation of the timeline.
//The TimelinePoint name and elapsed time are printed for each TimelinePoint in the Timeline's TimelineDescription.
//If a TimelinePoint is missing a corresponding Entry the TimelinePoint's name is rendered in red.
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

//EntryPair returns the EntryPair at the passed-in index of the Timeline.
//
//EntryPair is strict: only timeline points where both the preceding entry and requested entry are non-zero are returned.
//
//As a concrete example, consider a Timeline constructed with a TimelineDescription of:
//
//	description := {
//		{"Object-Created", CreateMatcher},
//		{"Object-Run", RunMatcher},
//		{"Object-Destroyed", DestroyMatcher}
//	}
//
//Then
//
//	timeline.EntryPair(1)
//
//Will return an EntryPair where the FirstEntry corresponds to the Object-Created Entry and the SecondEntry corresponds to the Object-Run event.
//
//The annotation associated with the resulting EntryPair is set to the Annotation associated with this Timeline.
//
//Finally, note that
//
//	timeline.EntryPair(0)
//
//returns the ZeroEntry as the FirstEntry in the pair.
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

//IsComplete returns true if all events in the timeline are present
func (t Timeline) IsComplete() bool {
	for i := range t.Description {
		if t.Entries[i].IsZero() {
			return false
		}
	}

	return true
}

//BeginsAt returns the timestamp at which the first non-zero entry in the Timeline occurs
//Note: BeginsAt does not include the ZeroEntry.
func (t Timeline) BeginsAt() time.Time {
	for _, entry := range t.Entries {
		if entry.IsZero() {
			continue
		}
		return entry.Timestamp
	}

	return time.Unix(0, 0)
}

//EndsAt returns the timestamp at which the last non-zero entry in the Timeline occurs
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
