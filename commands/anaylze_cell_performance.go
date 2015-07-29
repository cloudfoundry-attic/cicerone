package commands

import (
	"fmt"
	"strconv"
	"time"

	"github.com/cloudfoundry-incubator/cicerone/converters"
	. "github.com/cloudfoundry-incubator/cicerone/dsl"
	"github.com/onsi/say"
)

func init() {}

type AnalyzeCellPerformance struct{}

func (f *AnalyzeCellPerformance) Usage() string {
	return "analyze-cell-performance REP-LOG-FILE STARTING-TIMESTAMP ENDING-TIMESTAMP"
}

func (f *AnalyzeCellPerformance) Description() string {
	return `
Takes a single rep log file and two timestamps. 

Analyze-cell-performance then generates timeline plots and histograms
for the durations of key events.

e.g. analyze-cell-performance ~/rep.stdout.log
`
}

func (f *AnalyzeCellPerformance) Command(outputDir string, args ...string) error {
	if len(args) < 3 {
		return fmt.Errorf("Expected a rep log file and some timestamps")
	}

	entries, err := converters.EntriesFromLagerFile(args[0])
	if err != nil {
		return err
	}

	afterUnix, _ := strconv.Atoi(args[1])
	beforeUnix, _ := strconv.Atoi(args[2])

	after := time.Unix(int64(afterUnix), 0)
	before := time.Unix(int64(beforeUnix), 0)

	bySession := entries.Filter(MatchSource("rep")).Filter(MatchBetween(after, before)).GroupBy(GetSession)

	fmt.Printf("Found %d log entries\n", len(entries))
	fmt.Printf("Found %d sessions\n", len(bySession.Keys))

	bulkCycleTimelineDescription := TimelineDescription{
		{"Starting", MatchMessage(`sync\.starting`)},
		{"Finished", MatchMessage(`sync\.finished`)},
	}

	auctionFetchingTimelineDescription := TimelineDescription{
		{"StartFetching", MatchMessage(`rep.auction-fetch-state.handling`)},
		{"FinishedFetching", MatchMessage(`rep.auction-fetch-state.success`)},
	}

	auctionPerformingTimelineDescription := TimelineDescription{
		{"StartPerforming", MatchMessage(`rep.auction-perform-work.handling`)},
		{"FinishedPerforming", MatchMessage(`rep.auction-perform-work.success`)},
	}

	containerMetricTimelineDescription := TimelineDescription{
		{"StartFetching", MatchMessage(`rep.container-metrics-reporter.tick.started`)},
		{"FinishedFetching", MatchMessage(`rep.container-metrics-reporter.tick.done`)},
	}

	fmt.Printf("Average Bulk Sync Duration: %v\n", calculateAverageTime(bulkCycleTimelineDescription, bySession))
	fmt.Printf("Average Auction Fetching Duration: %v\n", calculateAverageTime(auctionFetchingTimelineDescription, bySession))
	fmt.Printf("Average Auction Performing Duration: %v\n", calculateAverageTime(auctionPerformingTimelineDescription, bySession))
	fmt.Printf("Average Fetching Container Metric Duration: %v\n", calculateAverageTime(containerMetricTimelineDescription, bySession))

	return nil
}

func calculateAverageTime(timelineDescription TimelineDescription, groupedEntries *GroupedEntries) time.Duration {
	timelines, err := groupedEntries.ConstructTimelines(timelineDescription)
	if err != nil {
		return 0 * time.Second
	}

	completeTimelines := timelines.CompleteTimelines()
	say.Println(0, say.Red("Complete Timelines: %d/%d (%.2f%%)\n",
		len(completeTimelines),
		len(timelines),
		float64(len(completeTimelines))/float64(len(timelines))*100.0))

	var totalDuration int64
	for _, timeline := range completeTimelines {
		totalDuration += timeline.EndsAt().UnixNano() - timeline.BeginsAt().UnixNano()
	}

	averageDuration := totalDuration / int64(len(completeTimelines))
	avgDuration, err := time.ParseDuration(fmt.Sprintf("%dns", averageDuration))
	if err != nil {
		return 0 * time.Second
	}

	return avgDuration
}
