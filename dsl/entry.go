package dsl

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"

	"github.com/pivotal-golang/lager"

	"github.com/pivotal-golang/lager/chug"
)

var jobIndexRegExp *regexp.Regexp
var entryIDCounter = 0

func init() {
	jobIndexRegExp = regexp.MustCompile(`\[job=([a-zA-Z0-9_-]+) index=(\d+)\]`)
}

//An Entry represtents a Cicerone log line
type Entry struct {
	chug.LogEntry
	Job   string
	Index int
	ID    int
}

//IsZero returns true if the Entry is the zero Entry
func (e Entry) IsZero() bool {
	return e.ID == 0
}

//VM returns a unique idenfitier of the machine that emitted the log line
//
//This corresponds to "job/index"
func (e Entry) VM() string {
	return fmt.Sprintf("%s/%d", e.Job, e.Index)
}

//WriteLagerFormatTo emits lager formatted output to the passed-in writer
func (e Entry) WriteLagerFormatTo(w io.Writer) error {
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

	l := lager.LogFormat{
		Timestamp: fmt.Sprintf("%.9f", float64(e.Timestamp.UnixNano())/1e9),
		Source:    fmt.Sprintf("%s:%s/%d", e.Source, e.Job, e.Index),
		Message:   e.Message,
		LogLevel:  e.LogLevel,
		Data:      data,
	}

	return json.NewEncoder(w).Encode(l)
}

//NewEntry takes a chug.Entry and produces a Chicerone Entry
//In particular, NewEntry analyzes the raw logline to extract the corresponding Job and Index associated with the logline
//The expected format for Job/Index is that returned by PaperTrail
//Job typically corresponds to the BOSH job
//Index typically corresponds to the BOSH index
func NewEntry(entry chug.Entry) (Entry, error) {
	if !entry.IsLager {
		return Entry{}, errors.New("not a chug entry")
	}

	result := jobIndexRegExp.FindStringSubmatch(string(entry.Raw))
	job := "none"
	index := 0
	if result != nil && len(result) == 3 {
		job = result[1]
		index, _ = strconv.Atoi(result[2])
	}

	entryIDCounter += 1

	return Entry{
		LogEntry: entry.Log,
		Job:      job,
		Index:    index,
		ID:       entryIDCounter,
	}, nil
}
