package main

import "testing"

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
