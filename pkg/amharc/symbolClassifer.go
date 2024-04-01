package amharc

import (
	"fmt"
	"image"
	"image/color"
	"math/rand"

	"gocv.io/x/gocv"
)

func findClef(img gocv.Mat) {
	clefPath := "/home/higher/Projects/github/aistrigh/img/templates/cleffy.png"

	templ := gocv.IMRead(clefPath, gocv.IMReadGrayScale)
	defer templ.Close()

	res := gocv.NewMat()
	defer res.Close()
	mask := gocv.NewMat()
	defer mask.Close()
	gocv.MatchTemplate(img, templ, &res, gocv.TmCcorrNormed, mask)

	thresh := float32(0.93)

	colorR := uint8(rand.Intn(255))
	colorG := uint8(rand.Intn(255))
	colorB := uint8(rand.Intn(255))
	colour := color.RGBA{colorR, colorG, colorB, 0}

	var rects []image.Rectangle
	var scores []float32

	for c := 0; c < res.Cols(); c++ {
		for r := 0; r < res.Rows(); r++ {
			if res.GetFloatAt(r, c) > thresh {
				rect := image.Rectangle{Min: image.Point{c, r}, Max: image.Point{c + templ.Size()[1], r + templ.Size()[0]}}
				rects = append(rects, rect)
				scores = append(scores, res.GetFloatAt(r, c))
			}
		}
	}
	if len(rects) == 0 {
		return
	}
	scoreThreshold := float32(0.9)
	nmsThreshold := float32(0.5)

	indices := gocv.NMSBoxes(rects, scores, scoreThreshold, nmsThreshold)
	fmt.Println(len(rects))
	fmt.Println(rects)
	fmt.Println(indices)
	for _, index := range indices {
		gocv.Rectangle(&img, rects[index], colour, 3)
	}
}
