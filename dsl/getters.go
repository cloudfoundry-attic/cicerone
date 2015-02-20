package dsl

import "strings"

type Getter interface {
	Get(entry Entry) (interface{}, bool)
}

type GetterFunc func(Entry) (interface{}, bool)

func (g GetterFunc) Get(entry Entry) (interface{}, bool) {
	return g(entry)
}

var GetVM = GetterFunc(func(entry Entry) (interface{}, bool) {
	return entry.VM(), true
})

var GetJob = GetterFunc(func(entry Entry) (interface{}, bool) {
	return entry.Job, true
})

var GetIndex = GetterFunc(func(entry Entry) (interface{}, bool) {
	return entry.Index, true
})

var GetLogLevel = GetterFunc(func(entry Entry) (interface{}, bool) {
	return entry.LogLevel, true
})

var GetSource = GetterFunc(func(entry Entry) (interface{}, bool) {
	return entry.Source, true
})

var GetMessage = GetterFunc(func(entry Entry) (interface{}, bool) {
	return entry.Message, true
})

func DataGetter(key ...string) Getter {
	return GetterFunc(func(entry Entry) (interface{}, bool) {
		for _, k := range key {
			subKeys := strings.Split(k, ".")
			if v, ok := getSubKey(entry.Data, subKeys); ok {
				return v, true
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
