package viz

import (
	"time"

	"code.google.com/p/plotinum/plotter"

	. "github.com/onsi/cicerone/dsl"
)

func NewEntryPairsHistogram(pairs EntryPairs, n int) *plotter.Histogram {
	durations := pairs.Durations()
	min := durations.Min()
	max := durations.Max()

	return NewScaledEntryPairsHistogram(pairs, n, min, max)
}

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
