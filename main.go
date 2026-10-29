package main

import "log"

func main() {
	if err := runGUI(NewClassicPalette()); err != nil {
		log.Fatal(err)
	}
}
