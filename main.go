package main

import (
	"aistrigh/pkg/amharc"
	"flag"
	"log"
)

var imgFilepath string

func init() {
	flag.StringVar(&imgFilepath, "i", "", "Filepath of image to process")
}
func main() {
	flag.Parse()
	if imgFilepath == "" {
		log.Fatalln("No image provided")
	}
	err := amharc.ReadSheet(imgFilepath)
	if err != nil {
		log.Fatalln(err)
	}

}
