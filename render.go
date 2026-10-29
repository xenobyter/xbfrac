package main

import "image"

// Render färbt ein Iterationsraster mit der Palette ein.
func Render(grid [][]int, maxIter int, p Palette) *image.RGBA {
	h := len(grid)
	w := 0
	if h > 0 {
		w = len(grid[0])
	}
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y, row := range grid {
		for x, n := range row {
			img.Set(x, y, p.Color(n, maxIter))
		}
	}
	return img
}
