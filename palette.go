package streakboard

import "image/color"

// Palette holds the cell colors for intensity levels 0 through 4.
type Palette [5]color.NRGBA

// GitHub's contribution-graph greens.
var (
	GitHubDark  = Palette{hex(0x161b22), hex(0x0e4429), hex(0x006d32), hex(0x26a641), hex(0x39d353)}
	GitHubLight = Palette{hex(0xebedf0), hex(0x9be9a8), hex(0x40c463), hex(0x30a14e), hex(0x216e39)}
)

// Catppuccin flavors (https://catppuccin.com/palette): empty cells use
// the flavor's surface0, active levels blend from base toward its green.
var (
	CatppuccinLatte     = ramp(hex(0xccd0da), hex(0xeff1f5), hex(0x40a02b))
	CatppuccinFrappe    = ramp(hex(0x414559), hex(0x303446), hex(0xa6d189))
	CatppuccinMacchiato = ramp(hex(0x363a4f), hex(0x24273a), hex(0xa6da95))
	CatppuccinMocha     = ramp(hex(0x313244), hex(0x1e1e2e), hex(0xa6e3a1))
)

func hex(rgb uint32) color.NRGBA {
	return color.NRGBA{R: uint8(rgb >> 16), G: uint8(rgb >> 8), B: uint8(rgb), A: 0xff}
}

func ramp(empty, base, full color.NRGBA) Palette {
	return Palette{empty, mix(base, full, 0.25), mix(base, full, 0.5), mix(base, full, 0.75), full}
}

func mix(a, b color.NRGBA, t float64) color.NRGBA {
	l := func(x, y uint8) uint8 { return uint8(float64(x) + t*(float64(y)-float64(x)) + 0.5) }
	return color.NRGBA{R: l(a.R, b.R), G: l(a.G, b.G), B: l(a.B, b.B), A: 0xff}
}
