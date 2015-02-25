package functions

import (
	"fmt"
	"image/color"
	"path/filepath"

	"code.google.com/p/plotinum/plot"

	. "github.com/onsi/cicerone/dsl"
	"github.com/onsi/cicerone/viz"
	"github.com/onsi/say"
)

func AnalyzeCFPushes(e Entries, outputDir string) error {
	byApplication := e.GroupBy(DataGetter("identifier"))
	firstEntry := byApplication.Entries[0][0]

	timelineDescription := TimelineDescription{
		{"Creating", MatchMessage(`Created app with guid`)},
		{"CC-Says-Start", MatchMessage(`Updated app with guid .* \(\{"state"=>"STARTED"\}\)`)},
		{"Creating-Stg", And(MatchMessage(`Creating container`), MatchSource("STG:STG/0"))},
		{"Created-Stg", And(MatchMessage(`Successfully created container`), MatchSource("STG:STG/0"))},
		{"Finish-DL-App", MatchMessage(`Downloaded app package`)},
		{"Finish-DL-Buildpack", MatchMessage(`Downloaded buildpacks`)},
		{"Finish-Builder", MatchMessage(`Staging complete`)},
		{"Finish-Upload", MatchMessage(`Uploading complete`)},
		{"Creating-Inst", And(MatchMessage(`Creating container`), MatchSource("CELL:CELL/0"))},
		{"Created-Inst", And(MatchMessage(`Successfully created container`), MatchSource("CELL:CELL/0"))},
		{"Healthy", MatchMessage(`healthcheck passed`)},
	}

	timelines := byApplication.ConstructTimelines(timelineDescription, firstEntry)
	completeTimelines := timelines.CompleteTimelines()
	say.Println(0, say.Red("Complete Timelines: %d/%d (%.2f%%)\n",
		len(completeTimelines),
		len(timelines),
		float64(len(completeTimelines))/float64(len(timelines))*100.0))

	plotCFPushesTimelinesAndHistograms(completeTimelines, outputDir, "real-time")

	for i := range completeTimelines {
		completeTimelines[i].ZeroEntry = completeTimelines[i].Entries[0]
	}

	fmt.Println(completeTimelines.DTStatsSlice())
	plotCFPushesTimelinesAndHistograms(completeTimelines, outputDir, "anchored")

	return nil
}

func plotCFPushesTimelinesAndHistograms(timelines Timelines, outputDir string, prefix string) {
	histograms := viz.NewUniformBoard(len(timelines.Description()), 2, 0.01)

	for i, timelinePoint := range timelines.Description() {
		entryPairs := timelines.EntryPairs(i)
		p, _ := plot.New()
		p.Title.Text = timelinePoint.Name
		h := viz.NewEntryPairsHistogram(entryPairs, 30)
		h.Color = color.RGBA{0, 0, 0, 255}
		p.Add(h)

		twentyPercent := int(float64(len(entryPairs)) * 0.2)
		h = viz.NewEntryPairsHistogram(entryPairs[:twentyPercent], 30)
		h.Color = color.RGBA{0, 0, 255, 255}
		p.Add(h)

		h = viz.NewEntryPairsHistogram(entryPairs[len(entryPairs)-twentyPercent:], 30)
		h.Color = color.RGBA{255, 0, 0, 255}
		p.Add(h)

		histograms.AddNextSubPlot(p)
	}

	for i, timelinePoint := range timelines.Description() {
		entryPairs := timelines.EntryPairs(i)
		p, _ := plot.New()
		p.Title.Text = timelinePoint.Name
		h := viz.NewScaledEntryPairsHistogram(entryPairs, 30, 0, timelines.EndsAfter())
		h.Color = color.RGBA{0, 0, 0, 255}
		p.Add(h)
		histograms.AddNextSubPlot(p)
	}
	fmt.Println(filepath.Join(outputDir, prefix+"-histograms.png"))
	histograms.Save(3.0*float64(len(timelines.Description())), 6.0, filepath.Join(outputDir, prefix+"-histograms.png"))

	timelineBoard := &viz.Board{}
	p, _ := plot.New()
	p.Title.Text = "Timelines"
	p.Add(viz.NewTimelinesPlotter(timelines, timelines.StartsAfter().Seconds(), timelines.EndsAfter().Seconds()))
	timelineBoard.AddSubPlot(p, viz.Rect{0, 0, 1.0, 1.0})
	timelineBoard.Save(16.0, 20.0, filepath.Join(outputDir, prefix+"-timelines.png"))
}
