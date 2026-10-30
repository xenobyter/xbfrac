package main

import (
	"math"
	"testing"

	"fyne.io/fyne/v2"
)

func TestIterations(t *testing.T) {
	if got := Iterations(complex(0, 0), 100); got != 100 {
		t.Errorf("Ursprung gehört zur Menge, got %d", got)
	}
	if got := Iterations(complex(2, 2), 100); got >= 100 {
		t.Errorf("2+2i divergiert, got %d", got)
	}
}

func TestComputeSize(t *testing.T) {
	grid, err := Compute(8, 4, -2, 1, -1, 1, 50)
	if err != nil {
		t.Fatal(err)
	}
	if len(grid) != 4 || len(grid[0]) != 8 {
		t.Errorf("falsche Größe: %dx%d", len(grid[0]), len(grid))
	}
}

func TestComputeIncludesBounds(t *testing.T) {
	grid, err := Compute(2, 2, 0, 2, 0, 2, 5)
	if err != nil {
		t.Fatal(err)
	}
	want := [][]int{{5, 2}, {2, 1}}
	for y := range want {
		for x, expected := range want[y] {
			if got := grid[y][x]; got != expected {
				t.Errorf("grid[%d][%d] = %d, want %d", y, x, got, expected)
			}
		}
	}
}

func TestComputeRejectsNegativeDimensions(t *testing.T) {
	for _, dimensions := range [][2]int{{-1, 1}, {1, -1}} {
		if _, err := Compute(dimensions[0], dimensions[1], -2, 1, -1, 1, 50); err == nil {
			t.Errorf("Compute(%d, %d, ...) returned no error", dimensions[0], dimensions[1])
		}
	}
}

func TestMapTapToComplex(t *testing.T) {
	bounds := params{xmin: -2.2, xmax: 1, ymin: -1.2, ymax: 1.2}
	imageSize := fyne.NewSize(imgWidth, imgHeight)
	tests := []struct {
		name       string
		position   fyne.Position
		widgetSize fyne.Size
		want       complex128
		wantOK     bool
	}{
		{
			name:       "center",
			position:   fyne.NewPos(450, 300),
			widgetSize: imageSize,
			want:       complex(-0.6, 0),
			wantOK:     true,
		},
		{
			name:       "top left maps to xmin ymin",
			position:   fyne.NewPos(0, 0),
			widgetSize: imageSize,
			want:       complex(bounds.xmin, bounds.ymin),
			wantOK:     true,
		},
		{
			name:       "bottom right maps to xmax ymax",
			position:   fyne.NewPos(900, 600),
			widgetSize: imageSize,
			want:       complex(bounds.xmax, bounds.ymax),
			wantOK:     true,
		},
		{
			name:       "letterbox margin",
			position:   fyne.NewPos(300, 99),
			widgetSize: fyne.NewSize(600, 600),
			wantOK:     false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, ok := mapTapToComplex(test.position, test.widgetSize, imageSize, bounds)
			if ok != test.wantOK {
				t.Fatalf("mapTapToComplex() ok = %v, want %v", ok, test.wantOK)
			}
			if !ok {
				return
			}
			if math.Abs(real(got)-real(test.want)) > 1e-9 || math.Abs(imag(got)-imag(test.want)) > 1e-9 {
				t.Errorf("mapTapToComplex() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestMapDragToBounds(t *testing.T) {
	tests := []struct {
		name       string
		start      fyne.Position
		end        fyne.Position
		widgetSize fyne.Size
		imageSize  fyne.Size
		bounds     params
		want       params
		wantOK     bool
	}{
		{
			name:       "reverse drag",
			start:      fyne.NewPos(675, 450),
			end:        fyne.NewPos(225, 150),
			widgetSize: fyne.NewSize(imgWidth, imgHeight),
			imageSize:  fyne.NewSize(imgWidth, imgHeight),
			bounds:     params{xmin: -2.2, xmax: 1, ymin: -1.2, ymax: 1.2, maxIter: 200},
			want:       params{xmin: -1.4, xmax: 0.2, ymin: -0.6, ymax: 0.6, maxIter: 200},
			wantOK:     true,
		},
		{
			name:       "letterbox offset and clamp",
			start:      fyne.NewPos(150, 150),
			end:        fyne.NewPos(700, 600),
			widgetSize: fyne.NewSize(600, 600),
			imageSize:  fyne.NewSize(900, 600),
			bounds:     params{xmin: 0, xmax: 8, ymin: -4, ymax: 4, maxIter: 100},
			want:       params{xmin: 2, xmax: 8, ymin: -3, ymax: 4, maxIter: 100},
			wantOK:     true,
		},
		{
			name:       "start in letterbox",
			start:      fyne.NewPos(150, 50),
			end:        fyne.NewPos(450, 450),
			widgetSize: fyne.NewSize(600, 600),
			imageSize:  fyne.NewSize(900, 600),
			bounds:     params{xmin: 0, xmax: 8, ymin: -4, ymax: 4},
		},
		{
			name:       "selection too small",
			start:      fyne.NewPos(10, 10),
			end:        fyne.NewPos(13, 13),
			widgetSize: fyne.NewSize(imgWidth, imgHeight),
			imageSize:  fyne.NewSize(imgWidth, imgHeight),
			bounds:     params{xmin: 0, xmax: 8, ymin: -4, ymax: 4},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, ok := mapDragToBounds(test.start, test.end, test.widgetSize, test.imageSize, test.bounds)
			if ok != test.wantOK {
				t.Fatalf("mapDragToBounds() ok = %v, want %v", ok, test.wantOK)
			}
			if !ok {
				return
			}
			for name, values := range map[string][2]float64{
				"xmin": {got.xmin, test.want.xmin},
				"xmax": {got.xmax, test.want.xmax},
				"ymin": {got.ymin, test.want.ymin},
				"ymax": {got.ymax, test.want.ymax},
			} {
				if math.Abs(values[0]-values[1]) > 1e-9 {
					t.Errorf("%s = %v, want %v", name, values[0], values[1])
				}
			}
			if got.maxIter != test.want.maxIter {
				t.Errorf("maxIter = %d, want %d", got.maxIter, test.want.maxIter)
			}
		})
	}
}

func TestZoomIn(t *testing.T) {
	bounds := params{xmin: -2.2, xmax: 1, ymin: -1.2, ymax: 1.2, maxIter: 200}
	got, err := zoomIn(bounds, complex(-0.5, 0.25))
	if err != nil {
		t.Fatal(err)
	}
	for name, values := range map[string][2]float64{
		"xmin": {got.xmin, -1.3},
		"xmax": {got.xmax, 0.3},
		"ymin": {got.ymin, -0.35},
		"ymax": {got.ymax, 0.85},
	} {
		if math.Abs(values[0]-values[1]) > 1e-9 {
			t.Errorf("%s = %v, want %v", name, values[0], values[1])
		}
	}
	if got.maxIter != bounds.maxIter {
		t.Errorf("maxIter = %d, want %d", got.maxIter, bounds.maxIter)
	}
}

func TestZoomInRejectsPrecisionExhaustion(t *testing.T) {
	bounds := params{xmin: 1e16, xmax: 1e16 + 2, ymin: 0, ymax: 2}
	if _, err := zoomIn(bounds, complex(1e16, 1)); err == nil {
		t.Fatal("zoomIn() returned no error when bounds collapse at float64 precision")
	}
}

func TestZoomOut(t *testing.T) {
	bounds := params{xmin: -2.2, xmax: 1, ymin: -1.2, ymax: 1.2, maxIter: 200}
	got, err := zoomOut(bounds, complex(-0.5, 0.25))
	if err != nil {
		t.Fatal(err)
	}
	for name, values := range map[string][2]float64{
		"xmin": {got.xmin, -3.7},
		"xmax": {got.xmax, 2.7},
		"ymin": {got.ymin, -2.15},
		"ymax": {got.ymax, 2.65},
	} {
		if math.Abs(values[0]-values[1]) > 1e-9 {
			t.Errorf("%s = %v, want %v", name, values[0], values[1])
		}
	}
	if got.maxIter != bounds.maxIter {
		t.Errorf("maxIter = %d, want %d", got.maxIter, bounds.maxIter)
	}
}
