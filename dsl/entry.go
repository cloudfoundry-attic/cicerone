package dsl

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/pivotal-golang/lager"

	"github.com/pivotal-golang/lager/chug"
)

//An Entry represtents a Cicerone log line
type Entry struct {
	chug.LogEntry
	Job   string
	Index int
}

func NewEntryFromChugLog(chugEntry chug.Entry) (Entry, error) {
	if !chugEntry.IsLager {
		return Entry{}, errors.New("not a chug entry")
	}

	entry := Entry{
		LogEntry: chugEntry.Log,
	}

	encodedJob, ok := entry.Data["cicerone-job"]
	if ok {
		delete(entry.Data, "cicerone-job")
		entry.Job = encodedJob.(string)
	}

	encodedIndex, ok := entry.Data["cicerone-index"]
	if ok {
		delete(entry.Data, "cicerone-index")
		entry.Index = encodedIndex.(int)
	}

	return entry, nil
}

//IsZero returns true if the Entry is the zero Entry
func (e Entry) IsZero() bool {
	return e.Source == "" && e.Message == ""
}

//VM returns a unique idenfitier of the machine that emitted the log line
//
//This corresponds to "job/index"
func (e Entry) VM() string {
	return fmt.Sprintf("%s/%d", e.Job, e.Index)
}

func (e Entry) LagerFormat() lager.LogFormat {
	data := lager.Data{}
	for k, v := range e.Data {
		data[k] = v
	}
	data["session"] = e.Session
	if e.Error != nil && e.Error.Error() != "" {
		data["error"] = e.Error.Error()
	}
	if e.Trace != "" {
		data["trace"] = e.Trace
	}
	data["cicerone-job"] = e.Job
	data["cicerone-index"] = e.Index

	return lager.LogFormat{
		Timestamp: fmt.Sprintf("%.9f", float64(e.Timestamp.UnixNano())/1e9),
		Source:    e.Source,
		Message:   e.Message,
		LogLevel:  e.LogLevel,
		Data:      data,
	}
}

//WriteLagerFormatTo emits lager formatted output to the passed-in writer
func (e Entry) WriteLagerFormatTo(w io.Writer) error {
	return json.NewEncoder(w).Encode(e.LagerFormat())
}
