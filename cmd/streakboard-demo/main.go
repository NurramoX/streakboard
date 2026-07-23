// Command streakboard-demo shows the streakboard Bubble Tea component
// with generated habit data. Run it in Ghostty or kitty; press t to
// cycle themes, n for new data, q to quit.
package main

import (
	"fmt"
	"math/rand/v2"
	"os"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"

	streakboard "github.com/NurramoX/streakboard"
	"github.com/NurramoX/streakboard/tui"
)

var themes = []struct {
	name    string
	palette streakboard.Palette
}{
	{"github", streakboard.GitHubDark},
	{"github-light", streakboard.GitHubLight},
	{"catppuccin-latte", streakboard.CatppuccinLatte},
	{"catppuccin-frappe", streakboard.CatppuccinFrappe},
	{"catppuccin-macchiato", streakboard.CatppuccinMacchiato},
	{"catppuccin-mocha", streakboard.CatppuccinMocha},
}

type model struct {
	board tui.Model
	theme int
}

func (m model) options() streakboard.Options {
	return streakboard.Options{Palette: themes[m.theme].palette, Max: 25, Scale: 3}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyPressMsg); ok {
		switch key.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "t":
			m.theme = (m.theme + 1) % len(themes)
			var cmd tea.Cmd
			m.board, cmd = m.board.Set(entries(), m.options())
			return m, cmd
		case "n":
			var cmd tea.Cmd
			m.board, cmd = m.board.Set(entries(), m.options())
			return m, cmd
		}
	}
	var cmd tea.Cmd
	m.board, cmd = m.board.Update(msg)
	return m, cmd
}

func (m model) View() tea.View {
	return tea.NewView(fmt.Sprintf("  habit board — %s\n\n%s\n\n  t theme · n new data · q quit\n",
		themes[m.theme].name, m.board.View()))
}

// entries fakes a year of habit data: weekday-heavy, lazy weekends, and
// an off week every couple of months.
func entries() []streakboard.Entry {
	var es []streakboard.Entry
	today := time.Now()
	for i := range 365 {
		d := today.AddDate(0, 0, -i)
		if (i/7)%9 == 4 {
			continue
		}
		p := 0.75
		if wd := d.Weekday(); wd == time.Saturday || wd == time.Sunday {
			p = 0.35
		}
		if rand.Float64() < p {
			es = append(es, streakboard.Entry{Date: d, Count: rand.IntN(25) + 1})
		}
	}
	return es
}

func main() {
	m := model{}
	m.board = tui.New(entries(), m.options())
	p := tea.NewProgram(m, tea.WithColorProfile(colorprofile.TrueColor))
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "streakboard-demo:", err)
		os.Exit(1)
	}
}
