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

type AnalyzeConvergenceForMissingCells struct{}

func (f *AnalyzeConvergenceForMissingCells) Usage() string {
	return "analyze-convergence-for-missing-cell CONVERGER_CICERONE_LOGS SESSION"
}

func (f *AnalyzeConvergenceForMissingCells) Description() string {
	return `
Takes as input:
    - a converger log (presumed to be in Cicerone/lager format)
    - a lager session that references a single convergence loop

It is assumed that the convergence loop covers a cell-disappearance event.

Cicerone then generates timeline plots identifying how long it took convergence to happen.
`
}

func (f *AnalyzeConvergenceForMissingCells) Command(outputDir string, args ...string) error {
	if len(args) != 2 {
		return fmt.Errorf("Expected a log file and a session")
	}

	entries, err := converters.EntriesFromLagerFile(args[0])
	if err != nil {
		return err
	}

	entries = entries.Filter(MatchSession(`^` + args[1] + `\.`))

	byLRP := entries.GroupBy(DataGetter("process-guid"))

	timelineDescription := TimelineDescription{
		{"Noticed-Missing-Cell", MatchMessage(`converge-lrps.calculate-convergence.missing-cell`)},
		{"Removing-Actual-LRP", MatchMessage(`start-missing-actual.remove-actual-lrp.starting`)},
		{"Removed-Actual-LRP", MatchMessage(`start-missing-actual.remove-actual-lrp.succeeded`)},
		{"Adding-Start-Auction", MatchMessage(`start-missing-actual.adding-start-auction`)},
	}

	timelines, err := byLRP.ConstructTimelines(timelineDescription)
	if err != nil {
		return err
	}

	completeTimelines := timelines.CompleteTimelines()
	say.Println(0, say.Red("Complete Timelines: %d/%d (%.2f%%)\n",
		len(completeTimelines),
		len(timelines),
		float64(len(completeTimelines))/float64(len(timelines))*100.0))

	plotTimelinesAndHistograms(completeTimelines, outputDir, "converger-timelines")

	return nil
}

func plotTimelinesAndHistograms(timelines Timelines, outputDir string, prefix string) {
	timelines.SortByStartTime()

	histograms := viz.NewEntryPairsHistogramBoard(timelines)
	histograms.Save(3.0*float64(len(timelines.Description())), 6.0, filepath.Join(outputDir, prefix+"-histograms.png"))

	correlationBoard, _ := viz.NewCorrelationBoard(timelines)
	correlationBoard.Save(24.0, 24.0, filepath.Join(outputDir, prefix+"-correlation.png"))

	timelineBoard := &viz.Board{}
	p, _ := plot.New()
	p.Title.Text = "Timelines"
	p.Add(viz.NewTimelinesPlotter(timelines, timelines.StartsAfter().Seconds(), timelines.EndsAfter().Seconds()))
	timelineBoard.AddSubPlot(p, viz.Rect{0, 0, 1.0, 1.0})
	timelineBoard.Save(16.0, 20.0, filepath.Join(outputDir, prefix+"-timelines.png"))
}
