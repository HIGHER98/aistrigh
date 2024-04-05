package amharc

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"math/rand"

	"gocv.io/x/gocv"
)

// Find all treble clefs in the full sheet
func findClefs(img gocv.Mat) ([]image.Rectangle, error) {
	clefPath := "/home/higher/Projects/github/aistrigh/img/templates/cleffy.png"

	templ := gocv.IMRead(clefPath, gocv.IMReadGrayScale)
	defer templ.Close()

	gocv.Resize(templ, &templ, image.Point{12, 27}, 0, 0, gocv.InterpolationArea)
	rects, err := matchTemplateMultiScale(img, templ, 0.1, 0.1)
	if err != nil {
		return nil, err
	}

	colorR := uint8(rand.Intn(255))
	colorG := uint8(rand.Intn(255))
	colorB := uint8(rand.Intn(255))
	colour := color.RGBA{colorR, colorG, colorB, 0}

	for _, rect := range rects {
		gocv.Rectangle(&img, rect, colour, 3)
	}

	return rects, nil
}

// match a template through scaling the img down in step steps until it reaches end
func matchTemplateMultiScale(img gocv.Mat, templ gocv.Mat, end, step float64) ([]image.Rectangle, error) {
	if end >= 1 || end <= 0 {
		return nil, errors.New("end should be between 0 and 1")
	}
	if step >= 1 || step <= 0 {
		return nil, errors.New("step should be between 0 and 1")
	}

	//	gocv.Canny(templ, &templ, 50, 200)

	clone := img.Clone()
	defer clone.Close()

	mask := gocv.NewMat()
	defer mask.Close()

	res := gocv.NewMat()
	defer res.Close()

	edged := gocv.NewMat()
	defer edged.Close()

	window := gocv.NewWindow("matchTemplateMultiScale")
	defer window.Close()
	window.MoveWindow(100, 100)
	window.ResizeWindow(1200, 1000)

	var clefs []image.Rectangle

	for scale := 1.0; scale >= end; scale = scale - step {

		gocv.Resize(img, &clone, image.Point{int(float64(img.Cols()) * scale), int(float64(img.Rows()) * scale)}, 0, 0, gocv.InterpolationArea)
		if clone.Cols() <= templ.Cols() || clone.Rows() <= templ.Rows() {
			// our image must always be bigger than our template
			break
		}

		//		gocv.Canny(clone, &edged, 50, 200)

		gocv.MatchTemplate(clone, templ, &res, gocv.TmCcorrNormed, mask)

		thresh := float32(0.96)
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
			fmt.Println("No rects found at scale", scale, "Image dimensions:", clone.Size())
			continue
		}
		scoreThreshold := float32(0.9)
		nmsThreshold := float32(0.5)

		indices := gocv.NMSBoxes(rects, scores, scoreThreshold, nmsThreshold)
		fmt.Println("Found", len(indices), "indices after NMS")

		fmt.Println(rects)
		fmt.Println(indices)
		for _, index := range indices {
			// add this rectangle to a data structure containing this rectangle and the scale it was recorded at
			//[ scale => [rects[index], rects[index+1], ...] ]

			// add this rectangle at the original scale to return
			clefs = append(clefs, image.Rectangle{
				Min: image.Point{int(float64(rects[index].Min.X) / scale), int(float64(rects[index].Min.Y) / scale)},
				Max: image.Point{int(float64(rects[index].Max.X) / scale), int(float64(rects[index].Max.Y) / scale)},
			})
		}

		window.IMShow(clone)
		window.WaitKey(3000)
		// on first appearance of any clefs, we can return since the clefs should all be the same size
		return clefs, nil
	}

	if len(clefs) == 0 {
		return nil, errors.New("No treble clefs found")
	}

	return clefs, nil
}
