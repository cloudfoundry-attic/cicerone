package functions

import (
	"fmt"
	"image/color"

	"code.google.com/p/plotinum/plot"

	. "github.com/onsi/sommelier/dsl"
	"github.com/onsi/sommelier/viz"
)

func FezzikTasks(e Entries) error {
	byTaskGuid := e.GroupBy(DataGetter("task-guid", "container-guid", "guid"))

	timelineDescription := TimelineDescription{
		{"Creating", MatchMessage(`create\.creating-task`)},
		{"Persisting", MatchMessage(`create\.requesting-task-auction`)},
		{"Auction-Submitted", MatchMessage(`create\.created`)},
		{"Starting", MatchMessage(`task-processor\.starting-task`)},
		{"Persisted-Starting", MatchMessage(`task-processor\.succeeded-starting-task`)},
		{"Created-Container", MatchMessage(`run-container\.run\.started`)},
		{"Spawned-Process", MatchMessage(`run-step-process\.succeeded-transitioning-to-running`)},
		{"Completing-Task", MatchMessage(`task-processor\.completing-task`)},
		{"Persisted-Completed", MatchMessage(`task-processor\.succeeded-completing-task`)},
		{"Resolved", MatchMessage(`resolved-task`)},
	}

	timelines := byTaskGuid.ConstructTimelines(timelineDescription, e[0])
	fmt.Println(timelines)

	fmt.Println(timelines.DTStatsSlice())

	histograms := viz.NewUniformBoard(10, 2, 0.01)

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
		h := viz.NewScaledEntryPairsHistogram(entryPairs, 30, 0, timelines.EndsAfter())
		h.Color = color.RGBA{255, 0, 0, 255}
		p.Add(h)
		histograms.AddNextSubPlot(p)
	}

	histograms.Save(30.0, 6.0, "histograms.png")

	timelines.SortByEndTime()

	timelineBoard := &viz.Board{}
	p, _ := plot.New()
	p.Add(viz.NewTimelinesPlotter(timelines, timelines.StartsAfter().Seconds(), timelines.EndsAfter().Seconds()))
	timelineBoard.AddSubPlot(p, viz.Rect{0, 0, 1.0, 1.0})
	timelineBoard.Save(16.0, 10.0, "timelines_by_end_time.png")

	timelines.SortByVMForEntryAtIndex(3)
	timelineBoard = &viz.Board{}
	p, _ = plot.New()
	p.Add(viz.NewTimelinesPlotter(timelines, timelines.StartsAfter().Seconds(), timelines.EndsAfter().Seconds()))
	timelineBoard.AddSubPlot(p, viz.Rect{0, 0, 1.0, 1.0})
	timelineBoard.Save(16.0, 10.0, "timelines_by_vm.png")

	timelines.SortByStartTime()
	timelineBoard = &viz.Board{}
	p, _ = plot.New()
	p.Add(viz.NewTimelinesPlotter(timelines, timelines.StartsAfter().Seconds(), timelines.EndsAfter().Seconds()))
	timelineBoard.AddSubPlot(p, viz.Rect{0, 0, 1.0, 1.0})
	timelineBoard.Save(16.0, 10.0, "timelines_by_start_time.png")

	return nil
}
