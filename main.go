package main

import (
	"fmt"
	"strings"
)

func main() {
	const chars = " .:-=+*#%@"
	grid := Compute(78, 24, -2.2, 1.0, -1.2, 1.2, 50)
	for _, row := range grid {
		var sb strings.Builder
		for _, n := range row {
			sb.WriteByte(chars[n*(len(chars)-1)/50])
		}
		fmt.Println(sb.String())
	}
}
