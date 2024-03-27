package amharc

import (
	"image"
	"image/color"

	"gocv.io/x/gocv"
)

const CircleMinArea = 1
const CircleMaxArea = 30

type circle struct {
	center image.Point
	radius int
}

func findNotes(bar gocv.Mat) []circle {

	img := bar.Clone()
	defer img.Close()
	//gocv.CvtColor(img, &img, gocv.ColorBGRAToGray)

	// mask = cv2.inRange(image_color, lower_bound, upper_bound)
	//mask := gocv.NewMat()
	gocv.InRangeWithScalar(img, gocv.NewScalar(0, 0, 0, 10), gocv.NewScalar(0, 255, 255, 195), &img)

	rectKernel := gocv.GetStructuringElement(gocv.MorphRect, image.Point{3, 3})

	// BorderConstant = 0
	gocv.DilateWithParams(img, &img, rectKernel, image.Point{-1, -1}, 1, 0, color.RGBA{0, 0, 0, 0})
	gocv.ErodeWithParams(img, &img, rectKernel, image.Point{-1, -1}, 1, 0)
	gocv.MorphologyEx(img, &img, gocv.MorphOpen, rectKernel)

	var circles []circle

	contours := gocv.FindContours(img, gocv.RetrievalExternal, gocv.ChainApproxSimple)
	for idx := 0; idx < contours.Size(); idx++ {
		area := gocv.ContourArea(contours.At(idx))
		if area < CircleMinArea || area > CircleMaxArea {
			continue
		}
		x, y, radius := gocv.MinEnclosingCircle(contours.At(idx))
		center := image.Pt(int(x), int(y))
		gocv.Circle(&bar, center, int(radius), color.RGBA{0, 255, 50, 0}, 1)
		circles = append(circles, circle{center, int(radius)})
	}

	return circles
}
