package dsl

//GroupedTimelines represent an ordered collection of Grouped Timelines
type GroupedTimelines struct {
	Keys      []interface{}
	Timelines []Timelines
	lookup    map[interface{}]int
}

//NewGroupedTimelines initializes a new GroupedTimelines
func NewGroupedTimelines() *GroupedTimelines {
	return &GroupedTimelines{
		lookup: map[interface{}]int{},
	}
}

//Append adds an timeline to the given Key
func (g *GroupedTimelines) Append(key interface{}, timeline Timeline) {
	g.AppendTimelines(key, Timelines{timeline})
}

//Append timelines appends a slice of Timeline to the given Key
func (g *GroupedTimelines) AppendTimelines(key interface{}, timelines Timelines) {
	_, hasKey := g.lookup[key]
	if !hasKey {
		g.Keys = append(g.Keys, key)
		g.Timelines = append(g.Timelines, Timelines{})
		g.lookup[key] = len(g.Keys) - 1
	}
	g.Timelines[g.lookup[key]] = append(g.Timelines[g.lookup[key]], timelines...)
}

//Look up timelines for a given key
func (g *GroupedTimelines) Lookup(key interface{}) (Timelines, bool) {
	index, ok := g.lookup[key]
	if !ok {
		return nil, false
	}
	return g.Timelines[index], true
}

//Description returns the TimelineDescription associated with the Timelines in the group
func (g *GroupedTimelines) Description() TimelineDescription {
	return g.Timelines[0].Description()
}

//EachGroup is an iterator (think functional thoughts) that loops over all Keys and Timelines in order
//
//  groupedTimelines.EachGroup(func(key interface{}, timelines Timelines) error {
//      fmt.Printf("%s: %s\n", key, timelines)
//      return nil
//  })
//
//will print all timelines in the group.  Returning non-nil will cause the iterator to abort.
func (g *GroupedTimelines) EachGroup(f func(interface{}, Timelines) error) error {
	for i := 0; i < len(g.Keys); i++ {
		err := f(g.Keys[i], g.Timelines[i])
		if err != nil {
			return err
		}
	}
	return nil
}
