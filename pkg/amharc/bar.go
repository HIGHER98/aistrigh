package amharc

import (
	"fmt"
	"image"
	"image/color"

	"gocv.io/x/gocv"
)

const MinArea = 20000
const MaxArea = 100000
const XPadding = 5
const YPadding = 30

// extract all the bars into different regions represented by image.Rectangle's
// Resize the image, apply blur, threshold and bring back to original size, find contours, bound these, pad those, return 'em
func extractBars(img gocv.Mat, originalImg gocv.Mat) []image.Rectangle {

	var rects []image.Rectangle
	gocv.Resize(img, &img, image.Point{100, 200}, 0, 0, gocv.InterpolationArea)
	gocv.Blur(img, &img, image.Point{20, 2})

	gocv.Threshold(img, &img, 200, 255, gocv.ThresholdBinary)
	rectKernel := gocv.GetStructuringElement(gocv.MorphRect, image.Point{12, 6})
	gocv.ErodeWithParams(img, &img, rectKernel, image.Point{-1, -1}, 2, 0)

	// create a whitespce border around the image before finding contours, or we may highlight a massive area
	border := 10
	gocv.Resize(img, &img, image.Point{ImageCols - border*2, ImageRows - border*2}, 0, 0, gocv.InterpolationArea)
	gocv.CopyMakeBorder(img, &img, border, border, border, border, gocv.BorderConstant, color.RGBA{255, 255, 255, 0})

	// find bars and draw them
	contours := gocv.FindContours(img, gocv.RetrievalTree, gocv.ChainApproxNone)

	fmt.Println(contours.Size())
	x, y := getAverageContourSize(contours)
	fmt.Println("Average contour size: ", x, y)
	pts := gocv.NewMat()
	defer pts.Close()

	for idx := 0; idx < contours.Size(); idx++ {
		area := gocv.ContourArea(contours.At(idx))
		if area < MinArea || area > MaxArea {
			continue
		}
		//gocv.DrawContours(&originalImg, contours, idx, color.RGBA{255, 255, 20, 0}, 2)
		rect := gocv.BoundingRect(contours.At(idx))
		// Add padding to this rect
		rect.Min.X = rect.Min.X - XPadding
		rect.Min.Y = rect.Min.Y - YPadding
		rect.Max.X = rect.Max.X + XPadding
		rect.Max.Y = rect.Max.Y + YPadding
		//gocv.Rectangle(&originalImg, rect, color.RGBA{0, 234, 106, 0}, 2)
		//text := fmt.Sprintf("Size: %d", area)
		//gocv.PutText(&originalImg, text, rect.Min, gocv.FontHersheyPlain, 1, color.RGBA{200, 200, 100, 0}, 2)

		fmt.Println("Contour with size", area, "at position", rect.Min, "max", rect.Max)

		rects = append(rects, rect)
	}

	return rects
}

func getAverageContourSize(contours gocv.PointsVector) (avgX int, avgY int) {
	xTally := 0
	yTally := 0
	bars := 0
	for idx := 0; idx < contours.Size(); idx++ {
		area := gocv.ContourArea(contours.At(idx))
		if area < MinArea || area > MaxArea {
			continue
		}
		bars = bars + 1
		rect := gocv.BoundingRect(contours.At(idx))
		xTally = xTally + (rect.Max.X - rect.Min.X)
		yTally = yTally + (rect.Max.Y - rect.Min.Y)
	}
	fmt.Println("x avg", xTally/bars, "y avg", yTally/bars)
	return xTally / bars, yTally / bars
}
