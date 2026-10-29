package main

// Iterations liefert die Anzahl der Iterationen, bis |z| > 2 für z = z²+c, maximal maxIter.
func Iterations(cr, ci float64, maxIter int) int {
	var zr, zi float64
	for i := 0; i < maxIter; i++ {
		zr2, zi2 := zr*zr, zi*zi
		if zr2+zi2 > 4 {
			return i
		}
		zi = 2*zr*zi + ci
		zr = zr2 - zi2 + cr
	}
	return maxIter
}

// Compute berechnet ein width×height-Raster der Iterationszahlen für den Ausschnitt [xmin,xmax]×[ymin,ymax].
func Compute(width, height int, xmin, xmax, ymin, ymax float64, maxIter int) [][]int {
	grid := make([][]int, height)
	for y := range grid {
		row := make([]int, width)
		ci := ymin + (ymax-ymin)*float64(y)/float64(height)
		for x := range row {
			cr := xmin + (xmax-xmin)*float64(x)/float64(width)
			row[x] = Iterations(cr, ci, maxIter)
		}
		grid[y] = row
	}
	return grid
}
