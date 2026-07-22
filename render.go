package streakboard

import (
	"image"
	"image/color"
	"math"
	"time"
)

// Options control Render. The zero value shows the last 365 days ending
// today, GitHub dark theme, at 2x GitHub's native geometry (10px cells,
// 3px gaps, 2px corner radius).
type Options struct {
	Palette Palette   // colors for levels 0-4; zero value means GitHubDark
	Scale   int       // multiplier over the native geometry; <1 means 2
	From    time.Time // first day shown; zero means 364 days before To
	To      time.Time // last day shown; zero means today
	Max     int       // count drawn at full intensity; <1 means the highest count in range
}

// Render draws one rounded square per day in [From, To] — one column per
// Sunday-anchored week — on a transparent background. Days without
// entries are drawn at level 0; positive counts are bucketed into levels
// 1-4 relative to Max, like GitHub's quartiles.
func Render(entries []Entry, o Options) *image.NRGBA {
	if o.Palette == (Palette{}) {
		o.Palette = GitHubDark
	}
	if o.Scale < 1 {
		o.Scale = 2
	}
	to := o.To
	if to.IsZero() {
		to = time.Now()
	}
	to = midnight(to)
	from := o.From
	if from.IsZero() {
		from = to.AddDate(0, 0, -364)
	}
	from = midnight(from)
	if from.After(to) {
		return image.NewNRGBA(image.Rectangle{})
	}

	counts := make(map[time.Time]int, len(entries))
	for _, e := range entries {
		counts[midnight(e.Date)] += e.Count
	}
	maxCount := o.Max
	if maxCount < 1 {
		for d, n := range counts {
			if !d.Before(from) && !d.After(to) && n > maxCount {
				maxCount = n
			}
		}
	}

	cell := 10 * o.Scale
	pitch := cell + 3*o.Scale
	radius := float64(2 * o.Scale)
	start := from.AddDate(0, 0, -int(from.Weekday())) // Sunday on or before From
	weeks := int(to.Sub(start)/(24*time.Hour))/7 + 1

	img := image.NewNRGBA(image.Rect(0, 0, weeks*pitch-3*o.Scale, 7*pitch-3*o.Scale))
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		days := int(d.Sub(start) / (24 * time.Hour))
		x := days / 7 * pitch
		y := days % 7 * pitch
		c := o.Palette[level(counts[d], maxCount)]
		fillRoundedRect(img, image.Rect(x, y, x+cell, y+cell), radius, c)
	}
	return img
}

// level buckets a count into 1-4 relative to max, like GitHub's quartiles.
func level(count, max int) int {
	if count < 1 || max < 1 {
		return 0
	}
	return min((4*count+max-1)/max, 4)
}

func midnight(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// fillRoundedRect fills r with c, rounding the corners with the given
// radius and anti-aliasing the edge. It overwrites destination pixels, so
// it assumes non-overlapping rectangles on a transparent background.
func fillRoundedRect(img *image.NRGBA, r image.Rectangle, radius float64, c color.NRGBA) {
	cx := float64(r.Min.X+r.Max.X) / 2
	cy := float64(r.Min.Y+r.Max.Y) / 2
	hw := float64(r.Dx())/2 - radius
	hh := float64(r.Dy())/2 - radius
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			dx := math.Max(math.Abs(float64(x)+0.5-cx)-hw, 0)
			dy := math.Max(math.Abs(float64(y)+0.5-cy)-hh, 0)
			cov := math.Min(math.Max(radius+0.5-math.Hypot(dx, dy), 0), 1)
			if cov == 0 {
				continue
			}
			p := c
			p.A = uint8(float64(c.A)*cov + 0.5)
			img.SetNRGBA(x, y, p)
		}
	}
}
