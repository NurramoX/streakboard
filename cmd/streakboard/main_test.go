package main

import (
	"strings"
	"testing"
	"time"
)

func TestReadEntries(t *testing.T) {
	entries, err := readEntries(strings.NewReader("2026-07-13 5\n\n2026-07-14 0\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("got %d entries, want 2", len(entries))
	}
	want := time.Date(2026, 7, 13, 0, 0, 0, 0, time.UTC)
	if !entries[0].Date.Equal(want) || entries[0].Count != 5 {
		t.Errorf("entry 0 = %+v", entries[0])
	}
}

func TestReadEntriesBadLine(t *testing.T) {
	for _, in := range []string{"2026-07-13", "not-a-date 5", "2026-07-13 many"} {
		if _, err := readEntries(strings.NewReader(in)); err == nil {
			t.Errorf("input %q: want error", in)
		}
	}
}
