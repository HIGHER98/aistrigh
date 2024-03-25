package amharc

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"gocv.io/x/gocv"
)

const ImageRows = 1100
const ImageCols = 850

func ReadSheet(filepath string) {
	img := gocv.IMRead(filepath, gocv.IMReadColor)
	defer img.Close()
	gocv.Resize(img, &img, image.Point{ImageCols, ImageRows}, 0, 0, gocv.InterpolationArea)
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
	for _, bar := range bars {

		fmt.Println(bar)
		window.IMShow(img.Region(bar))
		window.WaitKey(3000)
	}
}

const MinArea = 20000
const MaxArea = 100000
const XPadding = 5
const YPadding = 30

// Resize the image, apply blur, threshold and bring back to original size, find contours, bound these, pad those, return 'em
func extractBars(img gocv.Mat, originalImg gocv.Mat) []image.Rectangle {

	var rects []image.Rectangle
	gocv.Resize(img, &img, image.Point{100, 200}, 0, 0, gocv.InterpolationArea)
	gocv.Blur(img, &img, image.Point{20, 2})

	gocv.Threshold(img, &img, 200, 255, gocv.ThresholdBinary)
	rectKernel := gocv.GetStructuringElement(gocv.MorphRect, image.Point{12, 6})
	//rectKernel := gocv.GetStructuringElement(gocv.MorphRect, image.Point{13, 13})
	//gocv.Erode(img, &img, rectKernel)
	gocv.ErodeWithParams(img, &img, rectKernel, image.Point{-1, -1}, 2, 0)

	// create a whitespce border around the image before finding contours, or we may highlight a massive area
	border := 10
	gocv.Resize(img, &img, image.Point{ImageCols - border*2, ImageRows - border*2}, 0, 0, gocv.InterpolationArea)
	gocv.CopyMakeBorder(img, &img, border, border, border, border, gocv.BorderConstant, color.RGBA{255, 255, 255, 0})
	//return rects

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
		gocv.DrawContours(&originalImg, contours, idx, color.RGBA{255, 255, 20, 0}, 2)
		rect := gocv.BoundingRect(contours.At(idx))
		// Add padding to this rect
		rect.Min.X = rect.Min.X - XPadding
		rect.Min.Y = rect.Min.Y - YPadding
		rect.Max.X = rect.Max.X + XPadding
		rect.Max.Y = rect.Max.Y + YPadding
		gocv.Rectangle(&originalImg, rect, color.RGBA{0, 234, 106, 0}, 2)
		text := fmt.Sprintf("Size: %d", area)
		gocv.PutText(&originalImg, text, rect.Min, gocv.FontHersheyPlain, 1, color.RGBA{200, 200, 100, 0}, 2)

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

func edgeDetect(img gocv.Mat, originalImg gocv.Mat) {

	// Edge detection
	matLines := gocv.NewMat()
	defer matLines.Close()
	gocv.Canny(img, &img, 50, 200)
	gocv.HoughLinesP(img, &matLines, 1, math.Pi/180, 80)

	fmt.Println(matLines.Cols())
	fmt.Println(matLines.Rows())
	pv := gocv.NewPointVector()
	var rects []image.Rectangle

	for i := 0; i < matLines.Rows(); i++ {
		pt1 := image.Pt(int(matLines.GetVeciAt(i, 0)[0]), int(matLines.GetVeciAt(i, 0)[1]))
		pt2 := image.Pt(int(matLines.GetVeciAt(i, 0)[2]), int(matLines.GetVeciAt(i, 0)[3]))
		pv.Append(pt1)
		pv.Append(pt2)

		rects = append(rects, gocv.BoundingRect(pv))
		fmt.Println("point 1", pt1, "point 2", pt2)
		gocv.Line(&originalImg, pt1, pt2, color.RGBA{0, 255, 50, 0}, 1)
	}

	rects = gocv.GroupRectangles(rects, 1, 0.2)
	for _, r := range rects {
		gocv.Rectangle(&originalImg, r, color.RGBA{255, 0, 0, 0}, 1)
	}
}
