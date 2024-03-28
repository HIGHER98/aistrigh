package amharc

import (
	"image"
	"image/color"

	"gocv.io/x/gocv"
)

const CircleMinArea = 1
const CircleMaxArea = 30

type circles []circle

type circle struct {
	center image.Point
	radius int
}

func (cs circles) draw(img gocv.Mat) {
	for _, c := range cs {
		c.draw(&img)
	}
}

func (c circle) draw(img *gocv.Mat) {
	gocv.Circle(img, c.center, c.radius, color.RGBA{0, 255, 50, 0}, 1)
}

func findNotes(bar gocv.Mat) circles {
	img := bar.Clone()
	defer img.Close()
	//gocv.CvtColor(img, &img, gocv.ColorBGRAToGray)

	gocv.InRangeWithScalar(img, gocv.NewScalar(0, 0, 0, 10), gocv.NewScalar(0, 255, 255, 195), &img)

	rectKernel := gocv.GetStructuringElement(gocv.MorphRect, image.Point{3, 3})

	// BorderConstant = 0
	gocv.DilateWithParams(img, &img, rectKernel, image.Point{-1, -1}, 1, 0, color.RGBA{0, 0, 0, 0})
	gocv.ErodeWithParams(img, &img, rectKernel, image.Point{-1, -1}, 1, 0)
	gocv.MorphologyEx(img, &img, gocv.MorphOpen, rectKernel)

	var circles circles

	contours := gocv.FindContours(img, gocv.RetrievalExternal, gocv.ChainApproxSimple)
	for idx := 0; idx < contours.Size(); idx++ {
		area := gocv.ContourArea(contours.At(idx))
		if area < CircleMinArea || area > CircleMaxArea {
			continue
		}
		x, y, radius := gocv.MinEnclosingCircle(contours.At(idx))
		center := image.Pt(int(x), int(y))
		circles = append(circles, circle{center, int(radius)})
	}

	return circles
}
