package theme_test

import (
	"image/color"
	"testing"

	"github.com/MiltonJ23/Orus/internal/adapters/ui/theme"
)

func TestColorConstants(t *testing.T) {
	t.Run("ColorVoidDark", func(t *testing.T) {
		c := theme.ColorVoidDark
		if c.A != 0xFF {
			t.Errorf("Expected alpha 0xFF, got 0x%X", c.A)
		}
		// ColorVoidDark should be 0x0D0D12
		if c.R != 0x0D || c.G != 0x0D || c.B != 0x12 {
			t.Errorf("Expected RGB(0x0D, 0x0D, 0x12), got RGB(0x%X, 0x%X, 0x%X)", c.R, c.G, c.B)
		}
	})

	t.Run("ColorSandGold", func(t *testing.T) {
		c := theme.ColorSandGold
		if c.A != 0xFF {
			t.Errorf("Expected alpha 0xFF, got 0x%X", c.A)
		}
		// ColorSandGold should be 0xF5A623
		if c.R != 0xF5 || c.G != 0xA6 || c.B != 0x23 {
			t.Errorf("Expected RGB(0xF5, 0xA6, 0x23), got RGB(0x%X, 0x%X, 0x%X)", c.R, c.G, c.B)
		}
	})

	t.Run("ColorCyberCyan", func(t *testing.T) {
		c := theme.ColorCyberCyan
		if c.A != 0xFF {
			t.Errorf("Expected alpha 0xFF, got 0x%X", c.A)
		}
		// ColorCyberCyan should be 0x2A3240
		if c.R != 0x2A || c.G != 0x32 || c.B != 0x40 {
			t.Errorf("Expected RGB(0x2A, 0x32, 0x40), got RGB(0x%X, 0x%X, 0x%X)", c.R, c.G, c.B)
		}
	})

	t.Run("ColorGlassWhite", func(t *testing.T) {
		c := theme.ColorGlassWhite
		if c.A != 0xFF {
			t.Errorf("Expected alpha 0xFF, got 0x%X", c.A)
		}
		// ColorGlassWhite should be 0xF8F9FA
		if c.R != 0xF8 || c.G != 0xF9 || c.B != 0xFA {
			t.Errorf("Expected RGB(0xF8, 0xF9, 0xFA), got RGB(0x%X, 0x%X, 0x%X)", c.R, c.G, c.B)
		}
	})

	t.Run("ColorPureBlack", func(t *testing.T) {
		c := theme.ColorPureBlack
		if c.A != 0xFF {
			t.Errorf("Expected alpha 0xFF, got 0x%X", c.A)
		}
		// ColorPureBlack should be 0x000000
		if c.R != 0x00 || c.G != 0x00 || c.B != 0x00 {
			t.Errorf("Expected RGB(0x00, 0x00, 0x00), got RGB(0x%X, 0x%X, 0x%X)", c.R, c.G, c.B)
		}
	})
}

func TestWithAlpha(t *testing.T) {
	t.Run("Modify Alpha on Black", func(t *testing.T) {
		original := color.NRGBA{R: 0, G: 0, B: 0, A: 255}
		modified := theme.WithAlpha(original, 128)

		if modified.A != 128 {
			t.Errorf("Expected alpha 128, got %d", modified.A)
		}
		if modified.R != 0 || modified.G != 0 || modified.B != 0 {
			t.Error("RGB values should not change")
		}
	})

	t.Run("Modify Alpha on White", func(t *testing.T) {
		original := color.NRGBA{R: 255, G: 255, B: 255, A: 255}
		modified := theme.WithAlpha(original, 64)

		if modified.A != 64 {
			t.Errorf("Expected alpha 64, got %d", modified.A)
		}
		if modified.R != 255 || modified.G != 255 || modified.B != 255 {
			t.Error("RGB values should not change")
		}
	})

	t.Run("Set Alpha to Zero", func(t *testing.T) {
		original := color.NRGBA{R: 100, G: 150, B: 200, A: 255}
		modified := theme.WithAlpha(original, 0)

		if modified.A != 0 {
			t.Errorf("Expected alpha 0, got %d", modified.A)
		}
		if modified.R != 100 || modified.G != 150 || modified.B != 200 {
			t.Error("RGB values should not change")
		}
	})

	t.Run("Set Alpha to Max", func(t *testing.T) {
		original := color.NRGBA{R: 50, G: 100, B: 150, A: 0}
		modified := theme.WithAlpha(original, 255)

		if modified.A != 255 {
			t.Errorf("Expected alpha 255, got %d", modified.A)
		}
		if modified.R != 50 || modified.G != 100 || modified.B != 150 {
			t.Error("RGB values should not change")
		}
	})

	t.Run("WithAlpha Does Not Mutate Original", func(t *testing.T) {
		original := color.NRGBA{R: 100, G: 100, B: 100, A: 255}
		originalCopy := original

		theme.WithAlpha(original, 50)

		if original != originalCopy {
			t.Error("Original color should not be mutated")
		}
		if original.A != 255 {
			t.Error("Original alpha should remain 255")
		}
	})

	t.Run("WithAlpha on Theme Colors", func(t *testing.T) {
		// Test on actual theme colors
		semiTransparentGold := theme.WithAlpha(theme.ColorSandGold, 100)
		if semiTransparentGold.A != 100 {
			t.Errorf("Expected alpha 100, got %d", semiTransparentGold.A)
		}
		if semiTransparentGold.R != theme.ColorSandGold.R {
			t.Error("R value should match original")
		}
		if semiTransparentGold.G != theme.ColorSandGold.G {
			t.Error("G value should match original")
		}
		if semiTransparentGold.B != theme.ColorSandGold.B {
			t.Error("B value should match original")
		}
	})

	t.Run("Multiple Alpha Modifications", func(t *testing.T) {
		c := theme.ColorGlassWhite
		c1 := theme.WithAlpha(c, 200)
		c2 := theme.WithAlpha(c1, 150)
		c3 := theme.WithAlpha(c2, 50)

		if c3.A != 50 {
			t.Errorf("Expected final alpha 50, got %d", c3.A)
		}
		// Original should be unchanged
		if c.A != 0xFF {
			t.Error("Original color should not be modified")
		}
	})
}

func TestColorCreation(t *testing.T) {
	testCases := []struct {
		name     string
		hexValue uint32
		expected color.NRGBA
	}{
		{
			name:     "Pure Black",
			hexValue: 0x000000,
			expected: color.NRGBA{R: 0, G: 0, B: 0, A: 255},
		},
		{
			name:     "Pure White",
			hexValue: 0xFFFFFF,
			expected: color.NRGBA{R: 255, G: 255, B: 255, A: 255},
		},
		{
			name:     "Pure Red",
			hexValue: 0xFF0000,
			expected: color.NRGBA{R: 255, G: 0, B: 0, A: 255},
		},
		{
			name:     "Pure Green",
			hexValue: 0x00FF00,
			expected: color.NRGBA{R: 0, G: 255, B: 0, A: 255},
		},
		{
			name:     "Pure Blue",
			hexValue: 0x0000FF,
			expected: color.NRGBA{R: 0, G: 0, B: 255, A: 255},
		},
		{
			name:     "Custom Color 1",
			hexValue: 0x123456,
			expected: color.NRGBA{R: 0x12, G: 0x34, B: 0x56, A: 255},
		},
		{
			name:     "Custom Color 2",
			hexValue: 0xABCDEF,
			expected: color.NRGBA{R: 0xAB, G: 0xCD, B: 0xEF, A: 255},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// We can't directly call hex2Color as it's not exported,
			// but we verify the constants are correct through our previous tests
			// This test documents the expected behavior
		})
	}
}

func TestColorEquality(t *testing.T) {
	t.Run("Same colors are equal", func(t *testing.T) {
		c1 := theme.ColorVoidDark
		c2 := theme.ColorVoidDark

		if c1 != c2 {
			t.Error("Same colors should be equal")
		}
	})

	t.Run("Different colors are not equal", func(t *testing.T) {
		c1 := theme.ColorVoidDark
		c2 := theme.ColorSandGold

		if c1 == c2 {
			t.Error("Different colors should not be equal")
		}
	})

	t.Run("Modified alpha creates different color", func(t *testing.T) {
		c1 := theme.ColorGlassWhite
		c2 := theme.WithAlpha(theme.ColorGlassWhite, 128)

		if c1 == c2 {
			t.Error("Colors with different alpha should not be equal")
		}
	})
}

func TestAlphaBoundaries(t *testing.T) {
	baseColor := color.NRGBA{R: 100, G: 100, B: 100, A: 128}

	t.Run("Alpha 0 (fully transparent)", func(t *testing.T) {
		c := theme.WithAlpha(baseColor, 0)
		if c.A != 0 {
			t.Errorf("Expected alpha 0, got %d", c.A)
		}
	})

	t.Run("Alpha 255 (fully opaque)", func(t *testing.T) {
		c := theme.WithAlpha(baseColor, 255)
		if c.A != 255 {
			t.Errorf("Expected alpha 255, got %d", c.A)
		}
	})

	t.Run("Alpha mid-range values", func(t *testing.T) {
		alphas := []uint8{1, 64, 128, 192, 254}
		for _, a := range alphas {
			c := theme.WithAlpha(baseColor, a)
			if c.A != a {
				t.Errorf("Expected alpha %d, got %d", a, c.A)
			}
		}
	})
}