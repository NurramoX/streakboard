// Command streakboard renders an activity board to a PNG file from
// "YYYY-MM-DD COUNT" lines on stdin.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"image/png"
	"io"
	"maps"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	streakboard "github.com/NurramoX/streakboard"
)

var themes = map[string]streakboard.Palette{
	"github":               streakboard.GitHubDark,
	"github-light":         streakboard.GitHubLight,
	"catppuccin":           streakboard.CatppuccinMocha,
	"catppuccin-latte":     streakboard.CatppuccinLatte,
	"catppuccin-frappe":    streakboard.CatppuccinFrappe,
	"catppuccin-macchiato": streakboard.CatppuccinMacchiato,
	"catppuccin-mocha":     streakboard.CatppuccinMocha,
}

func main() {
	out := flag.String("o", "board.png", "output PNG path")
	theme := flag.String("theme", "github", "color theme: github, github-light, or catppuccin[-latte|-frappe|-macchiato|-mocha]")
	scale := flag.Int("scale", 2, "pixel scale factor")
	max := flag.Int("max", 0, "count drawn at full intensity (0 = highest count in data)")
	flag.Parse()

	if err := run(*out, *theme, *scale, *max); err != nil {
		fmt.Fprintln(os.Stderr, "streakboard:", err)
		os.Exit(1)
	}
}

func run(out, theme string, scale, max int) error {
	pal, ok := themes[theme]
	if !ok {
		return fmt.Errorf("unknown theme %q (have %s)",
			theme, strings.Join(slices.Sorted(maps.Keys(themes)), ", "))
	}

	entries, err := readEntries(os.Stdin)
	if err != nil {
		return err
	}
	img := streakboard.Render(entries, streakboard.Options{Palette: pal, Scale: scale, Max: max})

	f, err := os.OpenFile(out, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	if err := png.Encode(f, img); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

// readEntries parses "YYYY-MM-DD COUNT" lines; blank lines are skipped.
func readEntries(r io.Reader) ([]streakboard.Entry, error) {
	sc := bufio.NewScanner(r)
	var entries []streakboard.Entry
	line := 0
	for sc.Scan() {
		line++
		fields := strings.Fields(sc.Text())
		if len(fields) == 0 {
			continue
		}
		if len(fields) != 2 {
			return nil, fmt.Errorf("line %d: want \"YYYY-MM-DD COUNT\", got %q", line, sc.Text())
		}
		date, err := time.Parse("2006-01-02", fields[0])
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", line, err)
		}
		count, err := strconv.Atoi(fields[1])
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", line, err)
		}
		entries = append(entries, streakboard.Entry{Date: date, Count: count})
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}
