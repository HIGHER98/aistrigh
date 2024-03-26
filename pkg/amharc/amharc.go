package amharc

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"sort"

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
		//extractNotes()
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

func findStaff(bar gocv.Mat) []image.Rectangle {

	img := bar.Clone()
	gocv.CvtColor(img, &img, gocv.ColorBGRAToGray)

	// Edge detection
	matLines := gocv.NewMat()
	defer matLines.Close()
	gocv.Canny(img, &img, 50, 200)
	// func HoughLinesPWithParams(src Mat, lines *Mat, rho float32, theta float32, threshold int, minLineLength float32, maxLineGap float32)
	// func HoughLinesP(src Mat, lines *Mat, rho float32, theta float32, threshold int)
	gocv.HoughLinesPWithParams(img, &matLines, 1, math.Pi/180, 80, 200, 10)
	//  HoughLinesP( dst, lines, 1, CV_PI/180, 80, 30, 10 )

	lines := sortMatLines(matLines)
	fmt.Println("Number of lines found", matLines.Rows(), "lines", len(lines))
	lines.prettyPrint()
	if len(lines)%2 == 1 {
		fmt.Println("We've found an odd number of lines")
	}

	var rects []image.Rectangle
	for i := 0; i < len(lines)-1; i++ {
		//line := lines[i]

		if (lines[i].pt1.Y - lines[i+1].pt1.Y) <= 2 {
			// group found
			fmt.Println("group found ", "i+1", lines[i+1].pt1.Y, "i", lines[i].pt1.Y)
			fmt.Println("drawing")
			//for _, line := range lines {
			//colorR := uint8(rand.Intn(255))
			//colorG := uint8(rand.Intn(255))
			//colorB := uint8(rand.Intn(255))
			//		gocv.Line(&bar, line.pt1, line.pt2, color.RGBA{colorR, colorG, colorB, 0}, 1)

			rect := createRectangle(lines[i], lines[i+1])
			//gocv.Rectangle(&bar, rect, color.RGBA{colorR, colorG, colorB, 0}, 1)
			fmt.Println("rectangle is", rect)
			rects = append(rects, rect)
		}

	}
	return rects
}

// TODO: Backproject histogram

func createRectangle(line1, line2 line) image.Rectangle {
	// need to choose the longest line
	xLen := line1.pt2.X
	if line2.pt2.X > xLen {
		xLen = line2.pt2.X
	}
	// add padding
	minPt := image.Point{line1.pt1.X, line1.pt1.Y - 3}
	maxPt := image.Point{xLen, line2.pt2.Y + 3}
	fmt.Println("Creating rectangle with minPt", minPt.String(), "maxPt", maxPt.String())
	return image.Rectangle{minPt, maxPt}
}

type lines []line
type line struct {
	pt1    image.Point
	pt2    image.Point
	length float64
}

func (lines lines) prettyPrint() {
	for _, l := range lines {
		l.prettyPrint()
	}
}

func (l line) prettyPrint() {

	fmt.Printf("%s, %s\tLength=%f\n", l.pt1.String(), l.pt2.String(), l.length)
}

func sortMatLines(matLines gocv.Mat) lines {
	var l lines
	for i := 0; i < matLines.Rows(); i++ {
		pt1 := image.Pt(int(matLines.GetVeciAt(i, 0)[0]), int(matLines.GetVeciAt(i, 0)[1]))
		pt2 := image.Pt(int(matLines.GetVeciAt(i, 0)[2]), int(matLines.GetVeciAt(i, 0)[3]))
		lineLength := math.Sqrt(math.Pow(float64(pt2.Y-pt1.Y), 2) + math.Pow(float64(pt2.X-pt1.X), 2))

		l = append(l, line{pt1, pt2, lineLength})
	}

	sort.Slice(l, func(i, j int) bool {
		return l[i].pt1.Y > l[j].pt2.Y
	})
	return l
}

const CircleMinArea = 1
const CircleMaxArea = 30

type circle struct {
	center image.Point
	radius int
}

func findNotes(bar gocv.Mat) []circle {

	img := bar.Clone()
	defer img.Close()
	gocv.CvtColor(img, &img, gocv.ColorBGRAToGray)

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
		//rects = append(rects, rect)
		// TODO: appends to circles array
	}

	return circles

	/*
		window = gocv.NewWindow("findnotes")
		defer window.Close()
		window.MoveWindow(100, 100)
		window.ResizeWindow(1200, 1000)
		window.IMShow(bar)
		window.WaitKey(3000)

		return
	*/
}

func extractNotes(bar gocv.Mat) {
	// a single bar of music
	img := bar.Clone()
	gocv.CvtColor(img, &img, gocv.ColorBGRAToGray)

	// Edge detection - line
	matLines := gocv.NewMat()
	defer matLines.Close()
	gocv.Canny(img, &img, 50, 200)
	// func HoughLinesPWithParams(src Mat, lines *Mat, rho float32, theta float32, threshold int, minLineLength float32, maxLineGap float32)
	// func HoughLinesP(src Mat, lines *Mat, rho float32, theta float32, threshold int)
	gocv.HoughLinesPWithParams(img, &matLines, 1, math.Pi/180, 80, 80, 10)
	//  HoughLinesP( dst, lines, 1, CV_PI/180, 80, 30, 10 )

	fmt.Println(matLines.Cols())
	fmt.Println(matLines.Rows())
	for i := 0; i < matLines.Rows(); i++ {
		pt1 := image.Pt(int(matLines.GetVeciAt(i, 0)[0]), int(matLines.GetVeciAt(i, 0)[1]))
		pt2 := image.Pt(int(matLines.GetVeciAt(i, 0)[2]), int(matLines.GetVeciAt(i, 0)[3]))

		fmt.Println("point 1", pt1, "point 2", pt2)
		gocv.Line(&bar, pt1, pt2, color.RGBA{0, 255, 50, 0}, 1)
	}

	/*
		// Edge detection - circle
		circles := gocv.NewMat()
		defer circles.Close()
		//gocv.Canny(img, &img, 50, 200)

		//gocv.HoughCircles(img, &circles, gocv.HoughGradient, 1.5, float64(1))
		gocv.HoughCircles(img, &circles, gocv.HoughGradient, 20, 1)
		//func HoughCircles(src Mat, circles *Mat, method HoughMode, dp, minDist float64)
		fmt.Println("circlecols", circles.Cols())
		fmt.Println("circlesrows", circles.Rows())
		for i := 0; i < circles.Cols(); i++ {
			//center := image.Point{circles[i][0], circles[i]
			//fmt.Println(circles.GetVecfAt(i, 0))
			fmt.Println("circles.getVecfAt", i, "=", circles.GetVecfAt(0, i))

			center := image.Pt(int(circles.GetVecfAt(0, i)[0]), int(circles.GetVecfAt(0, i)[1]))
			radius := int(circles.GetVecfAt(0, i)[2])
			if radius > 10 {
				continue
			}

			fmt.Println("center", center, "radius", radius)
			//func Circle(img *Mat, center image.Point, radius int, c color.RGBA, thickness int)
			gocv.Circle(&bar, center, radius, color.RGBA{0, 255, 50, 0}, 1)
		}
	*/

}
