package converters

import (
	"os"
	"regexp"
	"strconv"

	. "github.com/cloudfoundry-incubator/cicerone/dsl"

	"github.com/pivotal-golang/lager/chug"
)

var papertrailRegExp *regexp.Regexp

func init() {
	papertrailRegExp = regexp.MustCompile(`\[job=([a-zA-Z0-9_-]+) index=(\d+)\]`)
}

// EntriesFromPapertrailFile takes a papertrail file and generates Cicerone Entries
// Job corresponds to the BOSH job
// Index corresponds to the BOSH index
func EntriesFromPapertrailFile(filename string) (Entries, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	out := make(chan chug.Entry)
	go chug.Chug(file, out)

	entries := Entries{}
	for chugEntry := range out {
		entry, err := newEntryFromPapertrail(chugEntry)
		if err != nil {
			continue
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

func newEntryFromPapertrail(chugEntry chug.Entry) (Entry, error) {
	entry, err := NewEntryFromChugLog(chugEntry)
	if err != nil {
		return Entry{}, err
	}

	result := papertrailRegExp.FindStringSubmatch(string(chugEntry.Raw))

	if len(result) == 3 {
		entry.Job = result[1]
		entry.Index, _ = strconv.Atoi(result[2])
	}

	return entry, nil
}
