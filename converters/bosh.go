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

// EntriesFromBOSHTree takes a path to a directory that looks like:
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
// min-time and max-time are used to limit the window of time in which to import logs
// The final set of logs are orderd by time (which likely varies from box to box!)
//
// Job corresponds to the bosh job extracted from the directory
// Index corresponds to the bosh index extracted from the directory
func EntriesFromBOSHTree(path string, minTime time.Time, maxTime time.Time) (Entries, error) {
	entries := map[string]map[string]map[string]Entries{}

	infos, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	vmEntriesChans := map[string]chan map[string]map[string]Entries{}
	vmErrorChans := map[string]chan error{}
	for _, info := range infos {
		if !info.IsDir() {
			continue
		}

		vm := info.Name()
		matches := boshTreeSubDirRegExp.FindStringSubmatch(vm)
		if matches == nil {
			continue
		}

		job := matches[1]
		index, _ := strconv.Atoi(matches[2])

		vmEntriesChans[vm] = make(chan map[string]map[string]Entries, 1)
		vmErrorChans[vm] = make(chan error, 1)
		go func(p string, min, max time.Time, j string, i int, entriesChan chan map[string]map[string]Entries, errChan chan error) {
			e, err := entriesFromBOSHProcesses(p, min, max, j, i)
			if err != nil {
				errChan <- err
			} else {
				entriesChan <- e
			}
		}(filepath.Join(path, vm), minTime, maxTime, job, index, vmEntriesChans[vm], vmErrorChans[vm])
	}

	for vm := range vmEntriesChans {
		select {
		case e := <-vmEntriesChans[vm]:
			entries[vm] = e
		case err := <-vmErrorChans[vm]:
			return nil, err
		}
	}

	allEntries := Entries{}
	for vm, vmEntries := range entries {
		say.Println(0, say.Green(vm))
		for process, processEntries := range vmEntries {
			say.Println(1, say.Yellow(process))
			for file, fileEntries := range processEntries {
				lineCountMessage := ""
				if len(fileEntries) == 0 {
					lineCountMessage = say.Red("EMPTY")
				} else {
					lineCountMessage = say.Green("%d", len(fileEntries))
				}
				say.Println(2, "%s %s", file, lineCountMessage)
				allEntries = append(allEntries, fileEntries...)
			}
		}
	}

	say.Println(0, "Sorting %d lines", len(allEntries))

	sort.Sort(allEntries)

	return allEntries, nil
}

func entriesFromBOSHProcesses(path string, minTime time.Time, maxTime time.Time, job string, index int) (map[string]map[string]Entries, error) {
	entries := map[string]map[string]Entries{}

	infos, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	processEntriesChans := map[string]chan map[string]Entries{}
	processErrorChans := map[string]chan error{}
	for _, info := range infos {
		if !info.IsDir() {
			continue
		}

		process := info.Name()
		processEntriesChans[process] = make(chan map[string]Entries, 1)
		processErrorChans[process] = make(chan error, 1)
		go func(p string, min, max time.Time, j string, i int, ps string, entriesChan chan map[string]Entries, errChan chan error) {
			e, err := entriesFromBOSHProcess(p, min, max, j, i, ps)
			if err != nil {
				errChan <- err
			} else {
				entriesChan <- e
			}
		}(filepath.Join(path, process), minTime, maxTime, job, index, process, processEntriesChans[process], processErrorChans[process])
	}

	for process := range processEntriesChans {
		select {
		case err := <-processErrorChans[process]:
			return nil, err
		case e := <-processEntriesChans[process]:
			entries[process] = e
		}
	}

	return entries, nil
}

func entriesFromBOSHProcess(path string, minTime time.Time, maxTime time.Time, job string, index int, process string) (map[string]Entries, error) {
	entries := map[string]Entries{}

	infos, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, info := range infos {
		if info.IsDir() {
			continue
		}

		file := info.Name()

		f, err := os.Open(filepath.Join(path, file))
		if err != nil {
			return nil, err
		}
		out := make(chan chug.Entry)
		go chug.Chug(f, out)

		say.Println(0, "%s %s/%d [%s] %s", say.Green("Processing"), job, index, process, file)
		fileEntries := Entries{}
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

			fileEntries = append(fileEntries, entry)
		}
		say.Println(0, "%s       %s/%d [%s] %s", say.Yellow("Done"), job, index, process, file)
		entries[file] = fileEntries
	}

	return entries, nil
}
