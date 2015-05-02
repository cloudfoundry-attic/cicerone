package commands

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"code.google.com/p/plotinum/plot"

	"github.com/onsi/cicerone/converters"
	. "github.com/onsi/cicerone/dsl"
	"github.com/onsi/cicerone/viz"
	"github.com/onsi/say"
	"github.com/pivotal-golang/lager"
)

var appGuidRegExp *regexp.Regexp

func init() {
	appGuidRegExp = regexp.MustCompile(`([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})`)
}

type AnalyzeCFPushes struct{}

func (f *AnalyzeCFPushes) Usage() string {
	return "analyze-cf-pushes APPLICATION_PUSH_LOGS_GLOB_PATTERN"
}

func (f *AnalyzeCFPushes) Description() string {
	return `
Takes a glob pattern for application push log file that each contain the
'cf logs --recent' output of 'cf push'.

Analyze-cf-pushes then generates timeline plots for each application and histograms
for the durations of key events.

e.g. analyze-cf-pushes "${HOME}/workspace/performance/10-cells/cf-pushes/optimization-1-no-logs/raw-pushes/**/log*"
`
}

func (f *AnalyzeCFPushes) Command(outputDir string, args ...string) error {
	if len(args) != 1 {
		return fmt.Errorf("Expected a glob pattern for application push logs")
	}

	files, err := filepath.Glob(args[0])
	if err != nil {
		return err
	}

	byApplication, err := loadCFPushFiles(files...)
	if err != nil {
		return err
	}

	timelineDescription := TimelineDescription{
		{"Created", MatchMessage(`Updated app with guid .* \(\{"diego"=>true`)},
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

	timelines, err := byApplication.ConstructTimelines(timelineDescription)
	if err != nil {
		return err
	}
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

	plotCFPushesHistogramsByApplication(completeTimelines, outputDir, "cf-pushes-by-app")

	return nil
}

func loadCFPushFiles(files ...string) (*GroupedEntries, error) {
	groups := NewGroupedEntries()
	for _, file := range files {
		entries, err := converters.EntriesFromLoggregatorLogs(file)
		if err != nil {
			return nil, err
		}

		appType := strings.Split(filepath.Base(file), "-")[1]
		for i := range entries {
			if entries[i].Data == nil {
				entries[i].Data = lager.Data{}
			}
			entries[i].Data["app-type"] = appType
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
	timelines.SortByStartTime()

	histograms := viz.NewEntryPairsHistogramBoard(timelines)
	histograms.Save(3.0*float64(len(timelines.Description())), 6.0, filepath.Join(outputDir, prefix+"-histograms.png"))

	correlationBoard, _ := viz.NewCorrelationBoard(timelines)
	correlationBoard.Save(24.0, 24.0, filepath.Join(outputDir, prefix+"-correlation.png"))

	timelineBoard := &viz.Board{}
	p, _ := plot.New()
	p.Title.Text = "Timelines"
	p.Add(viz.NewTimelinesPlotter(timelines, timelines.StartsAfter().Seconds(), timelines.EndsAfter().Seconds()))
	timelineBoard.AddSubPlot(p, viz.Rect{0, 0, 1.0, 1.0})
	timelineBoard.Save(16.0, 20.0, filepath.Join(outputDir, prefix+"-timelines.png"))
}

func plotCFPushesHistogramsByApplication(timelines Timelines, outputDir string, prefix string) {
	timelines.SortByStartTime()
	group := timelines.GroupBy(MatchMessage(`Creating container`), DataGetter("app-type"))

	histograms := viz.NewGroupedTimelineEntryPairsHistogramBoard(group)
	histograms.Save(3.0*float64(len(timelines.Description())), 3.0, filepath.Join(outputDir, prefix+"-histograms.png"))

	correlationBoard, _ := viz.NewGroupedCorrelationBoard(group)
	correlationBoard.Save(24.0, 24.0, filepath.Join(outputDir, prefix+"-correlation.png"))

}
