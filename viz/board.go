package viz

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"code.google.com/p/plotinum/plot"
	"code.google.com/p/plotinum/vg"
	"code.google.com/p/plotinum/vg/vgeps"
	"code.google.com/p/plotinum/vg/vgimg"
	"code.google.com/p/plotinum/vg/vgpdf"
	"code.google.com/p/plotinum/vg/vgsvg"
)

type SubPlot struct {
	Plot *plot.Plot
	Rect Rect
}

type Rect struct {
	X, Y, Width, Height float64
}

func (sp SubPlot) ScaledRect(width, height float64) plot.Rect {
	return plot.Rect{
		Min:  plot.Point{vg.Length(sp.Rect.X * width), vg.Length(sp.Rect.Y * height)},
		Size: plot.Point{vg.Length(sp.Rect.Width * width), vg.Length(sp.Rect.Height * height)},
	}
}

func NewUniformBoard(horizontal int, vertical int, padding float64) *UniformBoard {
	return &UniformBoard{
		Horizontal: horizontal,
		Vertical:   vertical,
		Padding:    padding,
	}
}

type UniformBoard struct {
	Board
	Horizontal int
	Vertical   int
	Padding    float64
	counter    int
}

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

func (b *UniformBoard) AddNextSubPlot(plot *plot.Plot) {
	i := b.counter % b.Horizontal
	j := b.counter / b.Horizontal
	b.AddSubPlotAt(plot, i, j)
	b.counter += 1
}

type Board struct {
	SubPlots []*SubPlot
}

func (b *Board) AddSubPlot(plot *plot.Plot, rect Rect) {
	b.SubPlots = append(b.SubPlots, &SubPlot{
		Plot: plot,
		Rect: rect,
	})
}

func (b *Board) Save(width, height float64, file string) (err error) {
	w, h := vg.Inches(width), vg.Inches(height)
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
		drawArea := plot.DrawArea{
			Canvas: c,
			Rect:   subplot.ScaledRect(float64(w), float64(h)),
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
