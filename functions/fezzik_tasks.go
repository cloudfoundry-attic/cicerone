package functions

import (
	"fmt"
	"image/color"
	"path/filepath"

	"code.google.com/p/plotinum/plot"

	. "github.com/onsi/cicerone/dsl"
	"github.com/onsi/cicerone/viz"
)

func FezzikTasks(e Entries, outputDir string) error {
	fmt.Println("Receptors that handled creates:", e.Filter(MatchMessage(`create\.creating-task`)).GroupBy(GetVM).Keys)
	fmt.Println("Receptors that handled resolves:", e.Filter(MatchMessage(`resolved-task`)).GroupBy(GetVM).Keys)

	byTaskGuid := e.GroupBy(DataGetter("task-guid", "container-guid", "guid"))

	startToEndTimelineDescription := TimelineDescription{
		{"Creating", MatchMessage(`create\.creating-task`)},
		{"Persisted", MatchMessage(`create\.requesting-task-auction`)},
		{"Auction-Submitted", MatchMessage(`create\.created`)},
		{"Starting", MatchMessage(`task-processor\.starting-task`)},
		{"Persisted-Starting", MatchMessage(`task-processor\.succeeded-starting-task`)},
		{"Created-Container", MatchMessage(`run-container\.run\.started`)},
		{"Spawned-Process", MatchMessage(`run-step-process\.succeeded-transitioning-to-running`)},
		{"Completing-Task", MatchMessage(`task-processor\.completing-task`)},
		{"Persisted-Completed", MatchMessage(`task-processor\.succeeded-completing-task`)},
		{"Resolved", MatchMessage(`resolved-task`)},
	}

	startToEndTimelines := byTaskGuid.ConstructTimelines(startToEndTimelineDescription, e[0])
	plotFezzikTaskTimelinesAndHistograms(startToEndTimelines, outputDir, "end-to-end", 3)

	startToScheduledTimelineDescription := TimelineDescription{
		{"Creating", MatchMessage(`create\.creating-task`)},
		{"Persisted", MatchMessage(`create\.requesting-task-auction`)},
		{"Auction-Submitted", MatchMessage(`create\.created`)},
		{"Starting", MatchMessage(`task-processor\.starting-task`)},
		{"Persisted-Starting", MatchMessage(`task-processor\.succeeded-starting-task`)},
	}

	startToScheduledTimelines := byTaskGuid.ConstructTimelines(startToScheduledTimelineDescription, e[0])
	plotFezzikTaskTimelinesAndHistograms(startToScheduledTimelines, outputDir, "scheduling", 0)

	return nil
}

func plotFezzikTaskTimelinesAndHistograms(timelines Timelines, outputDir string, prefix string, vmEventIndex int) {
	histograms := viz.NewUniformBoard(len(timelines.Description()), 2, 0.01)

	for i, timelinePoint := range timelines.Description() {
		entryPairs := timelines.EntryPairs(i)
		p, _ := plot.New()
		p.Title.Text = timelinePoint.Name
		h := viz.NewEntryPairsHistogram(entryPairs, 30)
		h.Color = color.RGBA{0, 0, 255, 255}
		p.Add(h)
		histograms.AddNextSubPlot(p)
	}

	for i, timelinePoint := range timelines.Description() {
		entryPairs := timelines.EntryPairs(i)
		p, _ := plot.New()
		p.Title.Text = timelinePoint.Name
		h := viz.NewScaledEntryPairsHistogram(entryPairs, 30, 0, timelines.EndsAfter())
		h.Color = color.RGBA{255, 0, 0, 255}
		p.Add(h)
		histograms.AddNextSubPlot(p)
	}
	histograms.Save(3.0*float64(len(timelines.Description())), 6.0, filepath.Join(outputDir, prefix+"-histograms.png"))

	timelines.SortByEndTime()
	timelineBoard := &viz.Board{}
	p, _ := plot.New()
	p.Title.Text = "Timelines by End Time"
	p.Add(viz.NewTimelinesPlotter(timelines, timelines.StartsAfter().Seconds(), timelines.EndsAfter().Seconds()))
	timelineBoard.AddSubPlot(p, viz.Rect{0, 0, 1.0, 1.0})
	timelineBoard.Save(16.0, 10.0, filepath.Join(outputDir, prefix+"-timelines-by-end-time.png"))

	//which VM?
	timelines.SortByVMForEntryAtIndex(vmEventIndex)
	timelineBoard = &viz.Board{}
	p, _ = plot.New()
	p.Title.Text = "Timelines by VM"
	p.Add(viz.NewTimelinesPlotter(timelines, timelines.StartsAfter().Seconds(), timelines.EndsAfter().Seconds()))
	timelineBoard.AddSubPlot(p, viz.Rect{0, 0, 1.0, 1.0})
	timelineBoard.Save(16.0, 10.0, filepath.Join(outputDir, prefix+"-timelines-by-vm.png"))

	timelines.SortByStartTime()
	timelineBoard = &viz.Board{}
	p, _ = plot.New()
	p.Title.Text = "Timelines by Start Time"
	p.Add(viz.NewTimelinesPlotter(timelines, timelines.StartsAfter().Seconds(), timelines.EndsAfter().Seconds()))
	timelineBoard.AddSubPlot(p, viz.Rect{0, 0, 1.0, 1.0})
	timelineBoard.Save(16.0, 10.0, filepath.Join(outputDir, prefix+"-timelines-by-start-time.png"))
}
