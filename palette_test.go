package main

import (
	"image/color"
	"testing"
)

func TestPaletteInsideIsBlack(t *testing.T) {
	p := NewClassicPalette()
	if got := p.Color(100, 100); got != p.Inside {
		t.Errorf("Inside erwartet, got %v", got)
	}
}

func TestPaletteHitsStopsAndCycles(t *testing.T) {
	p := NewClassicPalette()
	if got := p.Color(0, 1000); got != color.Color(p.Stops[0]) {
		t.Errorf("iter 0 = %v, want %v", got, p.Stops[0])
	}
	if p.Color(3, 1000) != p.Color(3+p.Period, 1000) {
		t.Error("Palette ist nicht periodisch")
	}
}

func TestAvailablePalettes(t *testing.T) {
	palettes := AvailablePalettes()
	if len(palettes) != 4 {
		t.Fatalf("got %d palettes, want 4", len(palettes))
	}
	for _, option := range palettes {
		t.Run(option.Name, func(t *testing.T) {
			if option.Palette == nil {
				t.Fatal("palette is nil")
			}
			if got := option.Palette.Color(100, 100); got == nil {
				t.Fatal("outside color is nil")
			}
			if got := option.Palette.Color(17, 100); got == option.Palette.Color(100, 100) {
				t.Fatal("inside and outside colors unexpectedly match")
			}
		})
	}
}

func TestLerp(t *testing.T) {
	a, b := color.RGBA{0, 0, 0, 255}, color.RGBA{200, 100, 50, 255}
	if got := interpolateColor(a, b, 0.5); got != (color.RGBA{100, 50, 25, 255}) {
		t.Errorf("lerp = %v", got)
	}
}
