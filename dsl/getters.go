package dsl

import (
	"encoding/json"
	"strings"
)

//Getter objects can return arbitrary data from a passed-in-Entry
type Getter interface {
	Get(entry Entry) (interface{}, bool)
}

//GetterFunc makes it easy to create Getters from bare functions
type GetterFunc func(Entry) (interface{}, bool)

//Get satisfies the Getter interface
func (g GetterFunc) Get(entry Entry) (interface{}, bool) {
	return g(entry)
}

//GetVM returns the VM associated with an entry
var GetVM = GetterFunc(func(entry Entry) (interface{}, bool) {
	return entry.VM(), true
})

//GetJob returns the Job associated with an entry
var GetJob = GetterFunc(func(entry Entry) (interface{}, bool) {
	return entry.Job, true
})

//GetJob returns the Job associated with an entry
var GetIndex = GetterFunc(func(entry Entry) (interface{}, bool) {
	return entry.Index, true
})

//GetLogLevel returns the LogLevel associated with an entry
var GetLogLevel = GetterFunc(func(entry Entry) (interface{}, bool) {
	return entry.LogLevel, true
})

//GetSource returns the source (typically process-name) associated with an entry
var GetSource = GetterFunc(func(entry Entry) (interface{}, bool) {
	return entry.Source, true
})

//GetMessage returns the message associated with an entry
var GetMessage = GetterFunc(func(entry Entry) (interface{}, bool) {
	return entry.Message, true
})

//GetSession returns the session associated with an entry
var GetSession = GetterFunc(func(entry Entry) (interface{}, bool) {
	return entry.Session, true
})

//DataGetter returns a Getter that can extract data from an Entry's Data field
//DataGetter takes multiple keys.  These are tried in order -- if a key is found in the Data field, the corresponding value is returned.
//A key can be a full-blown JSON path (e.g. `foo.bar.baz`) -- DataGetter will traverse the Data field as far as possible to fetch the corresponding value.
//
//These behaviors combine well wtih entries.GroupBy.  In particular, we are often inconsistent with how we name keys in our lager.Data -- sometimes this is intentional
//as a message passes from one layer of abstraction to another.  For example, TaskGuid becomes Guid becomes Container.Handle as it flows from Receptor=>Rep=>Executor=>Garden.
//
//To group by TaskGuid one can
//
//	entries.GroupBy(DataGetter("TaskGuid", "Guid", "Container.Handle"))
func DataGetter(keys ...string) Getter {
	transformationMap := TransformationMap{}
	for _, key := range keys {
		transformationMap[key] = NoOpTransformation
	}
	return TransformingGetter(transformationMap)
}

type TransformationFunction func(interface{}) (interface{}, bool)
type TransformationMap map[string]TransformationFunction

var NoOpTransformation = func(d interface{}) (interface{}, bool) {
	return d, true
}

var TrimTransformation = func(d interface{}) (interface{}, bool) {
	bs, err := json.Marshal(d)
	if err != nil {
		return nil, false
	}
	return strings.Trim(string(bs), `"`), true
}

func TrimWithPrefixTransformation(prefix string) TransformationFunction {
	return func(d interface{}) (interface{}, bool) {
		bs, err := json.Marshal(d)
		if err != nil {
			return nil, false
		}

		s := strings.Trim(string(bs), `"`)
		if !strings.Contains(s, prefix) {
			return nil, false
		}
		return strings.TrimPrefix(s, prefix), true
	}
}

func TransformingGetter(transformations TransformationMap) Getter {
	return GetterFunc(func(entry Entry) (interface{}, bool) {
		for k, f := range transformations {
			subKeys := strings.Split(k, ".")
			if rawValue, ok := getSubKey(entry.Data, subKeys); ok {
				return f(rawValue)
			}
		}

		return nil, false
	})
}

func getSubKey(data map[string]interface{}, subKeys []string) (interface{}, bool) {
	v, ok := data[subKeys[0]]
	if !ok {
		return nil, false
	}
	if len(subKeys) == 1 {
		return v, true
	}
	subData, ok := v.(map[string]interface{})
	if !ok {
		return nil, false
	}
	return getSubKey(subData, subKeys[1:])
}
