package dsl

//A TimelinePoint describes a point in a TimelineDescription
//
//The Name is any string to associate with the TimelinePoint
//The Matcher is used to identify Entries that should be associated with the TimelinePoint
type TimelinePoint struct {
	Name    string
	Matcher Matcher
}

//A TimelineDescription is an ordered list of TimelinePoints.
type TimelineDescription []TimelinePoint
