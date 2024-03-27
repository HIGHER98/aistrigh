package amharc

import (
	"image"
	"image/color"
	"math/rand"

	"gocv.io/x/gocv"
)

func findClef(img gocv.Mat) {
	//clefPath := "/home/higher/Projects/github/aistrigh/img/templates/treble-clef.png"
	clefPath := "/home/higher/Projects/github/aistrigh/img/templates/cleffy.png"
	templ := gocv.IMRead(clefPath, gocv.IMReadGrayScale)

	defer templ.Close()
	res := gocv.NewMat()
	defer res.Close()
	mask := gocv.NewMat()
	defer mask.Close()

	gocv.MatchTemplate(img, templ, &res, gocv.TmCcorrNormed, mask)
	_, maxVal, _, maxLoc := gocv.MinMaxLoc(res)
	if maxVal < 0.93 {
		// not within our confidence threshold
		return
	}

	colorR := uint8(rand.Intn(255))
	colorG := uint8(rand.Intn(255))
	colorB := uint8(rand.Intn(255))
	colour := color.RGBA{colorR, colorG, colorB, 0}
	rect := image.Rectangle{Min: maxLoc, Max: image.Point{maxLoc.X + templ.Size()[1], maxLoc.Y + templ.Size()[0]}}
	gocv.Rectangle(&img, rect, colour, 1)
}
