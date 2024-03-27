package amharc

import (
	"fmt"
	"image"
	"image/color"

	"gocv.io/x/gocv"
)

const ImageRows = 1100
const ImageCols = 850

func ReadSheet(filepath string) {
	img := gocv.IMRead(filepath, gocv.IMReadColor)
	defer img.Close()
	gocv.Resize(img, &img, image.Point{ImageCols, ImageRows}, 0, 0, gocv.InterpolationArea)
	//gocv.Blur(img, &img, image.Point{3, 3})
	//gocv.GaussianBlur(img, &img, image.Point{3, 3}, 1, 1, gocv.BorderDefault)
	gocv.CvtColor(img, &img, gocv.ColorBGRAToGray)
	originalImg := img.Clone() // contains a copy of the original image
	defer originalImg.Close()

	bars := extractBars(img, originalImg)
	originalImg.CopyTo(&img)

	// show our hard work....
	window := gocv.NewWindow(filepath)
	defer window.Close()
	window.MoveWindow(100, 100)
	window.ResizeWindow(1200, 1000)
	window.IMShow(img)
	window.WaitKey(3000)
	for _, bar := range bars {
		sls := findStaff(img.Region(bar))
		notePositions := findNotes(img.Region(bar))
		//	sls.draw(img.Region(bar))

		findClef(img.Region(bar))

		for _, notePosition := range notePositions {

			pitch := sls.contains(notePosition)

			fmt.Printf(" %s ", pitch)
			if pitch != "x" {
				drawNote(img.Region(bar), notePosition, pitch)
			}
		}
		fmt.Println()

		window.IMShow(img.Region(bar))
		window.WaitKey(3000)
	}
}

func drawNote(img gocv.Mat, notePosition circle, pitch string) {
	gocv.Circle(&img, notePosition.center, notePosition.radius, color.RGBA{255, 100, 0, 0}, 1)
	text := fmt.Sprintf("%s", pitch)
	gocv.PutText(&img, text, notePosition.center, gocv.FontHersheyPlain, 1, color.RGBA{200, 10, 100, 0}, 1)
}

// TODO: Backproject histogram
