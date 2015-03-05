package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/cloudfoundry/gunk/workpool"

	"github.com/onsi/cicerone/converters"
	. "github.com/onsi/cicerone/dsl"
	"github.com/onsi/say"
)

type disappearingCellEvent struct {
	Designation              string
	ConvergerActionTimestamp time.Time
}

func (d disappearingCellEvent) String() string {
	return fmt.Sprintf("%s - %s", d.Designation, d.ConvergerActionTimestamp.Format("15:04:05"))
}

//commented out events are actually no-ops

var disappearingCellEvents = []disappearingCellEvent{
	{"A", time.Unix(1425352333, 0)},
	{"B", time.Unix(1425352402, 0)},
	// {"C", time.Unix(1425352415, 0)},
	// {"D", time.Unix(1425355185, 0)},
	{"E", time.Unix(1425364808, 0)},
	{"F", time.Unix(1425367137, 0)},
	// {"G", time.Unix(1425367149, 0)},
	{"H", time.Unix(1425373269, 0)},
	// {"I", time.Unix(1425373286, 0)},
	{"J", time.Unix(1425378776, 0)},
	// {"K", time.Unix(1425378787, 0)},
	// {"L", time.Unix(1425379631, 0)},
	{"M", time.Unix(1425398902, 0)},
}

var disappearingCellPath = "/Users/onsi/workspace/performance/10-cells/cf-pushes/optimization-2-no-disk-quota/disappearing-cells"

type AnalyzeDisappearingCells struct {
	events   []disappearingCellEvent
	entries  map[string]Entries
	workPool *workpool.WorkPool
}

func (a *AnalyzeDisappearingCells) Usage() string {
	return "analyze-disappearing-cells"
}

func (a *AnalyzeDisappearingCells) Description() string {
	return `analyze-disappearing-cells EVENTS TO ANALYZE

    analyze-disappearing-cells A B C
or
    analyze-disappearing-cells
to analyze everything
`
}

func (a *AnalyzeDisappearingCells) Command(outputDir string, args ...string) error {
	a.workPool = workpool.NewWorkPool(runtime.NumCPU())
	a.pickEvents(args)
	say.Println(0, say.Green("Loading Entries"))
	a.loadEntries()

	say.Println(0, say.Green("Convergence Durations and Missing Cells Per Event"))
	a.forEach(a.findConvergenceDurations)

	say.Println(0, say.Green("Auction related concerns"))
	a.forEach(a.exploreAuctions)

	return nil
}

func (a *AnalyzeDisappearingCells) forEach(f func(event disappearingCellEvent, entries Entries) error) {
	for _, event := range a.events {
		err := f(event, a.entries[event.Designation])
		if err != nil {
			fmt.Println("bailing", err.Error())
			return
		}
	}
}

func (a *AnalyzeDisappearingCells) findConvergenceDurations(event disappearingCellEvent, entries Entries) error {
	logsInWindow := entries.Filter(MatchAfter(event.ConvergerActionTimestamp.Add(-time.Millisecond)))

	startingLRPConvergence, _ := logsInWindow.First(MatchMessage("cell-disappeared.converge-lrps.starting-convergence"))
	endingLRPConvergence, _ := logsInWindow.Filter(MatchAfter(startingLRPConvergence.Timestamp)).First(MatchMessage("cell-disappeared.converge-lrps.finished-convergence"))
	logsInBetween := logsInWindow.Filter(And(MatchAfter(startingLRPConvergence.Timestamp), MatchBefore(endingLRPConvergence.Timestamp)))
	dt := endingLRPConvergence.Timestamp.Sub(startingLRPConvergence.Timestamp)

	groups := logsInBetween.Filter(
		MatchMessage("watching-for-actual-lrp-changes.sending-delete"),
	).GroupBy(
		DataGetter("actual-lrp.cell_id"),
	)

	if dt < 0 {
		say.Println(1, say.Red("%s:", event))
	} else if dt < 3*time.Second {
		say.Println(1, say.Yellow("%s:", event))
	} else {
		say.Println(1, say.Green("%s:", event))
	}
	say.Println(2, "Convergence took %s and spans %s converger log-lines and %s total log-lines",
		say.Yellow("%s", dt),
		say.Yellow("%d", len(logsInBetween.Filter(MatchSource("converger")))),
		say.Yellow("%d", len(logsInBetween)),
	)

	if len(groups.Keys) == 0 {
		say.Println(2, say.Yellow("Not seeing any ActualLRPs get deleted"))
	} else {
		groups.EachGroup(func(key interface{}, entries Entries) error {
			say.Println(3, "%s: %d LRPs", key, len(entries))
			return nil
		})
	}

	return nil
}

func (a *AnalyzeDisappearingCells) exploreAuctions(event disappearingCellEvent, entries Entries) error {
	logsInWindow := entries.Filter(MatchAfter(event.ConvergerActionTimestamp.Add(-time.Millisecond)))
	startingLRPConvergence, _ := logsInWindow.First(MatchMessage("cell-disappeared.converge-lrps.starting-convergence"))
	logsInWindow = logsInWindow.Filter(MatchAfter(startingLRPConvergence.Timestamp))

	bySession := logsInWindow.Filter(MatchSource("auctioneer")).Filter(MatchMessage(`auction\.`)).GroupBy(GetSession)

	say.Println(1, "%s", event)

	if len(bySession.Keys) == 0 {
		say.Println(2, say.Red("No auction found!"))
		return nil
	}

	firstAuction := bySession.Entries[0]

	firstEntry := firstAuction[0]
	lastEntry := firstAuction[len(firstAuction)-1]
	dt := lastEntry.Timestamp.Sub(firstEntry.Timestamp)

	zoneStateEntry, _ := firstAuction.First(MatchMessage("auction.fetched-zone-state"))
	numStatesFetched, _ := DataGetter("cell-state-count").Get(zoneStateEntry)

	scheduledEntry, _ := firstAuction.First(MatchMessage("auction.scheduled"))
	successfulAuctions, _ := DataGetter("successful-lrp-start-auctions").Get(scheduledEntry)
	failedAuctions, _ := DataGetter("failed-lrp-start-auctions").Get(scheduledEntry)

	say.Println(2, "Ran auction %s after cell-missing event - fetched state from %s cells: %s succeeded, %s failed in %s", firstEntry.Timestamp.Sub(event.ConvergerActionTimestamp), say.Green("%.0f", numStatesFetched), say.Green("%.0f", successfulAuctions), say.Red("%.0f", failedAuctions), dt)

	logsDuringAuction := logsInWindow.Filter(MatchBefore(scheduledEntry.Timestamp))

	groupedByVMs := logsDuringAuction.Filter(MatchMessage("auction-delegate")).GroupBy(GetVM)

	groupedByVMs.EachGroup(func(key interface{}, entries Entries) error {
		providing, _ := entries.First(MatchMessage("auction-state.providing"))
		provided, _ := entries.First(MatchMessage("provided"))
		allocating, _ := entries.First(MatchMessage("lrp-allocate-instances.allocating"))
		allocated, _ := entries.First(MatchMessage("lrp-allocate-instances.allocated"))

		timeToProvide := provided.Timestamp.Sub(providing.Timestamp)
		placeholder, _ := DataGetter("available-resources.Containers").Get(provided)
		availableContainers := placeholder.(float64)
		placeholder, _ = DataGetter("available-resources.MemoryMB").Get(provided)
		availableMemory := placeholder.(float64)

		say.Println(2, "%s: Provided data in %s.  Has %.0f containers, %.0f memory available", key, timeToProvide, availableContainers, availableMemory)

		if allocating.IsZero() {
			say.Println(3, say.Red("nothing was allocated to this cell"))
		} else {
			placeholder, _ = DataGetter("lrp-starts").Get(allocating)
			numLRPs := placeholder.(float64)
			say.Println(3, "Allocated %.0f LRPs (took %s to allocate) => end up with %.0f containers available", numLRPs, allocated.Timestamp.Sub(allocating.Timestamp), availableContainers-numLRPs)
		}

		return nil
	})

	return nil
}
func (a *AnalyzeDisappearingCells) pickEvents(args []string) {
	if len(args) == 0 {
		a.events = disappearingCellEvents
	} else {
		a.events = []disappearingCellEvent{}
		for _, arg := range args {
			for _, event := range disappearingCellEvents {
				if event.Designation == arg {
					a.events = append(a.events, event)
				}
			}
		}
	}
}

func (a *AnalyzeDisappearingCells) loadEntries() {
	a.entries = map[string]Entries{}

	lock := &sync.Mutex{}

	wg := &sync.WaitGroup{}
	wg.Add(len(a.events))

	for _, event := range a.events {
		event := event
		a.workPool.Submit(func() {
			defer wg.Done()
			entries, err := converters.EntriesFromLagerFile(filepath.Join(disappearingCellPath, event.Designation+".log"))
			if err != nil {
				return
			}
			lock.Lock()
			a.entries[event.Designation] = entries
			lock.Unlock()
			say.Println(1, "Loaded %s", event)
		})
	}

	wg.Wait()
}

type SlurpDisappearingCells struct{}

func (f *SlurpDisappearingCells) Usage() string {
	return "disappearing-cells-slurp"
}

func (f *SlurpDisappearingCells) Description() string {
	return `One off: slurp disappearing cells`
}

func (f *SlurpDisappearingCells) Command(outputDir string, args ...string) error {
	wp := workpool.NewWorkPool(8)
	wg := &sync.WaitGroup{}
	wg.Add(len(disappearingCellEvents))

	for _, event := range disappearingCellEvents {
		event := event
		wp.Submit(func() {
			defer wg.Done()
			fmt.Println("Processing ", event.Designation)
			entries, err := converters.EntriesFromBOSHTree(
				"/Users/onsi/workspace/performance/10-cells/cf-pushes/optimization-2-no-disk-quota/bosh-logs",
				event.ConvergerActionTimestamp.Add(-10*time.Second),
				event.ConvergerActionTimestamp.Add(120*time.Second),
			)
			if err != nil {
				say.Println(0, say.Red(err.Error()))
				return
			}
			outputFile, err := os.Create("/Users/onsi/workspace/performance/10-cells/cf-pushes/optimization-2-no-disk-quota/disappearing-cells/" + event.Designation + ".log")
			if err != nil {
				say.Println(0, say.Red(err.Error()))
				return
			}

			entries.WriteLagerFormatTo(outputFile)
			fmt.Println("Finished ", event.Designation)
		})
	}

	wg.Wait()

	return nil
}
