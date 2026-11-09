package verification

import (
	"fmt"
	"math"
)

// These are actual browser rectangles, not inferred CSS declarations. Admission
// recomputes clipping so a chart heading alone cannot stand in for visible data.
type entryRect390 struct{ AbsoluteDocumentTop, X, Y, Top, Left, Right, Bottom, Width, Height float64 }
type entryClip390 struct{ Left, Top, Right, Bottom float64 }
type entryIntersection390 struct {
	Width, Height float64
	CenterInside  bool
}
type entryPaint390 struct {
	Tag           string
	Visible       bool
	Rect          *entryRect390
	Clip          *entryClip390
	ClipAncestors []entryScroll390
	Intersection  *entryIntersection390
}
type entryScroll390 struct {
	Tag, ID                               string
	ScrollTop, ScrollHeight, ClientHeight float64
	ScrollLeft, ScrollWidth, ClientWidth  float64
	OverflowX, OverflowY                  string
	Rect                                  *entryRect390
}
type entryGeometry390 struct {
	Viewport                  struct{ Width, Height float64 }
	Document, Main, Workspace *entryScroll390
	Root                      struct {
		Selector string
		Visible  bool
		Rect     *entryRect390
	}
	Clip                         entryClip390
	Ancestors                    []entryScroll390
	Primary, FirstItem           entryPaint390
	NumericPlot, UnavailablePlot *entryRect390
	Marks                        []struct {
		Key, State, FromNS, ToNS string
		Value                    *string
		Rect                     *entryRect390
		Clip                     *entryClip390
		ClipAncestors            []entryScroll390
		Intersection             *entryIntersection390
		Visible                  bool
	}
	Timeline *entryTimeline390
	Summary  struct {
		entryPaint390
		Text, FromNS, ToNS string
	}
	Status           struct{ Text string }
	ScrollX, ScrollY float64
}
type entryPoint390 struct {
	Key, FromNS, ToNS, State string
	Value                    *string
}
type entryTimeline390 struct {
	FromNS, ToNS, Kind, Unit string
	Total, Omitted           int
	Points                   []entryPoint390
}

func axisRectValid390(r *entryRect390) bool {
	if r == nil {
		return false
	}
	for _, value := range []float64{r.X, r.Y, r.Top, r.Left, r.Right, r.Bottom, r.Width, r.Height} {
		if math.IsInf(value, 0) || math.IsNaN(value) {
			return false
		}
	}
	return r.Width > 0 && r.Height >= 0 && math.Abs(r.Left-r.X) < 0.1 && math.Abs(r.Top-r.Y) < 0.1 && math.Abs(r.Right-r.Left-r.Width) < 0.1 && math.Abs(r.Bottom-r.Top-r.Height) < 0.1
}
func rectValid390(r *entryRect390) bool { return axisRectValid390(r) && r.Height > 0 }
func clipped390(r *entryRect390, c entryClip390) entryIntersection390 {
	return entryIntersection390{Width: max(0, min(r.Right, c.Right)-max(r.Left, c.Left)), Height: max(0, min(r.Bottom, c.Bottom)-max(r.Top, c.Top)), CenterInside: r.Left+r.Width/2 >= c.Left && r.Left+r.Width/2 <= c.Right && r.Top+r.Height/2 >= c.Top && r.Top+r.Height/2 <= c.Bottom}
}
func fullyInside390(r *entryRect390, c entryClip390) bool {
	return rectValid390(r) && r.Left >= c.Left-1 && r.Right <= c.Right+1 && r.Top >= c.Top-1 && r.Bottom <= c.Bottom+1
}
func paintClip390(width, height int, ancestors []entryScroll390) (entryClip390, error) {
	c := entryClip390{Right: float64(width), Bottom: float64(height)}
	for _, a := range ancestors {
		if !rectValid390(a.Rect) {
			return c, fmt.Errorf("invalid clipping ancestor")
		}
		if clipsAxis390(a.OverflowX) {
			c.Left = max(c.Left, a.Rect.Left)
			c.Right = min(c.Right, a.Rect.Right)
		}
		if clipsAxis390(a.OverflowY) {
			c.Top = max(c.Top, a.Rect.Top)
			c.Bottom = min(c.Bottom, a.Rect.Bottom)
		}
	}
	return c, nil
}
func clipsAxis390(value string) bool {
	return value == "auto" || value == "scroll" || value == "hidden" || value == "clip"
}
func sameClip390(a, b entryClip390) bool {
	return math.Abs(a.Left-b.Left) <= 0.1 && math.Abs(a.Right-b.Right) <= 0.1 && math.Abs(a.Top-b.Top) <= 0.1 && math.Abs(a.Bottom-b.Bottom) <= 0.1
}
func admitGeometry390(g entryGeometry390, width int, final bool) error {
	fail := func(reason string) error { return fmt.Errorf("native390 geometry: %s", reason) }
	height := 1000
	if width == 390 {
		height = 844
	}
	if g.Viewport.Width != float64(width) || g.Viewport.Height != float64(height) || g.Document == nil || !g.Root.Visible || !rectValid390(g.Root.Rect) || !g.Primary.Visible || !rectValid390(g.Primary.Rect) || !g.FirstItem.Visible || !rectValid390(g.FirstItem.Rect) {
		return fail("actual owner/primary/item or viewport missing")
	}
	clip, err := paintClip390(width, height, g.Ancestors)
	if err != nil {
		return fail(err.Error())
	}
	if math.Abs(clip.Left-g.Clip.Left) > 0.1 || math.Abs(clip.Right-g.Clip.Right) > 0.1 || math.Abs(clip.Top-g.Clip.Top) > 0.1 || math.Abs(clip.Bottom-g.Clip.Bottom) > 0.1 {
		return fail("reported clipping differs from browser ancestors")
	}
	for _, p := range []entryPaint390{g.Primary, g.FirstItem} {
		ownClip := clip
		if p.Clip != nil {
			ownClip, err = paintClip390(width, height, p.ClipAncestors)
			if err != nil || !sameClip390(ownClip, *p.Clip) {
				return fail("per-item clipping differs from actual ancestors")
			}
		} else if final {
			return fail("final per-item clipping proof missing")
		}
		want := clipped390(p.Rect, ownClip)
		if p.Intersection == nil || math.Abs(want.Width-p.Intersection.Width) > 0.1 || math.Abs(want.Height-p.Intersection.Height) > 0.1 || want.CenterInside != p.Intersection.CenterInside {
			return fail("reported intersection differs from actual rectangles")
		}
	}
	if final {
		primary := clipped390(g.Primary.Rect, clip)
		if primary.Height+0.1 < min(24, g.Primary.Rect.Height) || primary.Width <= 0 || !fullyInside390(g.FirstItem.Rect, *g.FirstItem.Clip) {
			return fail("first actual data item is not wholly visible at natural entry")
		}
		if g.Document.ScrollWidth > float64(width)+1 || g.ScrollX != 0 {
			return fail("horizontal page overflow")
		}
		if g.NumericPlot != nil && g.UnavailablePlot != nil {
			if !axisRectValid390(g.NumericPlot) || !rectValid390(g.UnavailablePlot) || math.Abs(g.NumericPlot.Left-g.UnavailablePlot.Left) > 1 || math.Abs(g.NumericPlot.Width-g.UnavailablePlot.Width) > 1 || g.UnavailablePlot.Top < g.NumericPlot.Bottom-1 {
				return fail("unavailable rail has a different time axis or numeric position")
			}
		}
		for _, mark := range g.Marks {
			if mark.Clip == nil {
				return fail("actual marker clipping proof missing")
			}
			markClip, e := paintClip390(width, height, mark.ClipAncestors)
			if e != nil || !sameClip390(markClip, *mark.Clip) {
				return fail("marker clipping differs from actual ancestors")
			}
			if !mark.Visible || !fullyInside390(mark.Rect, markClip) {
				return fail("an actual known/zero/unavailable mark is clipped at entry")
			}
		}
		if g.Timeline != nil && len(g.Marks) != len(g.Timeline.Points) {
			return fail("painted marks do not match supplied points")
		}
	}
	if math.Abs(g.Primary.Rect.AbsoluteDocumentTop-(g.Primary.Rect.Top+g.ScrollY)) > 0.1 || math.Abs(g.FirstItem.Rect.AbsoluteDocumentTop-(g.FirstItem.Rect.Top+g.ScrollY)) > 0.1 {
		return fail("absolute document positions changed independently of scroll")
	}
	return nil
}
