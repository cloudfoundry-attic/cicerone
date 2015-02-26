package commands

import (
	"fmt"
	"path/filepath"
	"regexp"

	"code.google.com/p/plotinum/plot"

	"github.com/onsi/cicerone/converters"
	. "github.com/onsi/cicerone/dsl"
	"github.com/onsi/cicerone/viz"
	"github.com/onsi/say"
)

var appGuidRegExp *regexp.Regexp

func init() {
	appGuidRegExp = regexp.MustCompile(`([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})`)
}

type AnalyzeCFPushes struct{}

func (f *AnalyzeCFPushes) Usage() string {
	return "analyze-cf-pushes APPLICATION_PUSH_LOGS..."
}

func (f *AnalyzeCFPushes) Description() string {
	return `
Takes a list of application push log file that each contain the
cf logs --recent output of cf push.

Analyze-cf-pushes then generates timeline plots for each application and histograms
for the durations of key events.

e.g. analyze-cf-pushes ~/workspace/performance/10-cells/cf-pushes/unoptimized/raw-pushes/**/push*
`
}

func (f *AnalyzeCFPushes) Command(outputDir string, args ...string) error {
	if len(args) == 0 {
		return fmt.Errorf("Expected a list of application push logs")
	}

	byApplication, err := loadCFPushFiles(args...)
	if err != nil {
		return err
	}

	firstEntry := byApplication.Entries[0][0]

	timelineDescription := TimelineDescription{
		{"Creating", MatchMessage(`Created app with guid`)},
		{"CC-Says-Start", MatchMessage(`Updated app with guid .* \(\{"state"=>"STARTED"\}\)`)},
		{"Creating-Stg", And(MatchMessage(`Creating container`), MatchJob("STG"))},
		{"Created-Stg", And(MatchMessage(`Successfully created container`), MatchJob("STG"))},
		{"Finish-DL-App", MatchMessage(`Downloaded app package`)},
		{"Finish-DL-Buildpack", MatchMessage(`Downloaded buildpacks`)},
		{"Finish-Builder", MatchMessage(`Staging complete`)},
		{"Finish-Upload", MatchMessage(`Uploading complete`)},
		{"Creating-Inst", And(MatchMessage(`Creating container`), MatchJob("CELL"), MatchIndex(0))},
		{"Created-Inst", And(MatchMessage(`Successfully created container`), MatchJob("CELL"), MatchIndex(0))},
		{"Healthy", MatchMessage(`healthcheck passed`)},
	}

	timelines := byApplication.ConstructTimelines(timelineDescription, firstEntry)
	completeTimelines := timelines.CompleteTimelines()
	say.Println(0, say.Red("Complete Timelines: %d/%d (%.2f%%)\n",
		len(completeTimelines),
		len(timelines),
		float64(len(completeTimelines))/float64(len(timelines))*100.0))

	plotCFPushesTimelinesAndHistograms(completeTimelines, outputDir, "cf-pushes-real-time")

	for i := range completeTimelines {
		completeTimelines[i].ZeroEntry = completeTimelines[i].Entries[0]
	}

	fmt.Println(completeTimelines.DTStatsSlice())
	plotCFPushesTimelinesAndHistograms(completeTimelines, outputDir, "cf-pushes-anchored")

	return nil
}

func loadCFPushFiles(files ...string) (*GroupedEntries, error) {
	groups := NewGroupedEntries()
	for _, file := range files {
		entries, err := converters.EntriesFromLoggregatorLogs(file)
		if err != nil {
			return nil, err
		}

		applicationGuid, ok := getApplicationGuid(entries)
		if ok {
			groups.AppendEntries(applicationGuid, entries)
		}
	}

	return groups, nil
}

func getApplicationGuid(e Entries) (string, bool) {
	entry, found := e.First(Or(MatchMessage("Created app with guid"), MatchMessage("Updated app with guid")))
	if !found {
		return "", false
	}

	matches := appGuidRegExp.FindStringSubmatch(entry.Message)
	if matches == nil {
		return "", false
	}

	return matches[1], true
}

func plotCFPushesTimelinesAndHistograms(timelines Timelines, outputDir string, prefix string) {
	plotTimelinesHistogramsBoard(timelines, filepath.Join(outputDir, prefix+"-histograms.png"))

	timelineBoard := &viz.Board{}
	p, _ := plot.New()
	p.Title.Text = "Timelines"
	timelines.SortByStartTime()
	p.Add(viz.NewTimelinesPlotter(timelines, timelines.StartsAfter().Seconds(), timelines.EndsAfter().Seconds()))
	timelineBoard.AddSubPlot(p, viz.Rect{0, 0, 1.0, 1.0})
	timelineBoard.Save(16.0, 20.0, filepath.Join(outputDir, prefix+"-timelines.png"))
}
