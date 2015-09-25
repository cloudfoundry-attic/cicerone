package commands

import (
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/gonum/plot"

	"github.com/cloudfoundry-incubator/cicerone/converters"
	. "github.com/cloudfoundry-incubator/cicerone/dsl"
	"github.com/cloudfoundry-incubator/cicerone/viz"
	"github.com/onsi/say"
)

type FezzikLRPs struct{}

func (f *FezzikLRPs) Usage() string {
	return "fezzik-lrps UNIFIED_BOSH_LOG PROCESS-GUID"
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
	if len(args) != 2 {
		return fmt.Errorf("First argument must be a path to a lager file, second must be a process guid")
	}

	e, err := converters.EntriesFromLagerFile(args[0])
	if err != nil {
		return err
	}

	byInstanceGuid := f.extractInstanceGuidGroups(e, args[1])

	say.Println(0, say.Green("Distribution"))
	distribution := map[interface{}]int{}
	byInstanceGuid.EachGroup(func(key interface{}, entries Entries) error {
		entry, _ := entries.First(MatchMessage(`\.allocating-container`))
		distribution[entry.VM()] += 1
		return nil
	})

	for vm, count := range distribution {
		say.Println(1, "%s: %s", say.Green("%s", vm), strings.Repeat("+", count))
	}

	lrpStartTimelineDescription := TimelineDescription{
		// Creating ActualLRP (proxy - this is the event emitted)
		{"Create-Event", MatchMessage(`watching-for-actual-lrp-changes.sending-create`)},
		// Executor reserving container
		{"Allocating", MatchMessage(`allocate-containers.allocating-container`)},
		// Rep marked LRP CLAIMED in BBS
		{"Claimed-ALRP", MatchMessage(`claim-actual-lrp.succeeded`)},
		// Executor created actual container in Garden
		{"Created-Container", MatchMessage(`run-container.create-in-garden.succeeded-creating-garden-container`)},
		// Executor configured container (memory limits, CPU limits, port mappings, etc.)
		{"Configured-Container", MatchMessage(`run-container.create-in-garden.succeeded-getting-garden-container-info`)},
		// Fetching download
		{"Fetched-Download", MatchMessage(`run-container.run.setup.download-step.fetch-complete`)},
		// Streamed download into container
		{"Streamed-in-Download", MatchMessage(`run-container.run.setup.download-step.stream-in-complete`)},
		// Started Running LRP (grace) in container
		{"Launch-Process", And(MatchMessage(`garden-server.run.spawned`), RegExpMatcher(DataGetter("spec.path"), `grace`))},
		// Started Running monitor process (nc) in container
		{"Launch-Monitor", And(MatchMessage(`garden-server.run.spawned`), RegExpMatcher(DataGetter("spec.path"), `nc`))},
		// Executor transitioning container to RUNNING
		{"Container-Is-Running", MatchMessage(`run-container.run.run-step-process.succeeded-transitioning-to-running`)},
		// Rep transitioned LRP to RUNNING in BBS
		{"Running-In-BBS", MatchMessage(`start-actual-lrp.succeeded`)},
		// Rep requesting container stop
		{"Stopping", MatchMessage(`lrp-stopper.stop.stopping`)},
		// LRP has been cancelled
		{"Stopped", MatchMessage(`run-container.run.run-step-process.step-cancelled`)},
		// Rep transitioned LRP to COMPLETED in BBS
		{"Remove-From-BBS", MatchMessage(`run-container.run.run-step-process.succeeded-transitioning-to-complete`)},
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
	plotFezzikLRPTimelinesAndHistograms(completeLRPStartTimelines, outputDir, "starting", 1)

	return nil
}

func (f *FezzikLRPs) extractInstanceGuidGroups(e Entries, processGuid string) *GroupedEntries {
	//find all instance-guid groupings
	//these might include instances for process guids *other* than the one we care about
	unfilteredByInstanceGuid := e.GroupBy(TransformingGetter(TransformationMap{
		"instance-guid":  TrimTransformation,
		"instance_guid":  TrimTransformation,
		"container-guid": TrimWithPrefixTransformation(processGuid + "-"),
		"guid":           TrimWithPrefixTransformation(processGuid + "-"),
		"container.guid": TrimWithPrefixTransformation(processGuid + "-"),
		"handle":         TrimWithPrefixTransformation(processGuid + "-"),
	}))

	//request.depot-client.allocate-containers.allocating-container allows us to correlate processguid with instanceguid
	//this fetches all such log-lines by the requested processGuid
	//then groups them by instance guid
	instances := e.Filter(And(
		MatchMessage("request.depot-client.allocate-containers.allocating-container"),
		RegExpMatcher(DataGetter("container.tags.process-guid"), processGuid),
	)).GroupBy(TransformingGetter(TransformationMap{"container.guid": TrimWithPrefixTransformation(processGuid + "-")}))

	//running-watcher.watching-for-actual-lrp-changes.sending-create is emitted soon after the actualLRP is created in the BBS
	//this is important information and is a proxy for when the ActualLRP enters the system
	createEventsByIndex := e.Filter(And(
		MatchMessage("running-watcher.watching-for-actual-lrp-changes.sending-create"),
		RegExpMatcher(DataGetter("actual-lrp.process_guid"), processGuid),
	)).GroupBy(DataGetter("actual-lrp.index"))

	//we construct the final grouping by...
	byInstanceGuid := NewGroupedEntries()

	//...iterating over all possible instance guids...
	unfilteredByInstanceGuid.EachGroup(func(key interface{}, entries Entries) error {
		_, ok := instances.Lookup(key)
		if !ok {
			//...and rejecting any that don't correlate with the process guid in question
			return nil
		}

		//we then work very hard to pick out the create event associated with the instance guid (by index)
		indexEntry, _ := entries.First(MatchMessage("request.depot-client.allocate-containers.allocating-container"))
		indexInterface, _ := DataGetter("container.tags.process-index").Get(indexEntry)
		index, _ := strconv.ParseFloat(indexInterface.(string), 64)

		createEventsForIndex, ok := createEventsByIndex.Lookup(index)
		if ok {
			entries = append(Entries{createEventsForIndex[0]}, entries...)
			sort.Sort(entries)
		}

		byInstanceGuid.AppendEntries(key, entries)
		return nil
	})

	return byInstanceGuid
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
