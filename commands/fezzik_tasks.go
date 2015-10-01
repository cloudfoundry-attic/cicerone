package commands

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gonum/plot"

	"github.com/cloudfoundry-incubator/cicerone/converters"
	. "github.com/cloudfoundry-incubator/cicerone/dsl"
	"github.com/cloudfoundry-incubator/cicerone/viz"
	"github.com/onsi/say"
)

type FezzikTasks struct{}

func (f *FezzikTasks) Usage() string {
	return "fezzik-tasks UNIFIED_BOSH_LOGS <OPTIONAL-TASK-GUID-PREFIX>"
}

func (f *FezzikTasks) Description() string {
	return `
Takes a unified bosh log file that covers a Fezzik Task run
and generates timeline plots for all Tasks and histograms
for the durations of key events.

e.g. fezzik-tasks ~/workspace/performance/10-cells/fezzik-40xtasks/optimization-4-better-logs.log
`
}

func (f *FezzikTasks) Command(outputDir string, args ...string) error {
	if len(args) == 0 {
		return fmt.Errorf("First argument must be a lager file")
	}

	e, err := converters.EntriesFromLagerFile(args[0])
	if err != nil {
		return err
	}

	if len(args) == 2 {
		e = e.Filter(RegExpMatcher(DataGetter("task-guid", "container-guid", "guid", "container.guid", "allocation-request.Guid", "handle"), args[1]))
	}

	fmt.Println("BBSs that handled creates:", e.Filter(MatchMessage(`desire-task\.starting`)).GroupBy(GetVM).Keys)
	fmt.Println("BBSs that handled resolves:", e.Filter(MatchMessage(`resolved-task`)).GroupBy(GetVM).Keys)

	byTaskGuid := e.GroupBy(DataGetter("task-guid", "container-guid", "guid", "container.guid", "allocation-request.Guid", "handle"))

	startToEndTimelineDescription := TimelineDescription{
		// bbs says desire-task.starting when it hears about our task
		{"Desiring-Task", MatchMessage(`desire-task\.starting`), 1},
		// bbs says the task is persisted
		{"Persisted-Task", MatchMessage(`desire-task\.succeeded-persisting-task`), 1},
		// bbs says create.created after the auction has been submitted (this entails a round-trip to the auctioneer)
		{"Auction-Submitted", MatchMessage(`desire-task\.finished`), 1},
		// executor says allocating-container when the rep asks it to allocate a container for the task (this measures how long it took the auction to place the task on the rep)
		{"Allocating-Container", MatchMessage(`\.allocating-container`), 1},
		// the rep says processing-reserved-container when the executor emits the allocation event
		{"Notified-Of-Allocation", MatchMessage(`\.processing-reserved-container`), 1},
		// the rep says succeeded-starting-task when it succesfully transitions the task from PENDING to RUNNING in the BBS
		{"Running-In-BBS", MatchMessage(`start-task\.finished`), 1},
		// the executor says succeded-creating-container-in-garden when the garden container is created and ready to go
		{"Created-Container", MatchMessage(`\.succeeded-creating-garden-container`), 1},
		// setting up egress rules for task container
		{"Set-Up-Container-Network", MatchMessage(`\.succeeded-setting-up-net-out`), 1},
		{"Started-Running", MatchMessage(`\.run-step-process\.succeeded-transitioning-to-running`), 1},
		{"Process-Created", MatchMessage(`successful-process-create`), 1},
		{"Process-Exited", MatchMessage(`process-exit`), 1},
		{"Finished-Running", MatchMessage(`\.run-step-process\.finished`), 1},
		// the rep says that its fetching the result file
		{"Fetching-Container-Result", MatchMessage(`\.fetching-container-result`), 1},
		// the rep says task-processor.completing-task when it hears the task is complete
		{"Fetched-Container-Result", MatchMessage(`task-processor\.completing-task`), 1},
		// the rep says succeeded-completing-task when it transitions the task from RUNNING to COMPLETE
		{"Persisted-Completed", MatchMessage(`task-processor\.succeeded-completing-task`), 1},
		// the bbs says resolved-task when it transitions the task to RESOLVED (after hitting the fezzik callback)
		{"Resolved", MatchMessage(`resolved-task`), 1},
	}

	say.Println(0, say.Green("Distribution"))
	byVM := e.Filter(MatchMessage(`\.allocating-container`)).GroupBy(GetVM)
	byVM.EachGroup(func(key interface{}, entries Entries) error {
		say.Println(1, "%s: %s", say.Green("%s", key), strings.Repeat("+", len(entries)))
		return nil
	})

	startToEndTimelines, err := byTaskGuid.ConstructTimelines(startToEndTimelineDescription)
	if err != nil {
		return err
	}
	completeStartToEndTimelines := startToEndTimelines.CompleteTimelines()

	say.Println(0, say.Red("Complete Start-To-End Timelines: %d/%d (%.2f%%)",
		len(completeStartToEndTimelines),
		len(startToEndTimelines),
		float64(len(completeStartToEndTimelines))/float64(len(startToEndTimelines))*100.0))
	//	fmt.Println(completeStartToEndTimelines.DTStatsSlice())
	plotFezzikTaskTimelinesAndHistograms(startToEndTimelines, outputDir, "end-to-end", 7)

	startToScheduledTimelineDescription := TimelineDescription{
		// bbs says desire-task.starting when it hears about our task
		{"Desiring-Task", MatchMessage(`desire-task\.starting`), 1},
		{"Persisted-Task", MatchMessage(`desire-task\.succeeded-persisting-task`), 1},
		// bbs says create.created after the auction has been submitted (this entails a round-trip to the auctioneer)
		{"Auction-Submitted", MatchMessage(`desire-task\.finished`), 1},
		// executor says allocating-container when the rep asks it to allocate a container for the task (this measures how long it took the auction to place the task on the rep)
		{"Allocating-Container", MatchMessage(`\.allocating-container`), 1},
		// the rep says processing-reserved-container when the executor emits the allocation event
		{"Notified-Of-Allocation", MatchMessage(`\.processing-reserved-container`), 1},
		// the rep says succeeded-starting-task when it succesfully transitions the task from PENDING to RUNNING in the BBS
		{"Running-In-BBS", MatchMessage(`start-task\.finished`), 1},
	}

	startToScheduledTimelines, err := byTaskGuid.ConstructTimelines(startToScheduledTimelineDescription)
	if err != nil {
		return err
	}
	completeStartToScheduledTimelines := startToScheduledTimelines.CompleteTimelines()
	say.Println(0, say.Red("Complete Start-To-Scheduled Timelines: %d/%d (%.2f%%)",
		len(completeStartToScheduledTimelines),
		len(startToScheduledTimelines),
		float64(len(completeStartToScheduledTimelines))/float64(len(startToScheduledTimelines))*100.0))
	fmt.Println(startToScheduledTimelines.DTStatsSlice())
	plotFezzikTaskTimelinesAndHistograms(startToScheduledTimelines, outputDir, "scheduling", 0)

	return nil
}

func plotFezzikTaskTimelinesAndHistograms(timelines Timelines, outputDir string, prefix string, vmEventIndex int) {
	histograms := viz.NewEntryPairsHistogramBoard(timelines)
	histograms.Save(3.0*float64(len(timelines.Description())), 6.0, filepath.Join(outputDir, prefix+"-histograms.svg"))

	correlationBoard, _ := viz.NewCorrelationBoard(timelines)
	err := correlationBoard.Save(24.0, 24.0, filepath.Join(outputDir, prefix+"-correlation.svg"))
	if err != nil {
		fmt.Println(err.Error())
	}

	timelines.SortByEndTime()
	timelineBoard := &viz.Board{}
	p, _ := plot.New()
	p.Title.Text = "Timelines by End Time"
	p.Add(viz.NewTimelinesPlotter(timelines, timelines.StartsAfter().Seconds(), timelines.EndsAfter().Seconds()))
	timelineBoard.AddSubPlot(p, viz.Rect{0, 0, 1.0, 1.0})
	timelineBoard.Save(16.0, 10.0, filepath.Join(outputDir, prefix+"-timelines-by-end-time.svg"))

	timelines.SortByStartTime()
	timelineBoard = &viz.Board{}
	p, _ = plot.New()
	p.Title.Text = "Timelines by Start Time"
	p.Add(viz.NewTimelinesPlotter(timelines, timelines.StartsAfter().Seconds(), timelines.EndsAfter().Seconds()))
	timelineBoard.AddSubPlot(p, viz.Rect{0, 0, 1.0, 1.0})
	timelineBoard.Save(16.0, 10.0, filepath.Join(outputDir, prefix+"-timelines-by-start-time.svg"))

	//which VM?
	timelines.SortByVMForEntryAtIndex(vmEventIndex)
	timelineBoard = &viz.Board{}
	p, _ = plot.New()
	p.Title.Text = "Timelines by VM"
	p.Add(viz.NewTimelinesPlotter(timelines, timelines.StartsAfter().Seconds(), timelines.EndsAfter().Seconds()))
	timelineBoard.AddSubPlot(p, viz.Rect{0, 0, 1.0, 1.0})
	timelineBoard.Save(16.0, 10.0, filepath.Join(outputDir, prefix+"-timelines-by-vm.svg"))
}
