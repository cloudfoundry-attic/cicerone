package viz

import (
	"image/color"

	"github.com/gonum/plot"
	"github.com/gonum/plot/vg/draw"

	. "github.com/cloudfoundry-incubator/cicerone/dsl"
)

//TimelinesPlotter plots a stack of Timelines in the specified time range
type TimelinesPlotter struct {
	Timelines  Timelines
	MinSeconds float64
	MaxSeconds float64
	Padding    float64
}

func NewTimelinesPlotter(timelines Timelines, minSeconds float64, maxSeconds float64) *TimelinesPlotter {
	return &TimelinesPlotter{
		Timelines:  timelines,
		MinSeconds: minSeconds,
		MaxSeconds: maxSeconds,
		Padding:    0.1,
	}
}

type timelineEvent struct {
	X     float64
	Color color.Color
}

func (t *TimelinesPlotter) Plot(da draw.Canvas, p *plot.Plot) {
	trX, trY := p.Transforms(&da)
	y := t.Padding
	for _, timeline := range t.Timelines {
		events := []timelineEvent{}
		for i, entry := range timeline.Entries {
			if entry.IsZero() {
				continue
			}
			events = append(events, timelineEvent{
				X:     entry.Timestamp.Sub(timeline.ZeroEntry.Timestamp).Seconds(),
				Color: OrderedColors[i],
			})
		}
		bottom := trY(y)
		top := trY(y + 1.0)
		for i := 1; i < len(events); i++ {
			left := trX(events[i-1].X)
			right := trX(events[i].X)
			da.SetColor(events[i].Color)
			da.Fill(pathRectangle(top, right, bottom, left))
		}
		y += 1.0 + t.Padding
	}

	description := t.Timelines.Description()
	dx := (t.MaxSeconds - t.MinSeconds) / float64(len(description)+1)

	x := t.MinSeconds + dx
	for i := 1; i < len(description); i++ {
		da.SetColor(OrderedColors[i])
		da.Fill(pathRectangle(trY(y+t.legendHeight()*0.5), trX(x+dx), trY(y+t.legendHeight()*0.1), trX(x)))
		x += dx
	}

	x = t.MinSeconds + dx

	for i := 0; i < len(description); i++ {
		textStyle := draw.TextStyle{
			Color: OrderedColors[i],
			Font:  defaultFont,
		}

		da.FillText(textStyle, trX(x), trY(y+t.legendHeight()*0.6), -0.5, 0, description[i].Name)
		x += dx
	}
}

func (t *TimelinesPlotter) legendHeight() float64 {
	return float64(len(t.Timelines)) * 0.05
}

func (t *TimelinesPlotter) DataRange() (xmin, xmax, ymin, ymax float64) {
	ymin = 0.0
	ymax = float64(len(t.Timelines)) + t.Padding*float64(len(t.Timelines)+1) + t.legendHeight()
	xmin = t.MinSeconds
	xmax = t.MaxSeconds

	return
}
