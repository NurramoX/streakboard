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
	for _, want := range []string{"a=d,d=I,", "a=t,f=100,q=2,", "a=p,U=1,q=2,", ",r=7,c=80"} {
		if !strings.Contains(out, want) {
			t.Errorf("upload missing %q", want)
		}
	}

	// 80 columns fit 40 weeks: 7 lines of 80 placeholder cells each.
	lines := strings.Split(m.View(), "\n")
	if len(lines) != 7 {
		t.Fatalf("got %d lines, want 7", len(lines))
	}
	if got := strings.Count(lines[0], "\U0010EEEE"); got != 80 {
		t.Errorf("got %d cells per line, want 80", got)
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
	if got := strings.Count(strings.Split(m.View(), "\n")[0], "\U0010EEEE"); got != 2*53 {
		t.Errorf("got %d cells per line, want %d", got, 2*53)
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
