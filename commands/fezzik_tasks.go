package commands

import (
	"fmt"
	"path/filepath"
	"strings"

	"code.google.com/p/plotinum/plot"

	"github.com/onsi/cicerone/converters"
	. "github.com/onsi/cicerone/dsl"
	"github.com/onsi/cicerone/viz"
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
		e = e.Filter(RegExpMatcher(DataGetter("task-guid", "container-guid", "guid", "container.guid"), args[1]))
	}

	fmt.Println("Receptors that handled creates:", e.Filter(MatchMessage(`create\.creating-task`)).GroupBy(GetVM).Keys)
	fmt.Println("Receptors that handled resolves:", e.Filter(MatchMessage(`resolved-task`)).GroupBy(GetVM).Keys)

	byTaskGuid := e.GroupBy(DataGetter("task-guid", "container-guid", "guid", "container.guid"))

	startToEndTimelineDescription := TimelineDescription{
		// receptor says create.creating-task when it hears about our task
		{"Creating", MatchMessage(`create\.creating-task`)},
		// receptor says create.requesting-task-auction after it has bbs.DesiredTask
		{"Persisted-Task", MatchMessage(`create\.requesting-task-auction`)},
		// receptor says create.did-fetch-auctioneer-address after it fetches the auctioneer address from the BBS
		{"Fetched-Auctioneer-Addr", MatchMessage(`create\.did-fetch-auctioneer-address`)},
		// receptor says create.created after the auction has been submitted (this entails a round-trip to the auctioneer)
		{"Auction-Submitted", MatchMessage(`task-handler\.create\.created`)},
		// executor says allocating-container when the rep asks it to allocate a container for the task (this measures how long it took the auction to place the task on the rep)
		{"Allocating-Container", MatchMessage(`\.allocating-container`)},
		// the rep says processing-reserved-container when the executor emits the allocation event
		{"Notified-Of-Allocation", MatchMessage(`\.processing-reserved-container`)},
		// the rep says succeeded-starting-task when it succesfully transitions the task from PENDING to RUNNING in the BBS
		{"Running-In-BBS", MatchMessage(`\.succeeded-starting-task`)},
		// the executor says succeded-creating-container-in-garden when the garden container is created and ready to go
		{"Created-Container", MatchMessage(`\.succeeded-creating-container-in-garden`)},
		// the rep says task-processor.completing-task when it hears the task is complete
		{"Completing-Task", MatchMessage(`task-processor\.completing-task`)},
		// the rep says succeeded-completing-task when it transitions the task from RUNNING to COMPLETE
		{"Persisted-Completed", MatchMessage(`task-processor\.succeeded-completing-task`)},
		// the receptor says resolved-task when it transitions the task to RESOLVED (after hitting the fezzik callback)
		{"Resolved", MatchMessage(`resolved-task`)},
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
	plotFezzikTaskTimelinesAndHistograms(startToEndTimelines, outputDir, "end-to-end", 7)

	startToScheduledTimelineDescription := TimelineDescription{
		{"Creating", MatchMessage(`create\.creating-task`)},
		{"Persisted-Task", MatchMessage(`create\.requesting-task-auction`)},
		{"Fetched-Auctioneer-Addr", MatchMessage(`create\.did-fetch-auctioneer-address`)},
		{"Auction-Submitted", MatchMessage(`create\.created`)},
		{"Allocating-Container", MatchMessage(`\.allocating-container`)},
		{"Notified-Of-Allocation", MatchMessage(`\.processing-reserved-container`)},
		{"Running-In-BBS", MatchMessage(`\.succeeded-starting-task`)},
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
	histograms.Save(3.0*float64(len(timelines.Description())), 6.0, filepath.Join(outputDir, prefix+"-histograms.png"))

	correlationBoard, _ := viz.NewCorrelationBoard(timelines)
	err := correlationBoard.Save(24.0, 24.0, filepath.Join(outputDir, prefix+"-correlation.png"))
	if err != nil {
		fmt.Println(err.Error())
	}

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
