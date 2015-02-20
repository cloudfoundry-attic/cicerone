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

	timelineDescription := TimelineDescription{
		{"Memory", MatchMessage(`garden-server\.limit-memory\.limited`)},
		{"Disk", MatchMessage(`garden-server\.limit-disk\.limited`)},
		{"CPU", MatchMessage(`garden-server\.limit-cpu\.limited`)},
		{"Start-Running", MatchMessage(`garden-server\.run\.spawned`)},
		{"Finish-Running", MatchMessage(`garden-server\.run\.exited`)},
		{"Start-Streaming", MatchMessage(`garden-server\.stream-out\.streaming-out`)},
		{"Finish-Streaming", MatchMessage(`garden-server\.stream-out\.streamed-out`)},
	}

	timelines := byTaskGuid.ConstructTimelines(timelineDescription, e[0])
	fmt.Println(timelines)

	fmt.Println(timelines.DTStatsSlice())

	board := viz.NewUniformBoard(7, 2, 0.01)

	for i, timelinePoint := range timelineDescription {
		entryPairs := timelines.EntryPairs(i)
		p, _ := plot.New()
		p.Title.Text = timelinePoint.Name
		h := viz.NewEntryPairsHistogram(entryPairs, 30)
		h.Color = color.RGBA{0, 0, 255, 255}
		p.Add(h)
		board.AddNextSubPlot(p)
	}

	for i, timelinePoint := range timelineDescription {
		entryPairs := timelines.EntryPairs(i)
		p, _ := plot.New()
		p.Title.Text = timelinePoint.Name
		h := viz.NewScaledEntryPairsHistogram(entryPairs, 30, 0, 45*time.Second)
		h.Color = color.RGBA{255, 0, 0, 255}
		p.Add(h)
		board.AddNextSubPlot(p)
	}

	return board.Save(21.0, 6.0, "test.png")

	//timeline description => pair stats (histograms)

}
