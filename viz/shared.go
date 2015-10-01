package viz

import (
	"fmt"
	"image/color"

	"github.com/gonum/plot"
	"github.com/gonum/plot/vg"
)

var defaultFont vg.Font

func init() {
	var err error
	plot.DefaultFont = "Helvetica"
	defaultFont, err = vg.MakeFont("Helvetica", 6)
	if err != nil {
		fmt.Println(err.Error())
	}
}

var OrderedColors = []color.RGBA{
	{0, 0, 0, 255},
	{255, 0, 0, 255},
	{0, 200, 0, 255},
	{0, 0, 255, 255},
	{125, 0, 0, 255},
	{0, 125, 0, 255},
	{0, 0, 125, 255},
	{125, 125, 0, 255},
	{125, 0, 125, 255},
	{0, 125, 125, 255},
	{125, 125, 125, 255},
	{255, 125, 0, 255},
	{0, 125, 255, 255},
	{0, 0, 0, 255},
	{200, 200, 200, 255},
	{255, 0, 0, 255},
	{0, 200, 0, 255},
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

func pathRectangle(top vg.Length, right vg.Length, bottom vg.Length, left vg.Length) vg.Path {
	p := vg.Path{}
	p.Move(left, top)
	p.Line(right, top)
	p.Line(right, bottom)
	p.Line(left, bottom)
	p.Close()
	return p
}
