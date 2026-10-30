package main

import "log"

func main() {
	if err := runGUI(); err != nil {
		log.Fatal(err)
	}
}
