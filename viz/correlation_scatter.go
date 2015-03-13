package viz

import (
	"fmt"
	"image/color"

	"code.google.com/p/plotinum/plot"
	"code.google.com/p/plotinum/plotter"
	. "github.com/onsi/cicerone/dsl"
)

func newCorrelationScatter(xDurations Durations, yDurations Durations, c color.Color) (*plotter.Scatter, error) {
	xys := plotter.XYs{}
	for i := range xDurations {
		xys = append(xys, struct{ X, Y float64 }{xDurations[i].Seconds(), yDurations[i].Seconds()})
	}
	s, err := plotter.NewScatter(xys)
	if err != nil {
		return nil, err
	}
	s.GlyphStyle = plot.GlyphStyle{
		Color:  c,
		Radius: 2,
		Shape:  plot.CircleGlyph{},
	}
	return s, nil
}

//Constructs and returns a correlation board between all possible entry pairs
func NewCorrelationBoard(timelines Timelines) (*UniformBoard, error) {
	//timelines must be complete!
	timelines = timelines.CompleteTimelines()

	size := len(timelines.Description())

	board := NewUniformBoard(size, size, 0.01)
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			p, _ := plot.New()

			iPairs, jPairs := timelines.MatchedEntryPairs(i, j)

			xDurations := iPairs.Durations()
			yDurations := jPairs.Durations()

			twentyPercent := int(float64(len(xDurations)) * 0.2)
			eightyPercent := len(xDurations) - twentyPercent

			s, err := newCorrelationScatter(xDurations[:twentyPercent], yDurations[:twentyPercent], color.RGBA{0, 0, 255, 255})
			if err != nil {
				return nil, err
			}
			p.Add(s)

			s, err = newCorrelationScatter(xDurations[twentyPercent:eightyPercent], yDurations[twentyPercent:eightyPercent], color.RGBA{0, 0, 0, 255})
			if err != nil {
				return nil, err
			}
			p.Add(s)

			s, err = newCorrelationScatter(xDurations[eightyPercent:], yDurations[eightyPercent:], color.RGBA{255, 0, 0, 255})
			if err != nil {
				return nil, err
			}
			p.Add(s)

			p.X.Label.Text = timelines.Description()[i].Name
			p.X.Label.Color = OrderedColors[i]
			p.Y.Label.Text = timelines.Description()[j].Name
			p.Y.Label.Color = OrderedColors[j]
			board.AddSubPlotAt(p, i, j)
		}
	}
	return board, nil
}

//Constructs and returns a correlation board between all possible entry pairs
func NewGroupedCorrelationBoard(group *GroupedTimelines) (*UniformBoard, error) {
	size := len(group.Description())

	board := NewUniformBoard(size, size, 0.01)
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			p, _ := plot.New()

			for k, timelines := range group.Timelines {
				iPairs, jPairs := timelines.MatchedEntryPairs(i, j)

				xDurations := iPairs.Durations()
				yDurations := jPairs.Durations()

				s, err := newCorrelationScatter(xDurations, yDurations, OrderedColors[k])
				if err != nil {
					return nil, err
				}
				p.Add(s)
				if i == 0 && j == 0 {
					p.Legend.Add(fmt.Sprintf("%s", group.Keys[k]), s)
				}
			}

			p.X.Label.Text = group.Description()[i].Name
			p.X.Label.Color = OrderedColors[i]
			p.Y.Label.Text = group.Description()[j].Name
			p.Y.Label.Color = OrderedColors[j]
			board.AddSubPlotAt(p, i, j)
		}
	}
	return board, nil
}
