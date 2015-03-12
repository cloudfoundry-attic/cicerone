package commands

import (
	"fmt"
	"path/filepath"

	"code.google.com/p/plotinum/plot"

	"github.com/onsi/cicerone/converters"
	. "github.com/onsi/cicerone/dsl"
	"github.com/onsi/cicerone/viz"
	"github.com/onsi/say"
)

type FezzikLRPs struct{}

func (f *FezzikLRPs) Usage() string {
	return "fezzik-lrps UNIFIED_BOSH_LOG"
}

func (f *FezzikLRPs) Description() string {
	return `
Takes a unified BOSH log file that covers one run of an "It"
where Fezzik has launched many LRPs in parallel, and generates
timeline plots for all LRPs and histograms for the durations
of key events.

e.g. fezzik-lrps ~/workspace/performance/10-cells/fezzik-40xlrps/optimization-4-better-logs.log
`
}

func (f *FezzikLRPs) Command(outputDir string, args ...string) error {
	if len(args) != 1 {
		return fmt.Errorf("First argument must be a path to a unified BOSH log file")
	}

	e, err := converters.EntriesFromLagerFile(args[0])
	if err != nil {
		return err
	}

	byInstanceGuid := e.GroupBy(DataGetter("instance-guid", "container-guid", "guid", "container.guid", "handle"))

	lrpStartTimelineDescription := TimelineDescription{
		// Executor reserving container
		{"Reserving-Container", MatchMessage(`allocate-containers.allocating-container`)},
		// Rep marked LRP CLAIMED in BBS
		{"Claimed-Actual-LRP", MatchMessage(`claim-actual-lrp.succeeded`)},
		// Executor created actual container in Garden
		{"Created-Container", MatchMessage(`run-container.create-in-garden.succeeded-creating-garden-container`)},
		// Executor configured container (memory limits, CPU limits, port mappings, etc.)
		{"Configured-Container", MatchMessage(`run-container.create-in-garden.succeeded-getting-garden-container-info`)},
		// Fetching download
		{"Fetched-Download", MatchMessage(`run-container.run.download-step.fetch-complete`)},
		// Streamed download into container
		{"Streamed-in-Download", MatchMessage(`run-container.run.download-step.stream-in-complete`)},
		// Started Running LRP (grace) in container
		{"Started-Running-LRP", And(MatchMessage(`garden-server.run.spawned`), RegExpMatcher(DataGetter("spec.path"), `grace`))},
		// Started Running monitor process (nc) in container
		{"Started-Running-Monitor", And(MatchMessage(`garden-server.run.spawned`), RegExpMatcher(DataGetter("spec.path"), `nc`))},
		// Executor transitioning container to RUNNING
		{"Transitioned-to-Running", MatchMessage(`run-container.run.run-step-process.succeeded-transitioning-to-running`)},
		// Rep transitioned LRP to STARTED in BBS
		{"Transitioned-to-Started", MatchMessage(`start-actual-lrp.succeeded`)},
		// // Rep requesting container stop
		// {"Stopping-Container", MatchMessage(`lrp-stopper.stop.stopping`)},
		// // LRP has been cancelled
		// {"Stopped-Container", MatchMessage(`run-container.run.run-step-process.step-cancelled`)},
		// // Rep transitioned LRP to COMPLETEd in BBS
		// {"Transitioned-to-Completed", MatchMessage(`run-container.run.run-step-process.succeeded-transitioning-to-complete`)},
	}

	lrpStartTimelines, err := byInstanceGuid.ConstructTimelines(lrpStartTimelineDescription)
	if err != nil {
		return err
	}

	completeLRPStartTimelines := lrpStartTimelines.CompleteTimelines()
	say.Println(0, say.Red("Complete Starting Timelines: %d/%d (%.2f%%)\n",
		len(completeLRPStartTimelines),
		len(lrpStartTimelines),
		float64(len(completeLRPStartTimelines))/float64(len(lrpStartTimelines))*100.0))
	plotFezzikLRPTimelinesAndHistograms(completeLRPStartTimelines, outputDir, "starting", 0)

	return nil
}

func plotFezzikLRPTimelinesAndHistograms(timelines Timelines, outputDir string, prefix string, vmEventIndex int) {
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
