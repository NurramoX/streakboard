// Package tui provides a Bubble Tea component that shows a streakboard
// as a pixel image embedded in the layout, using the Kitty graphics
// protocol's Unicode placeholders (supported by Ghostty and kitty).
//
// The board is ordinary text to Bubble Tea — placeholder cells the
// terminal overlays with the image — so it scrolls, moves, and
// composes like any other view content, and keeps rendering from
// scrollback after the program exits. GitHub-style month and weekday
// labels are composed around the image as plain dimmed text in the
// terminal's own font. The image bytes are sent out-of-band via tea.Raw
// commands; run the program on a TrueColor profile so the id-encoding
// placeholder colors are not downsampled.
package tui

import (
	"math/rand/v2"
	"strings"
	"sync/atomic"
	"time"

	tea "charm.land/bubbletea/v2"

	streakboard "github.com/NurramoX/streakboard"
	"github.com/NurramoX/streakboard/kitty"
)

const (
	rows        = 7 // one terminal row per weekday
	colsPerWeek = 2 // two terminal columns per week keeps day cells square
	maxWeeks    = 53
	gutter      = 4 // columns for the "Mon " weekday labels left of the board
)

// lastID hands out one kitty image id per board so several boards can
// coexist in one program. It is seeded randomly so the placeholder
// text a previous run left in scrollback never matches a live id and
// gets repainted with this run's image.
var lastID atomic.Uint32

func init() { lastID.Store(rand.Uint32()) }

// nextID returns a fresh image id in 1..kitty.MaxID.
func nextID() uint32 { return lastID.Add(1)%kitty.MaxID + 1 }

// Model is a Bubble Tea component displaying one streakboard. Create it
// with New, route Update messages to it, and place View's text anywhere
// in the layout.
type Model struct {
	id      uint32
	entries []streakboard.Entry
	opts    streakboard.Options
	weeks   int
	view    string
}

// New returns a board showing entries rendered with o's palette, scale,
// and max. o.From and o.To are ignored: the window always ends today
// and sizes itself to the terminal width, up to one year.
func New(entries []streakboard.Entry, o streakboard.Options) Model {
	return Model{id: nextID(), entries: entries, opts: o}
}

// Set replaces the board's data and options and re-uploads the image.
func (m Model) Set(entries []streakboard.Entry, o streakboard.Options) (Model, tea.Cmd) {
	m.entries = entries
	m.opts = o
	return m.refresh()
}

func (m Model) Init() tea.Cmd { return nil }

// Update resizes the board window; the first tea.WindowSizeMsg also
// triggers the initial image upload. Other messages are ignored.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if size, ok := msg.(tea.WindowSizeMsg); ok {
		if weeks := min(maxWeeks, max(1, (size.Width-gutter)/colsPerWeek)); weeks != m.weeks {
			m.weeks = weeks
			return m.refresh()
		}
	}
	return m, nil
}

// View returns the board as text: a month-name header line, then 7
// placeholder lines — 2 cells per shown week, prefixed with a
// Mon/Wed/Fri gutter — that the terminal overlays with the image. It is
// empty until the first tea.WindowSizeMsg arrives.
func (m Model) View() string { return m.view }

// refresh renders the visible window, rebuilds the placeholder text,
// and returns the command that re-uploads the image under the board's
// id.
func (m Model) refresh() (Model, tea.Cmd) {
	if m.weeks == 0 {
		return m, nil
	}
	o := m.opts
	o.To = time.Now()
	sunday := o.To.AddDate(0, 0, -int(o.To.Weekday()))
	o.From = sunday.AddDate(0, 0, -7*(m.weeks-1))
	img := streakboard.Render(m.entries, o)

	cols := m.weeks * colsPerWeek
	transmit, err := kitty.TransmitPNG(m.id, img)
	if err != nil {
		m.view = err.Error()
		return m, nil
	}
	place, err := kitty.Place(m.id, rows, cols)
	if err != nil {
		m.view = err.Error()
		return m, nil
	}
	ph, err := kitty.Placeholder(m.id, rows, cols)
	if err != nil {
		m.view = err.Error()
		return m, nil
	}
	m.view = compose(o.From, o.To, m.weeks, ph)

	var b strings.Builder
	b.Write(kitty.Delete(m.id))
	b.Write(transmit)
	b.Write(place)
	return m, tea.Raw(b.String())
}

// compose wraps the placeholder text with GitHub-style labels: a
// month-name header line on top and Mon/Wed/Fri beside their rows.
// Labels are ordinary dimmed text; the dim attribute is reset before
// each placeholder run so it cannot touch the id-encoding foreground
// color.
func compose(from, to time.Time, weeks int, ph string) string {
	weekdays := [rows]string{1: "Mon", 3: "Wed", 5: "Fri"}
	var b strings.Builder
	b.WriteString(strings.Repeat(" ", gutter))
	b.WriteString(dim(monthHeader(from, to, weeks)))
	for r, line := range strings.Split(ph, "\n") {
		b.WriteByte('\n')
		if wd := weekdays[r]; wd != "" {
			b.WriteString(dim(wd))
			b.WriteString(strings.Repeat(" ", gutter-len(wd)))
		} else {
			b.WriteString(strings.Repeat(" ", gutter))
		}
		b.WriteString(line)
	}
	return b.String()
}

func dim(s string) string { return "\x1b[2m" + s + "\x1b[22m" }

// monthHeader returns a line as wide as the board with a "Jan"-style
// label over each week that contains the 1st of a month. from is the
// Sunday starting week 0; days after to are outside the window. A label
// in the final week would poke past the board's right edge and is
// dropped.
func monthHeader(from, to time.Time, weeks int) string {
	buf := []byte(strings.Repeat(" ", weeks*colsPerWeek))
	for w := range weeks {
		for i := range 7 {
			d := from.AddDate(0, 0, 7*w+i)
			if d.Day() != 1 || d.After(to) {
				continue
			}
			label := d.Format("Jan")
			if col := w * colsPerWeek; col+len(label) <= len(buf) {
				copy(buf[col:], label)
			}
		}
	}
	return strings.TrimRight(string(buf), " ")
}
