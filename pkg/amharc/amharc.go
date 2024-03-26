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
	originalImg := img.Clone() // contains a copy of the original image
	defer originalImg.Close()

	gocv.CvtColor(img, &img, gocv.ColorBGRAToGray)
	bars := extractBars(img, originalImg)
	originalImg.CopyTo(&img)

	// show our hard work....
	window := gocv.NewWindow(filepath)
	defer window.Close()
	window.MoveWindow(100, 100)
	window.ResizeWindow(1200, 1000)
	window.IMShow(img)
	window.WaitKey(3000)
	barImg := gocv.NewMat()
	defer barImg.Close()
	for i, bar := range bars {
		staffRects := findStaff(img.Region(bar))
		notePositions := findNotes(img.Region(bar))
		fmt.Println("found these notes in bar", i, notePositions)

		for _, notePosition := range notePositions {

			for i, rect := range staffRects {
				//fmt.Println("Checking if note is in rect", notePosition.center, rect)
				if notePosition.center.In(rect) {
					note := "x"
					switch i {
					case 0:
						note = "f"
					case 1:
						note = "d"
					case 2:
						note = "b"
					case 3:
						note = "g"
					case 4:
						note = "e"
					default:
						note = "x"

					}
					fmt.Printf(" %s ", note)
					drawCircle(img.Region(bar), notePosition)
				}

			}
		}

		window.IMShow(img.Region(bar))
		window.WaitKey(3000)
	}
}

func drawCircle(img gocv.Mat, notePosition circle) {

	gocv.Circle(&img, notePosition.center, notePosition.radius, color.RGBA{255, 0, 0, 0}, 1)
}

// TODO: Backproject histogram
