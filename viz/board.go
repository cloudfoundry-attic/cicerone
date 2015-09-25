package viz

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/gonum/plot"
	"github.com/gonum/plot/vg"
	"github.com/gonum/plot/vg/draw"
	"github.com/gonum/plot/vg/vgeps"
	"github.com/gonum/plot/vg/vgimg"
	"github.com/gonum/plot/vg/vgpdf"
	"github.com/gonum/plot/vg/vgsvg"
)

//Rect represents a rectangle
type Rect struct {
	X, Y, Width, Height float64
}

//SubPlot represents a plotinum plot placed within a particular unit rectangle
type SubPlot struct {
	Plot *plot.Plot
	Rect Rect
}

//ScaledRect returns a plotinum Rectangle, appropriately scaled to match the SubPlot's unit rectangle
func (sp SubPlot) ScaledRect(width, height float64) draw.Rectangle {
	return draw.Rectangle{
		Min: draw.Point{vg.Length(sp.Rect.X * width), vg.Length(sp.Rect.Y * height)},
		Max: draw.Point{
			vg.Length(sp.Rect.X*width) + vg.Length(sp.Rect.Width*width),
			vg.Length(sp.Rect.Y*height) + vg.Length(sp.Rect.Height*height),
		},
	}
}

//NewUniformBoard returns a UniformBoard ready to accept horizontalxvertical subplots
func NewUniformBoard(horizontal int, vertical int, padding float64) *UniformBoard {
	return &UniformBoard{
		Horizontal: horizontal,
		Vertical:   vertical,
		Padding:    padding,
	}
}

//UniformBoard is a Board that can automatically place a grid of subplots
type UniformBoard struct {
	Board
	Horizontal int
	Vertical   int
	Padding    float64
	counter    int
}

//AddSubPlotAt adds the passed in plotinum subplot to the UniformBoard.  The subplot is ith subplot across, and the jth plot up.
func (b *UniformBoard) AddSubPlotAt(plot *plot.Plot, i int, j int) error {
	if i >= b.Horizontal {
		return fmt.Errorf("i:%d >= Horizontal:%d", i, b.Horizontal)
	}
	if j >= b.Vertical {
		return fmt.Errorf("j:%d >= Vertical:%d", j, b.Vertical)
	}

	width := (1.0 - float64(b.Horizontal+1)*b.Padding) / float64(b.Horizontal)
	height := (1.0 - float64(b.Vertical+1)*b.Padding) / float64(b.Vertical)
	x := (float64(i)+1)*b.Padding + float64(i)*width
	y := (float64(j)+1)*b.Padding + float64(j)*height

	r := Rect{
		X:      x,
		Y:      y,
		Width:  width,
		Height: height,
	}

	b.AddSubPlot(plot, r)

	return nil
}

//AddNextSubPlot populates the next available subplot in the UniformBoard with the passed-in plot
func (b *UniformBoard) AddNextSubPlot(plot *plot.Plot) {
	i := b.counter % b.Horizontal
	j := b.counter / b.Horizontal
	b.AddSubPlotAt(plot, i, j)
	b.counter += 1
}

//Board represents a collection of SubPlots
type Board struct {
	SubPlots []*SubPlot
}

//AddSubPlot adds a plotinum plot as a SubPlot on the board.
//
//Rect represents the rectangle the SubPlot will occupy -- the entire board has width 1.0 and height 1.0 and the rectangle (called a unit rectangle) should consume some fraction of this space.
func (b *Board) AddSubPlot(plot *plot.Plot, rect Rect) {
	b.SubPlots = append(b.SubPlots, &SubPlot{
		Plot: plot,
		Rect: rect,
	})
}

//Save saves the board (i.e. all subplots, appropriately laid out) to the specified filename.
//It basically rips off the implementation of Save in plotinum to support various file formats.
func (b *Board) Save(width, height float64, file string) (err error) {
	w, h := vg.Inch*vg.Length(width), vg.Inch*vg.Length(height)
	var c interface {
		vg.Canvas
		Size() (w, h vg.Length)
		io.WriterTo
	}
	switch ext := strings.ToLower(filepath.Ext(file)); ext {

	case ".eps":
		c = vgeps.NewTitle(w, h, file)

	case ".jpg", ".jpeg":
		c = vgimg.JpegCanvas{Canvas: vgimg.New(w, h)}

	case ".pdf":
		c = vgpdf.New(w, h)

	case ".png":
		c = vgimg.PngCanvas{Canvas: vgimg.New(w, h)}

	case ".svg":
		c = vgsvg.New(w, h)

	case ".tiff":
		c = vgimg.TiffCanvas{Canvas: vgimg.New(w, h)}

	default:
		return fmt.Errorf("Unsupported file extension: %s", ext)
	}

	for _, subplot := range b.SubPlots {
		w, h := c.Size()
		drawArea := draw.Canvas{
			Canvas:    c,
			Rectangle: subplot.ScaledRect(float64(w), float64(h)),
		}

		subplot.Plot.Draw(drawArea)
	}

	f, err := os.Create(file)
	if err != nil {
		return err
	}
	if _, err = c.WriteTo(f); err != nil {
		return err
	}
	return f.Close()
}
