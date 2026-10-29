package main

import "testing"

func TestIterations(t *testing.T) {
	if got := Iterations(0, 0, 100); got != 100 {
		t.Errorf("Ursprung gehört zur Menge, got %d", got)
	}
	if got := Iterations(2, 2, 100); got >= 100 {
		t.Errorf("2+2i divergiert, got %d", got)
	}
}

func TestComputeSize(t *testing.T) {
	g := Compute(8, 4, -2, 1, -1, 1, 50)
	if len(g) != 4 || len(g[0]) != 8 {
		t.Errorf("falsche Größe: %dx%d", len(g[0]), len(g))
	}
}
