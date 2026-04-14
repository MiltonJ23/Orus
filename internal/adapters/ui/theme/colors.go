package theme

import "image/color"

// hex2Color converts a 24-bit hex color to an NRGBA value with full opacity.
func hex2Color(c uint32) color.NRGBA {
	return color.NRGBA{
		R: uint8(c >> 16),
		G: uint8(c >> 8),
		B: uint8(c),
		A: 0xFF,
	}
}

// WithAlpha returns a copy of c with the given alpha transparency.
func WithAlpha(c color.NRGBA, alpha uint8) color.NRGBA {
	c.A = alpha
	return c
}

// Mecha-Egyptian color palette.
var (
	ColorVoidDark   = hex2Color(0x0D0D12) // Jet black background
	ColorSandGold   = hex2Color(0xF5A623) // Egyptian gold accent
	ColorCyberCyan  = hex2Color(0x2A3240) // Nile slate surface
	ColorGlassWhite = hex2Color(0xF8F9FA) // Alabaster white text
	ColorPureBlack  = hex2Color(0x000000) // Deepest shadow
)
