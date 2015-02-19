package main

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
			if v, ok := entry.Data[k]; ok {
				return v, true
			}
		}

		return nil, false
	})
}
