package commands

import (
	"image/color"

	. "github.com/onsi/cicerone/dsl"

	"code.google.com/p/plotinum/plot"
	"github.com/onsi/cicerone/viz"
)

func plotTimelinesHistogramsBoard(timelines Timelines, filename string) error {
	histograms := viz.NewUniformBoard(len(timelines.Description()), 2, 0.01)

	for i, timelinePoint := range timelines.Description() {
		entryPairs := timelines.EntryPairs(i)
		p, _ := plot.New()
		p.Title.Text = timelinePoint.Name
		h := viz.NewEntryPairsHistogram(entryPairs, 30)
		h.Color = color.RGBA{0, 0, 0, 255}
		p.Add(h)

		twentyPercent := int(float64(len(entryPairs)) * 0.2)
		h = viz.NewEntryPairsHistogram(entryPairs[:twentyPercent], 30)
		h.Color = color.RGBA{0, 0, 255, 255}
		p.Add(h)

		h = viz.NewEntryPairsHistogram(entryPairs[len(entryPairs)-twentyPercent:], 30)
		h.Color = color.RGBA{255, 0, 0, 255}
		p.Add(h)

		histograms.AddNextSubPlot(p)
	}

	for i, timelinePoint := range timelines.Description() {
		entryPairs := timelines.EntryPairs(i)
		p, _ := plot.New()
		p.Title.Text = timelinePoint.Name
		h := viz.NewScaledEntryPairsHistogram(entryPairs, 30, 0, timelines.EndsAfter())
		h.Color = color.RGBA{0, 0, 0, 255}
		p.Add(h)
		histograms.AddNextSubPlot(p)
	}

	return histograms.Save(3.0*float64(len(timelines.Description())), 6.0, filename)
}
