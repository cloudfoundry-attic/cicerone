package viz

import (
	"fmt"
	"image/color"
	"time"

	"code.google.com/p/plotinum/plot"
	"code.google.com/p/plotinum/plotter"

	. "github.com/onsi/cicerone/dsl"
)

//NewEntryPairsHistogram plots a Historam (using n bins) of the durations in the passed in EntryPairs
//The weight (i.e. height) of each bin is simply the number of pairs in the bin.
//The minimum and maximum bins are computed from the minimum/maximum duration in the EntryPairs collection
func NewEntryPairsHistogram(pairs EntryPairs, n int) *plotter.Histogram {
	durations := pairs.Durations()
	min := durations.Min()
	max := durations.Max()

	return NewScaledEntryPairsHistogram(pairs, n, min, max)
}

//NewScaledEntyrPairsHistogram allows you to specify the minimum and maximum bounds of the Histogram.
//This is useful to compare different EntryPairs on equal footing.
func NewScaledEntryPairsHistogram(pairs EntryPairs, n int, min time.Duration, max time.Duration) *plotter.Histogram {
	durations := pairs.Durations()

	bins := []plotter.HistogramBin{}

	dt := (max - min) / time.Duration(n)
	for i := 0; i < n; i++ {
		low := min + dt*time.Duration(i)
		high := min + dt*time.Duration(i+1)
		if i == n-1 {
			high = max
		}
		bins = append(bins, plotter.HistogramBin{
			Min:    low.Seconds(),
			Max:    high.Seconds(),
			Weight: float64(durations.CountInRange(low, high)),
		})
	}

	return &plotter.Histogram{
		Bins:      bins,
		LineStyle: plotter.DefaultLineStyle,
	}
}

func NewEntryPairsHistogramBoard(timelines Timelines) *UniformBoard {
	histograms := NewUniformBoard(len(timelines.Description()), 2, 0.01)

	for i, timelinePoint := range timelines.Description() {
		entryPairs := timelines.EntryPairs(i)
		p, _ := plot.New()
		p.Title.Text = timelinePoint.Name
		p.Title.Color = OrderedColors[i]
		h := NewEntryPairsHistogram(entryPairs, 30)
		h.Color = color.RGBA{0, 0, 0, 255}
		p.Add(h)

		twentyPercent := int(float64(len(entryPairs)) * 0.2)
		h = NewEntryPairsHistogram(entryPairs[:twentyPercent], 30)
		h.Color = color.RGBA{0, 0, 255, 255}
		p.Add(h)

		h = NewEntryPairsHistogram(entryPairs[len(entryPairs)-twentyPercent:], 30)
		h.Color = color.RGBA{255, 0, 0, 255}
		p.Add(h)

		histograms.AddNextSubPlot(p)
	}

	for i, timelinePoint := range timelines.Description() {
		entryPairs := timelines.EntryPairs(i)
		p, _ := plot.New()
		p.Title.Text = timelinePoint.Name
		p.Title.Color = OrderedColors[i]
		h := NewScaledEntryPairsHistogram(entryPairs, 30, 0, timelines.EndsAfter())
		h.Color = color.RGBA{0, 0, 0, 255}
		p.Add(h)
		histograms.AddNextSubPlot(p)
	}

	return histograms
}

func NewGroupedTimelineEntryPairsHistogramBoard(group *GroupedTimelines) *UniformBoard {
	histograms := NewUniformBoard(len(group.Description()), 1, 0.01)

	for i, timelinePoint := range group.Description() {
		p, _ := plot.New()
		p.Title.Text = timelinePoint.Name
		p.Title.Color = OrderedColors[i]
		for j, timelines := range group.Timelines {
			entryPairs := timelines.EntryPairs(i)
			h := NewEntryPairsHistogram(entryPairs, 30)
			h.Color = OrderedColors[j]
			p.Add(h)
			if i == 0 {
				p.Legend.Add(fmt.Sprintf("%s", group.Keys[j]), &LineThumbnailer{h.LineStyle})
			}
		}
		histograms.AddNextSubPlot(p)
	}

	return histograms
}
