// Package tui provides a Bubble Tea component that shows a streakboard
// as a pixel image embedded in the layout, using the Kitty graphics
// protocol's Unicode placeholders (supported by Ghostty and kitty).
//
// The board is ordinary text to Bubble Tea — placeholder cells the
// terminal overlays with the image — so it scrolls, moves, and
// composes like any other view content. The image bytes are sent
// out-of-band via tea.Raw commands; run the program on a TrueColor
// profile so the id-encoding placeholder colors are not downsampled.
package tui

import (
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
)

// lastID hands out one kitty image id per board so several boards can
// coexist in one program.
var lastID atomic.Uint32

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
	return Model{id: lastID.Add(1), entries: entries, opts: o}
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
		if weeks := min(maxWeeks, max(1, size.Width/colsPerWeek)); weeks != m.weeks {
			m.weeks = weeks
			return m.refresh()
		}
	}
	return m, nil
}

// View returns the placeholder text the terminal overlays with the
// image: 7 lines of 2 cells per shown week. It is empty until the
// first tea.WindowSizeMsg arrives.
func (m Model) View() string { return m.view }

// Close returns the command that removes the board's image from the
// terminal; sequence it before tea.Quit.
func (m Model) Close() tea.Cmd {
	return tea.Raw(string(kitty.Delete(m.id)))
}

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
	m.view = ph

	var b strings.Builder
	b.Write(kitty.Delete(m.id))
	b.Write(transmit)
	b.Write(place)
	return m, tea.Raw(b.String())
}
