package tui

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	streakboard "github.com/NurramoX/streakboard"
)

// raw runs cmd and returns the RawMsg payload it carries.
func raw(t *testing.T, cmd tea.Cmd) string {
	t.Helper()
	if cmd == nil {
		t.Fatal("want a command, got nil")
	}
	msg, ok := cmd().(tea.RawMsg)
	if !ok {
		t.Fatalf("want tea.RawMsg, got %T", cmd())
	}
	s, ok := msg.Msg.(string)
	if !ok {
		t.Fatalf("want string payload, got %T", msg.Msg)
	}
	return s
}

func TestModelUploadsOnFirstSize(t *testing.T) {
	entries := []streakboard.Entry{{Date: time.Now(), Count: 3}}
	m := New(entries, streakboard.Options{Max: 5})

	m, cmd := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	out := raw(t, cmd)
	for _, want := range []string{"a=d,d=I,", "a=t,f=100,q=2,", "a=p,U=1,q=2,", ",r=7,c=76"} {
		if !strings.Contains(out, want) {
			t.Errorf("upload missing %q", want)
		}
	}

	// 80 columns minus the gutter fit 38 weeks: a month-name header,
	// then 7 lines of 76 placeholder cells each behind the gutter.
	lines := strings.Split(m.View(), "\n")
	if len(lines) != 8 {
		t.Fatalf("got %d lines, want 8", len(lines))
	}
	if got := strings.Count(lines[1], "\U0010EEEE"); got != 76 {
		t.Errorf("got %d cells per line, want 76", got)
	}
	for line, wd := range map[int]string{2: "Mon", 4: "Wed", 6: "Fri"} {
		if !strings.Contains(lines[line], wd) {
			t.Errorf("line %d missing %q gutter label", line, wd)
		}
	}
}

func TestModelIgnoresRepeatSize(t *testing.T) {
	m := New(nil, streakboard.Options{})
	m, _ = m.Update(tea.WindowSizeMsg{Width: 40, Height: 24})
	if _, cmd := m.Update(tea.WindowSizeMsg{Width: 41, Height: 24}); cmd != nil {
		t.Error("same week count re-uploaded the image")
	}
}

func TestModelCapsAtOneYear(t *testing.T) {
	m := New(nil, streakboard.Options{})
	m, _ = m.Update(tea.WindowSizeMsg{Width: 500, Height: 24})
	if got := strings.Count(strings.Split(m.View(), "\n")[1], "\U0010EEEE"); got != 2*53 {
		t.Errorf("got %d cells per line, want %d", got, 2*53)
	}
}

// sunday parses a date that a fixture claims is a Sunday.
func sunday(t *testing.T, s string) time.Time {
	t.Helper()
	d, err := time.Parse("2006-01-02", s)
	if err != nil || d.Weekday() != time.Sunday {
		t.Fatalf("bad fixture %q: %v, weekday %v", s, err, d.Weekday())
	}
	return d
}

func TestMonthHeader(t *testing.T) {
	tests := []struct {
		name  string
		from  string
		weeks int
		days  int // window length, in days from from
		want  string
	}{
		{name: "label over each week with a 1st", from: "2026-02-01", weeks: 10, days: 66,
			want: "Feb     Mar     Apr"},
		{name: "no 1st in window", from: "2026-01-04", weeks: 3, days: 20, want: ""},
		{name: "1st in final week dropped", from: "2026-01-04", weeks: 5, days: 34, want: ""},
		{name: "1st after to ignored", from: "2026-06-28", weeks: 2, days: 2, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			from := sunday(t, tt.from)
			got := monthHeader(from, from.AddDate(0, 0, tt.days), tt.weeks)
			if got != tt.want {
				t.Errorf("monthHeader(%s, +%dd, %d weeks) = %q, want %q",
					tt.from, tt.days, tt.weeks, got, tt.want)
			}
		})
	}
}

func TestModelsGetDistinctIDs(t *testing.T) {
	a, b := New(nil, streakboard.Options{}), New(nil, streakboard.Options{})
	if a.id == b.id {
		t.Errorf("both boards got id %d", a.id)
	}
}

func TestViewEmptyBeforeSize(t *testing.T) {
	if v := New(nil, streakboard.Options{}).View(); v != "" {
		t.Errorf("View before sizing = %q, want empty", v)
	}
}
