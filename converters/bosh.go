package converters

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"time"

	. "github.com/onsi/cicerone/dsl"
	"github.com/onsi/say"
	"github.com/pivotal-golang/lager/chug"
)

var boshTreeSubDirRegExp *regexp.Regexp

func init() {
	boshTreeSubDirRegExp = regexp.MustCompile(`([a-zA-Z0-9_-]+)-(\d+)`)
}

// EntriesFromBoshDump takes a path to a directory that looks like:
//
// /JOB-INDEX
//           /PROCESS
//           /PROCESS
// /JOB-INDEX
//           /PROCESS
//
// For example:
//
// /cell_z1-0/executor, /cell_z1-1/receptor
//
// And slurps the whole bunch in, extracting and annotating Cicerone entries as it goes.
//
// If provided, min-time and max-time are used to limit the window of time in which to import logs
// The final set of logs are orderd by time (which likely varies from box to box!)
//
// Job corresponds to the bosh job extracted from the directory
// Index corresponds to the bosh index extracted from the directory
func EntriesFromBOSHTree(path string, minTime time.Time, maxTime time.Time) (Entries, error) {
	entries := Entries{}

	infos, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, info := range infos {
		if !info.IsDir() {
			continue
		}
		matches := boshTreeSubDirRegExp.FindStringSubmatch(info.Name())
		if matches == nil {
			continue
		}

		job := matches[1]
		index, _ := strconv.Atoi(matches[2])

		say.Println(0, say.Green("%s/%d", job, index))

		e, err := entriesFromBOSHProcesses(filepath.Join(path, info.Name()), minTime, maxTime, job, index)
		if err != nil {
			return nil, err
		}

		entries = append(entries, e...)

		if len(e) == 0 {
			say.Println(1, say.Red("%s/%d: EMPTY", job, index))
		}
	}

	say.Println(0, "Sorting %d lines", len(entries))

	sort.Sort(entries)

	return entries, nil
}

func entriesFromBOSHProcesses(path string, minTime time.Time, maxTime time.Time, job string, index int) (Entries, error) {
	entries := Entries{}

	infos, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, info := range infos {
		if !info.IsDir() {
			continue
		}

		say.Println(1, say.Yellow(info.Name()))

		e, err := entriesFromBOSHProcess(filepath.Join(path, info.Name()), minTime, maxTime, job, index)
		if err != nil {
			return nil, err
		}
		entries = append(entries, e...)

		if len(e) == 0 {
			say.Println(2, say.Red("EMPTY"))
		}
	}

	return entries, nil
}

func entriesFromBOSHProcess(path string, minTime time.Time, maxTime time.Time, job string, index int) (Entries, error) {
	entries := Entries{}

	infos, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, info := range infos {
		if info.IsDir() {
			continue
		}

		out := make(chan chug.Entry)
		f, err := os.Open(filepath.Join(path, info.Name()))
		if err != nil {
			return nil, err
		}
		go chug.Chug(f, out)

		n := 0
		for chugEntry := range out {
			if !chugEntry.IsLager {
				continue
			}
			if chugEntry.Log.Timestamp.Before(minTime) {
				continue
			}
			if chugEntry.Log.Timestamp.After(maxTime) {
				break
			}

			entry, err := NewEntryFromChugLog(chugEntry)
			if err != nil {
				continue
			}
			entry.Job = job
			entry.Index = index

			entries = append(entries, entry)
			n += 1
		}
		if n != 0 {
			say.Println(2, "%s %s", info.Name(), say.Green("%d", n))
		}
	}

	return entries, nil
}
