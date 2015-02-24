package viz

import (
	"time"

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
