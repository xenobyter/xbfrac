// Package main stellt einen interaktiven Mandelbrot-Viewer bereit.
package main

import "log"

func main() {
	if err := runGUI(); err != nil {
		log.Fatal(err)
	}
}
