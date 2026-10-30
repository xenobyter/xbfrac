package main

import (
	"image/color"
	"testing"
)

type stubPalette struct{}

func (stubPalette) Color(iter, maxIter int) color.Color {
	return color.RGBA{uint8(iter), uint8(maxIter), 0, 255}
}

func TestRenderMapsGridThroughPalette(t *testing.T) {
	img := Render([][]int{{1, 2, 3}, {4, 5, 6}}, 9, stubPalette{})
	if b := img.Bounds(); b.Dx() != 3 || b.Dy() != 2 {
		t.Fatalf("Größe %v", b)
	}
	if got := img.RGBAAt(2, 1); got != (color.RGBA{6, 9, 0, 255}) {
		t.Errorf("Pixel (2,1) = %v", got)
	}
}

func TestRenderPaletteChangesColors(t *testing.T) {
	grid := [][]int{{17}}
	classic := Render(grid, 100, NewClassicPalette())
	fire := Render(grid, 100, AvailablePalettes()[1].Palette)
	classicR, classicG, classicB, _ := classic.At(0, 0).RGBA()
	fireR, fireG, fireB, _ := fire.At(0, 0).RGBA()
	if classicR == fireR && classicG == fireG && classicB == fireB {
		t.Fatal("expected selected palettes to render different colors")
	}
}

func TestRenderEmptyGrid(t *testing.T) {
	if b := Render(nil, 10, stubPalette{}).Bounds(); !b.Empty() {
		t.Errorf("leeres Bild erwartet, got %v", b)
	}
}
