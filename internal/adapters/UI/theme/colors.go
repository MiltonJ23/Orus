package theme

import "image/color"

// hex2Color is a helper to convert hex to Gio's color.NRGBA
func hex2Color(c uint32) color.NRGBA {
	return color.NRGBA{
		R: uint8(c >> 16),
		G: uint8(c >> 8),
		B: uint8(c),
		A: 0xFF,
	}
}

var (
	ColorVoidDark = hex2Color(0x0D0D12)

	ColorSandGold = hex2Color(0xD4AF37)

	ColorCyberCyan = hex2Color(0x3D4451)

	ColorGlassWhite = hex2Color(0xFFFFFF)
)
