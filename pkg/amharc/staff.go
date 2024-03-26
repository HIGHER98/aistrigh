package amharc

import (
	"fmt"
	"image"
	"math"
	"sort"

	"gocv.io/x/gocv"
)

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
		if (lines[i].pt1.Y - lines[i+1].pt1.Y) <= 2 {
			// group found
			fmt.Println("group found ", "i+1", lines[i+1].pt1.Y, "i", lines[i].pt1.Y)
			fmt.Println("drawing")
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
