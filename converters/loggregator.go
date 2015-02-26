package converters

import (
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/cicerone/dsl"

	"github.com/pivotal-golang/lager"
)

var loggregatorRegExp *regexp.Regexp

func init() {
	loggregatorRegExp = regexp.MustCompile(`(\d\d\d\d-\d\d-\d\dT\d\d:\d\d:\d\d\.\d\d)\+\d\d\d\d \[(.*)/(\d+)\]\s+(OUT|ERR) (.*)`)
}

// EntriesFromLoggregatorLogs takes a file generated via loggregator output
// and generates Cicerone entries.  The log-lines are not assumed to be lager logs.
// Job and Source correspond to the loggregator source (e.g. APP, CELL)
// Index corresponds to the loggregator index (e.g. APP/2 yields 2)
func EntriesFromLoggregatorLogs(filename string) (Entries, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return Entries{}, err
	}

	logs := strings.Split(string(data), "\n")
	entries := Entries{}
	for _, log := range logs {
		results := loggregatorRegExp.FindStringSubmatch(log)
		if results != nil {
			entry := Entry{}

			entry.Timestamp, _ = time.Parse("2006-01-02T15:04:05.00", results[1])
			entry.Source = results[2]
			entry.Job = results[2]
			entry.Index, _ = strconv.Atoi(results[3])
			if results[4] == "OUT" {
				entry.LogLevel = lager.INFO
			} else {
				entry.LogLevel = lager.ERROR
			}
			entry.Message = results[5]

			entries = append(entries, entry)
		}
	}

	return entries, nil
}
