package viz

import "code.google.com/p/plotinum/plot"

type LineThumbnailer struct {
	plot.LineStyle
}

func (l *LineThumbnailer) Thumbnail(da *plot.DrawArea) {
	y := da.Center().Y
	da.StrokeLine2(l.LineStyle, da.Min.X, y, da.Max().X, y)
}
