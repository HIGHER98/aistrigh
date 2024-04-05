package amharc

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"os"

	"gocv.io/x/gocv"
)

const ImageRows = 1100
const ImageCols = 850

func ReadSheet(filepath string) error {
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return errors.New(fmt.Sprintf("File '%s' not found", filepath))
	}

	img := gocv.IMRead(filepath, gocv.IMReadColor)
	defer img.Close()
	gocv.Resize(img, &img, image.Point{ImageCols, ImageRows}, 0, 0, gocv.InterpolationArea)
	gocv.CvtColor(img, &img, gocv.ColorBGRAToGray)

	bars, err := findClefs(img)
	if err != nil {
		return err
	}

	gocv.CvtColor(img, &img, gocv.ColorGrayToBGRA)
	// show our hard work....
	window := gocv.NewWindow(filepath)
	defer window.Close()
	window.MoveWindow(100, 100)
	window.ResizeWindow(1200, 1000)
	window.IMShow(img)
	window.WaitKey(3000)

	return nil
	//bars := extractBars(img)
	for _, bar := range bars {
		sls := findStaff(img.Region(bar))
		notePositions := findNotes(img.Region(bar))

		gocv.CvtColor(img, &img, gocv.ColorGrayToBGRA)
		// for debugging
		//notePositions.draw(img.Region(bar))
		//sls.draw(img.Region(bar))

		for _, notePosition := range notePositions {

			pitch := sls.contains(notePosition)

			fmt.Printf(" %s ", pitch)
			if pitch != UndefinedPitch {
				drawNote(img.Region(bar), notePosition, pitch)
			}
		}
		window.IMShow(img.Region(bar))
		window.WaitKey(3000)

		gocv.CvtColor(img, &img, gocv.ColorBGRAToGray)
	}
	return nil
}

func drawNote(img gocv.Mat, notePosition circle, pitch string) {
	gocv.Circle(&img, notePosition.center, notePosition.radius, color.RGBA{255, 100, 0, 0}, 1)
	text := fmt.Sprintf("%s", pitch)
	gocv.PutText(&img, text, notePosition.center, gocv.FontHersheyPlain, 1, color.RGBA{200, 10, 100, 0}, 1)
}
