package functions

import (
	"fmt"
	"image/color"
	"time"

	"code.google.com/p/plotinum/plot"

	. "github.com/onsi/sommelier/dsl"
	"github.com/onsi/sommelier/viz"
)

func FezzikTasks(e Entries) error {
	byTaskGuid := e.GroupBy(DataGetter("task-guid", "container-guid", "guid"))

	timelineDescription := TimelineDescription{
		{"Receptor-Creating", MatchMessage(`create\.creating-task`)},
		{"Receptor-Persisting-Done", MatchMessage(`create\.requesting-task-auction`)},
		{"Receptor-Auction-Submitted", MatchMessage(`create\.created`)},
		{"Rep-Starting-Task", MatchMessage(`task-processor\.starting-task`)},
		{"Rep-Succeeded-Starting-Task-In-BBS", MatchMessage(`task-processor\.succeeded-starting-task`)},
		{"Executor-Created-Container", MatchMessage(`run-container\.run\.started`)},
		{"Executor-Spawned-Process", MatchMessage(`run-step-process\.succeeded-transitioning-to-running`)},
		{"Rep-Start-Completing-Task", MatchMessage(`task-processor\.completing-task`)},
		{"Rep-Succeeded-Completing-Task-In-BBS", MatchMessage(`task-processor\.succeeded-completing-task`)},
		{"Receptor-Resolved-Task", MatchMessage(`resolved-task`)},
	}

	byTaskGuid = byTaskGuid.Filter(Or(
		MatchMessage(`task-handler\.create`),
		MatchMessage(`task-processor\.starting-task`),
		MatchMessage(`task-processor\.succeeded-starting-task`),
		MatchMessage(`run-container\.run\.started`),
		MatchMessage(`run-step-process\.succeeded-transitioning-to-running`),
		MatchMessage(`task-processor\.completing-task`),
		MatchMessage(`task-processor\.succeeded-completing-task`),
		MatchMessage(`resolving-task`),
		MatchMessage(`resolved-task`),
	))

	timelines := byTaskGuid.ConstructTimelines(timelineDescription, e[0])
	fmt.Println(timelines)

	fmt.Println(timelines.DTStatsSlice())

	board := viz.NewUniformBoard(5, 4, 0.01)

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
		h := viz.NewScaledEntryPairsHistogram(entryPairs, 30, 0, 15*time.Second)
		h.Color = color.RGBA{255, 0, 0, 255}
		p.Add(h)
		board.AddNextSubPlot(p)
	}

	return board.Save(12.0, 12.0, "test.png")

	//timeline description => pair stats (histograms)

}
