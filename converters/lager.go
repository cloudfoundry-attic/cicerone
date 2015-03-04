package converters

import (
	"os"

	. "github.com/onsi/cicerone/dsl"
	"github.com/pivotal-golang/lager/chug"
)

func EntriesFromLagerFile(filename string) (Entries, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	out := make(chan chug.Entry)
	go chug.Chug(file, out)

	entries := Entries{}
	for chugEntry := range out {
		entry, err := NewEntryFromChugLog(chugEntry)
		if err != nil {
			continue
		}
		entries = append(entries, entry)
	}

	return entries, nil
}
