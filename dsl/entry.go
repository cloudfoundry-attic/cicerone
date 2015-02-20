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

type Entry struct {
	chug.LogEntry
	Job   string
	Index int
	ID    int
}

func (e Entry) IsZero() bool {
	return e.ID == 0
}

func (e Entry) VM() string {
	return fmt.Sprintf("%s/%d", e.Job, e.Index)
}

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
