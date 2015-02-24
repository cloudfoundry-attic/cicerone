package dsl

import "time"

//EntryPair represents a span in time between two Entries
//It includes an arbitrary annotation
type EntryPair struct {
	FirstEntry  Entry
	SecondEntry Entry
	Annotation  interface{}
}

//DT returns the time.Duration between the two events in the EntryPair
func (e EntryPair) DT() time.Duration {
	return e.SecondEntry.Timestamp.Sub(e.FirstEntry.Timestamp)
}
