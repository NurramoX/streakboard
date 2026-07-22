package streakboard

import (
	"image"
	"testing"
	"time"
)

func date(d int) time.Time {
	return time.Date(2026, 7, d, 0, 0, 0, 0, time.UTC)
}

// center returns the pixel at the middle of the cell in the given
// week column and weekday row, for a scale-1 render (13px pitch).
func center(img *image.NRGBA, week, row int) [4]uint8 {
	c := img.NRGBAAt(week*13+5, row*13+5)
	return [4]uint8{c.R, c.G, c.B, c.A}
}

func rgba(p Palette, level int) [4]uint8 {
	c := p[level]
	return [4]uint8{c.R, c.G, c.B, c.A}
}

func TestRenderGeometry(t *testing.T) {
	// Mon 2026-07-13 through Wed 2026-07-22 spans two Sunday-anchored weeks.
	var entries []Entry
	for d := 13; d <= 22; d++ {
		entries = append(entries, Entry{Date: date(d), Count: 1})
	}
	img := Render(entries, Options{Scale: 1, From: date(13), To: date(22)})

	// Two week columns at scale 1: 2*13-3 = 23 wide; 7 rows: 7*13-3 = 88 high.
	if got := img.Bounds(); got.Dx() != 23 || got.Dy() != 88 {
		t.Fatalf("bounds = %v, want 23x88", got)
	}

	// Every count equals the max, so every day is level 4.
	// Mon 2026-07-20 sits in week 1, row 1.
	if got := center(img, 1, 1); got != rgba(GitHubDark, 4) {
		t.Errorf("2026-07-20 cell = %v, want level 4", got)
	}

	// Sun 2026-07-12 is before From: nothing drawn there.
	if got := center(img, 0, 0); got[3] != 0 {
		t.Errorf("cell before From = %v, want transparent", got)
	}
}

func TestRenderFillsMissingDays(t *testing.T) {
	img := Render(nil, Options{Scale: 1, From: date(13), To: date(22)})
	// Mon 2026-07-20 has no entry: drawn at level 0, not left transparent.
	if got := center(img, 1, 1); got != rgba(GitHubDark, 0) {
		t.Errorf("empty day = %v, want level 0", got)
	}
}

func TestRenderLevels(t *testing.T) {
	// Sun 2026-06-28 (week 0, row 0) onward, counts 1..8; max is 8, so
	// GitHub-style quartiles: ceil(4c/8) = 1,1,2,2,3,3,4,4.
	var entries []Entry
	for i := range 8 {
		entries = append(entries, Entry{Date: time.Date(2026, 6, 28+i, 0, 0, 0, 0, time.UTC), Count: i + 1})
	}
	img := Render(entries, Options{Scale: 1, From: entries[0].Date, To: entries[7].Date})
	want := []int{1, 1, 2, 2, 3, 3, 4, 4}
	for i, lv := range want {
		if got := center(img, i/7, i%7); got != rgba(GitHubDark, lv) {
			t.Errorf("count %d = %v, want level %d", i+1, got, lv)
		}
	}
}

func TestRenderFixedMax(t *testing.T) {
	// With Max pinned to 100, a count of 5 is level 1 even though it is
	// the highest count on the board; a count over Max clamps to 4.
	entries := []Entry{
		{Date: date(13), Count: 5},
		{Date: date(14), Count: 500},
	}
	img := Render(entries, Options{Scale: 1, From: date(13), To: date(14), Max: 100})
	if got := center(img, 0, 1); got != rgba(GitHubDark, 1) {
		t.Errorf("count 5 of max 100 = %v, want level 1", got)
	}
	if got := center(img, 0, 2); got != rgba(GitHubDark, 4) {
		t.Errorf("count 500 of max 100 = %v, want level 4", got)
	}
}

func TestRenderSumsSameDay(t *testing.T) {
	// 2+3 on the same date with Max 5 must reach level 4; without
	// summing, neither entry alone would.
	entries := []Entry{
		{Date: date(13), Count: 2},
		{Date: date(13), Count: 3},
	}
	img := Render(entries, Options{Scale: 1, From: date(13), To: date(13), Max: 5})
	if got := center(img, 0, 1); got != rgba(GitHubDark, 4) {
		t.Errorf("summed day = %v, want level 4", got)
	}
}
