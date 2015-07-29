package commands

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/cloudfoundry-incubator/cicerone/converters"
)

type SlurpBosh struct{}

func (f *SlurpBosh) Usage() string {
	return "slurp-bosh BOSH_TREE MIN_TIME MAX_TIME OUTPUT"
}

func (f *SlurpBosh) Description() string {
	return `
Parses a BOSH_TREE and generates a single unified OUTPUT file
containing the loglines between MIN_TIME and MAX_TIME (passed in as unix timestamps).

A BOSH_TREE is a directory with sub-directories that look like JOB-INDEX
each containing subdirectories that are the name of a process (e.g. executor)
and contain log files.

e.g. slurp-bosh ~/workspace/performance/10-cells/cf-pushes/unoptimized/bosh-logs/ 1424820500 1424828000 $HOME/workspace/performance/10-cells/cf-pushes/unoptimized-unified-bosh-logs.log
`
}

func (f *SlurpBosh) Command(outputDir string, args ...string) error {
	if len(args) != 4 {
		return fmt.Errorf("slurp-bosh needs 4 arguments: BOSH_TREE MIN_TIME MAX_TIME OUTPUT")
	}

	minTimestamp, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return err
	}
	maxTimestamp, err := strconv.ParseInt(args[2], 10, 64)
	if err != nil {
		return err
	}
	entries, err := converters.EntriesFromBOSHTree(args[0], time.Unix(minTimestamp, 0), time.Unix(maxTimestamp, 0))
	if err != nil {
		return err
	}

	outputFile, err := os.Create(args[3])
	if err != nil {
		return err
	}

	return entries.WriteLagerFormatTo(outputFile)
}
