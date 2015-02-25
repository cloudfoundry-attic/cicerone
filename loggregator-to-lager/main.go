//For example:
//go run main.go /Users/onsi/workspace/performance/10-cells/cf-pushes/**/push-* > /Users/onsi/workspace/performance/10-cells/cf-pushes/aggregated-pushes.log
package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pivotal-golang/lager"

	. "github.com/onsi/cicerone/dsl"
)

var re *regexp.Regexp

func init() {
	re = regexp.MustCompile(`(\d\d\d\d-\d\d-\d\dT\d\d:\d\d:\d\d\.\d\d)\+\d\d\d\d \[(.*)/(\d+)\]\s+(OUT|ERR) (.*)`)
}

func main() {
	files := os.Args[1:]
	entries := Entries{}
	for _, file := range files {
		e := LogsFromFile(file, filepath.Base(file))
		entries = append(entries, e...)
	}
	sort.Sort(entries)
	entries.WriteLagerFormatTo(os.Stdout)
}

func LogsFromFile(file string, identifier string) Entries {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err.Error())
	}

	logs := strings.Split(string(data), "\n")
	entries := Entries{}
	for _, log := range logs {
		if re.MatchString(log) {
			results := re.FindStringSubmatch(log)
			entry := Entry{}

			entry.Timestamp, _ = time.Parse("2006-01-02T15:04:05.00", results[1])
			entry.Source = results[2]
			entry.Job = results[2]
			entry.Index, _ = strconv.Atoi(results[3])
			entry.Message = "imported." + results[5]
			entry.Data = lager.Data{"identifier": identifier}
			if results[4] == "OUT" {
				entry.LogLevel = lager.INFO
			} else {
				entry.LogLevel = lager.ERROR
			}

			entries = append(entries, entry)
		}
	}

	return entries
}
