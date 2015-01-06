// Package main provides ...
package main

import (
	"os"
)

func main() {
	homePath := os.Getenv("HOME")

	// Text size
	dpi := 96
	text_width := 5 * (dpi / 96)

	monitor := os.Args[0]
	location := 12886294

	tags := [5]rune{'x', 'x', 'x', 'x', 'x'}

}
