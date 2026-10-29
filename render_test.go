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

func TestRenderEmptyGrid(t *testing.T) {
	if b := Render(nil, 10, stubPalette{}).Bounds(); !b.Empty() {
		t.Errorf("leeres Bild erwartet, got %v", b)
	}
}
