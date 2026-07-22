# streakboard

A GitHub-style contribution board ("commit board") you feed with
your own per-day activity data — habits, vocabulary reviews,
anything countable. Renders to a PNG and embeds in Bubble Tea
TUIs as a real pixel image via the Kitty graphics protocol
(Ghostty and kitty).

## Bubble Tea component

```go
board := tui.New(entries, streakboard.Options{Max: 20, Scale: 3})
// route Update messages to it, put board.View() in your layout,
// and sequence board.Close() before tea.Quit.
```

The board sizes itself to the terminal (two columns per week, up
to a year) and is ordinary text to Bubble Tea: Unicode
placeholder cells the terminal overlays with the image, so it
scrolls and composes like any view content. GitHub-style month
and weekday labels are composed around the image as dimmed text
in the terminal's own font. Image bytes travel
out-of-band as tea.Raw commands. Needs bubbletea v2 and a
TrueColor profile (`tea.WithColorProfile`).

Try it in Ghostty: `go run ./cmd/streakboard-demo` (t cycles
themes, n regenerates data, q quits).

Package `kitty` underneath is framework-agnostic and
stdlib-only: TransmitPNG / Place / Placeholder / Delete, if you
want to integrate without Bubble Tea.

## Library

```go
entries := []streakboard.Entry{
	{Date: time.Now(), Count: 12}, // 12 vocab words reviewed today
}
img := streakboard.Render(entries, streakboard.Options{
	Max:     20, // 20+ per day = full intensity; 0 scales to the best day
	Palette: streakboard.CatppuccinMocha,
})
// img is an *image.NRGBA on a transparent background
```

Palettes: `GitHubDark` (default), `GitHubLight`, and the four
Catppuccin flavors `CatppuccinLatte`, `CatppuccinFrappe`,
`CatppuccinMacchiato`, `CatppuccinMocha` — or bring your own
`Palette` (five `color.NRGBA` values, level 0 through 4).

`Options` zero value shows the last 365 days ending today, dark
theme, at 2x GitHub's native geometry (10px cells, 3px gaps, 2px
radius). `From`/ `To` pick a different window; missing days
render as empty (level 0) cells, and same-day entries are
summed. Counts bucket into levels 1-4 relative to `Max`, like
GitHub's quartiles.

## CLI

Reads `YYYY-MM-DD COUNT` lines from stdin:

```sh
go run ./cmd/streakboard -max 20 -o board.png < activity.txt
kitty +kitten icat board.png
```

Flags: `-o` output path, `-scale N`, `-max N`, and `-theme` with
`github`, `github-light`, `catppuccin` (mocha),
`catppuccin-latte`, `catppuccin-frappe`, `catppuccin-macchiato`,
or `catppuccin-mocha`.
