package viz

import (
	"fmt"
	"image/color"

	"code.google.com/p/plotinum/plot"
	"code.google.com/p/plotinum/vg"

	. "github.com/onsi/sommelier/dsl"
)

var defaultFont vg.Font

func init() {
	var err error
	plot.DefaultFont = "Helvetica"
	defaultFont, err = vg.MakeFont("Helvetica", 10)
	if err != nil {
		fmt.Println(err.Error())
	}
}

var TimelineColors = []color.RGBA{
	{0, 0, 0, 255},
	{255, 0, 0, 255},
	{0, 255, 0, 255},
	{0, 0, 255, 255},
	{125, 0, 0, 255},
	{0, 125, 0, 255},
	{0, 0, 125, 255},
	{125, 125, 0, 255},
	{125, 0, 125, 255},
	{0, 125, 125, 255},
	{125, 125, 125, 255},
	{200, 200, 200, 255},
	{255, 125, 0, 255},
	{0, 125, 255, 255},
}

type TimelinesPlotter struct {
	Timelines  Timelines
	MinSeconds float64
	MaxSeconds float64
	Padding    float64
}

func PathRectangle(top vg.Length, right vg.Length, bottom vg.Length, left vg.Length) vg.Path {
	p := vg.Path{}
	p.Move(left, top)
	p.Line(right, top)
	p.Line(right, bottom)
	p.Line(left, bottom)
	p.Close()
	return p
}

func NewTimelinesPlotter(timelines Timelines, minSeconds float64, maxSeconds float64) *TimelinesPlotter {
	return &TimelinesPlotter{
		Timelines:  timelines,
		MinSeconds: minSeconds,
		MaxSeconds: maxSeconds,
		Padding:    0.1,
	}
}

type TimelineEvent struct {
	X     float64
	Color color.Color
}

func (t *TimelinesPlotter) Plot(da plot.DrawArea, p *plot.Plot) {
	trX, trY := p.Transforms(&da)
	y := t.Padding
	for _, timeline := range t.Timelines {
		events := []TimelineEvent{}
		for i, entry := range timeline.Entries {
			if entry.IsZero() {
				continue
			}
			events = append(events, TimelineEvent{
				X:     entry.Timestamp.Sub(timeline.ZeroEntry.Timestamp).Seconds(),
				Color: TimelineColors[i],
			})
		}
		bottom := trY(y)
		top := trY(y + 1.0)
		for i := 1; i < len(events); i++ {
			left := trX(events[i-1].X)
			right := trX(events[i].X)
			da.SetColor(events[i].Color)
			da.Fill(PathRectangle(top, right, bottom, left))
		}
		y += 1.0 + t.Padding
	}

	description := t.Timelines.Description()
	dx := (t.MaxSeconds - t.MinSeconds) / float64(len(description)+1)

	x := t.MinSeconds + dx
	for i := 1; i < len(description); i++ {
		da.SetColor(TimelineColors[i])
		da.Fill(PathRectangle(trY(y+1), trX(x+dx), trY(y), trX(x)))
		x += dx
	}

	textStyle := plot.TextStyle{
		Color: color.Black,
		Font:  defaultFont,
	}

	x = t.MinSeconds + dx

	for i := 0; i < len(description); i++ {
		da.FillText(textStyle, trX(x), trY(y+1.1), -0.5, 0, description[i].Name)
		x += dx
	}
}

func (t *TimelinesPlotter) DataRange() (xmin, xmax, ymin, ymax float64) {
	ymin = 0.0
	ymax = float64(len(t.Timelines)) + t.Padding*float64(len(t.Timelines)+1) + 2
	xmin = t.MinSeconds
	xmax = t.MaxSeconds

	return
}
