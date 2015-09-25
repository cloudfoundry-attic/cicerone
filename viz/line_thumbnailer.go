package viz

import "github.com/gonum/plot/vg/draw"

type LineThumbnailer struct {
	draw.LineStyle
}

func (l *LineThumbnailer) Thumbnail(c *draw.Canvas) {
	y := c.Center().Y
	c.StrokeLine2(l.LineStyle, c.Min.X, y, c.Max.X, y)
}
