package functions

import (
	"fmt"
	"image/color"
	"time"

	"code.google.com/p/plotinum/plot"
	. "github.com/onsi/sommelier/dsl"
	"github.com/onsi/sommelier/viz"
)

func VizziniParallelGarden(e Entries) error {
	byTaskGuid := e.GroupBy(DataGetter("handle"))

	//can we get Creating and Created in here? need to pass GroupBy a Getter than can parse JSON data, etc...
	timelineDescription := TimelineDescription{
		{"Memory", MatchMessage(`garden-server\.limit-memory\.limited`)},
		{"Disk", MatchMessage(`garden-server\.limit-disk\.limited`)},
		{"CPU", MatchMessage(`garden-server\.limit-cpu\.limited`)},
		{"Start-Running", MatchMessage(`garden-server\.run\.spawned`)},
		{"Finish-Running", MatchMessage(`garden-server\.run\.exited`)},
		{"Start-Streaming", MatchMessage(`garden-server\.stream-out\.streaming-out`)},
		{"Finish-Streaming", MatchMessage(`garden-server\.stream-out\.streamed-out`)},
	}

	firstEvent, _ := e.First(MatchMessage(`garden-server.create.creating`))

	timelines := byTaskGuid.ConstructTimelines(timelineDescription, firstEvent)
	fmt.Println(timelines)

	fmt.Println(timelines.DTStatsSlice())

	histograms := viz.NewUniformBoard(7, 2, 0.01)

	for i, timelinePoint := range timelineDescription {
		entryPairs := timelines.EntryPairs(i)
		p, _ := plot.New()
		p.Title.Text = timelinePoint.Name
		h := viz.NewEntryPairsHistogram(entryPairs, 30)
		h.Color = color.RGBA{0, 0, 255, 255}
		p.Add(h)
		histograms.AddNextSubPlot(p)
	}

	for i, timelinePoint := range timelineDescription {
		entryPairs := timelines.EntryPairs(i)
		p, _ := plot.New()
		p.Title.Text = timelinePoint.Name
		h := viz.NewScaledEntryPairsHistogram(entryPairs, 30, 0, 45*time.Second)
		h.Color = color.RGBA{255, 0, 0, 255}
		p.Add(h)
		histograms.AddNextSubPlot(p)
	}

	histograms.Save(21.0, 6.0, "test.png")

	timelineBoard := &viz.Board{}
	p, _ := plot.New()
	p.Add(viz.NewTimelinesPlotter(timelines, timelines.StartsAfter().Seconds(), timelines.EndsAfter().Seconds()))
	timelineBoard.AddSubPlot(p, viz.Rect{0, 0, 1.0, 1.0})
	timelineBoard.Save(10.0, 10.0, "timelines.png")

	return nil
}
