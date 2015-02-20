package viz

import (
	"image/color"

	. "github.com/onsi/cicerone/dsl"
)

type EntryPairBandPlotter struct {
	EntryPairs  EntryPairs
	FirstColor  color.Color
	SecondColor color.Color
}

func NewEntryPairBandPlotter(pairs EntryPairs) *EntryPairBandPlotter {
	return &EntryPairBandPlotter{
		EntryPairs:  pairs,
		FirstColor:  color.RGBA{0, 0, 255, 255},
		SecondColor: color.RGBA{255, 0, 0, 255},
	}
}

// type TimelineEvent struct {
// 	X     float64
// 	Color color.Color
// }

// func (t *TimelinesPlotter) Plot(da plot.DrawArea, p *plot.Plot) {
// 	trX, trY := p.Transforms(&da)
// 	y := t.Padding
// 	for _, timeline := range t.Timelines {
// 		events := []TimelineEvent{}
// 		for i, entry := range timeline.Entries {
// 			if entry.IsZero() {
// 				continue
// 			}
// 			events = append(events, TimelineEvent{
// 				X:     entry.Timestamp.Sub(timeline.ZeroEntry.Timestamp).Seconds(),
// 				Color: TimelineColors[i],
// 			})
// 		}
// 		bottom := trY(y)
// 		top := trY(y + 1.0)
// 		for i := 1; i < len(events); i++ {
// 			left := trX(events[i-1].X)
// 			right := trX(events[i].X)
// 			da.SetColor(events[i].Color)
// 			da.Fill(PathRectangle(top, right, bottom, left))
// 		}
// 		y += 1.0 + t.Padding
// 	}

// 	description := t.Timelines.Description()
// 	dx := (t.MaxSeconds - t.MinSeconds) / float64(len(description)+1)

// 	x := t.MinSeconds + dx
// 	for i := 1; i < len(description); i++ {
// 		da.SetColor(TimelineColors[i])
// 		da.Fill(PathRectangle(trY(y+t.legendHeight()*0.5), trX(x+dx), trY(y+t.legendHeight()*0.1), trX(x)))
// 		x += dx
// 	}

// 	textStyle := plot.TextStyle{
// 		Color: color.Black,
// 		Font:  defaultFont,
// 	}

// 	x = t.MinSeconds + dx

// 	for i := 0; i < len(description); i++ {
// 		da.FillText(textStyle, trX(x), trY(y+t.legendHeight()*0.6), -0.5, 0, description[i].Name)
// 		x += dx
// 	}
// }

// func (t *TimelinesPlotter) legendHeight() float64 {
// 	return float64(len(t.Timelines)) * 0.05
// }

// func (t *TimelinesPlotter) DataRange() (xmin, xmax, ymin, ymax float64) {
// 	ymin = 0.0
// 	ymax = float64(len(t.Timelines)) + t.Padding*float64(len(t.Timelines)+1) + t.legendHeight()
// 	xmin = t.MinSeconds
// 	xmax = t.MaxSeconds

// 	return
// }
