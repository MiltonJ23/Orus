package theme

import "image/color"

func hex2Color(c uint32) color.NRGBA {
	return color.NRGBA{
		R: uint8(c >> 16),
		G: uint8(c >> 8),
		B: uint8(c),
		A: 0xFF,
	}
}

func WithAlpha(c color.NRGBA, alpha uint8) color.NRGBA {
	c.A = alpha
	return c
}

var (
	ColorVoidDark   = hex2Color(0x0D0D12) // Noir de jais
	ColorSandGold   = hex2Color(0xF5A623)
	ColorCyberCyan  = hex2Color(0x2A3240) // Ardoise du Nil
	ColorGlassWhite = hex2Color(0xF8F9FA) // Blanc Albâtre
	ColorPureBlack  = hex2Color(0x000000)
)
