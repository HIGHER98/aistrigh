package main

import (
	"aistrigh/pkg/amharc"
	"os"
)

func main() {
	filePath := os.Args[1]

	amharc.ReadSheet(filePath)
}
