package main

import "errors"

// Iterations liefert die Anzahl der Iterationen, bis |z| > 2 für z = z²+c, maximal maxIter.
func Iterations(c complex128, maxIter int) int {
	var z complex128
	for i := 0; i < maxIter; i++ {
		real, imag := real(z), imag(z)
		if real*real+imag*imag > 4 {
			return i
		}
		z = z*z + c
	}
	return maxIter
}

// Compute berechnet ein width×height-Raster für den inklusiven Ausschnitt [xmin,xmax]×[ymin,ymax].
// Breite und Höhe müssen mindestens 2 sein.
func Compute(width, height int, xmin, xmax, ymin, ymax float64, maxIter int) ([][]int, error) {
	if width < 2 || height < 2 {
		return nil, errors.New("Breite und Höhe müssen mindestens 2 sein")
	}

	grid := make([][]int, height)
	for y := range grid {
		row := make([]int, width)
		ci := ymin + (ymax-ymin)*float64(y)/float64(height-1)
		for x := range row {
			cr := xmin + (xmax-xmin)*float64(x)/float64(width-1)
			row[x] = Iterations(complex(cr, ci), maxIter)
		}
		grid[y] = row
	}
	return grid, nil
}
