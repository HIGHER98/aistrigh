package amharc

import (
	"fmt"
	"image"
	"math"
	"sort"

	"gocv.io/x/gocv"
)

func findStaff(bar gocv.Mat) staffLines {

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
	fmt.Println("Number of lines found", len(lines))
	//lines.prettyPrint()
	if len(lines)%2 == 1 {
		fmt.Println("We've found an odd number of lines")
	}

	var rects []image.Rectangle
	for i := 0; i < len(lines)-1; i++ {
		if (lines[i+1].pt1.Y - lines[i].pt1.Y) <= 2 {
			// group found
			rect := createRectangle(lines[i], lines[i+1])
			/*
				colorR := uint8(rand.Intn(255))
				colorG := uint8(rand.Intn(255))
				colorB := uint8(rand.Intn(255))
				//gocv.Line(&bar, line.pt1, line.pt2, color.RGBA{colorR, colorG, colorB, 0}, 1)
				gocv.Rectangle(&bar, rect, color.RGBA{colorR, colorG, colorB, 0}, 1)
				/*
					text := fmt.Sprintf("Rectangle: (%s, %s) (%s, %s)", rect.Min.X, rect.Min.Y, rect.Max.X, rect.Max.Y)
					gocv.PutText(&bar, text, rect.Min, gocv.FontHersheyPlain, 1, color.RGBA{colorR, colorG, colorB, 0}, 2)
			*/
			rects = append(rects, rect)
		}

	}
	fmt.Println("creating staffline...")

	sls := createStaffLines(rects)
	/*
		gocv.Rectangle(&bar, sls[0].rect, color.RGBA{255, 0, 0, 0}, 1)
		gocv.Rectangle(&bar, sls[1].rect, color.RGBA{255, 0, 255, 0}, 1)
		gocv.Rectangle(&bar, sls[2].rect, color.RGBA{0, 255, 0, 0}, 1)
		gocv.Rectangle(&bar, sls[3].rect, color.RGBA{0, 255, 0, 0}, 1)
		gocv.Rectangle(&bar, sls[4].rect, color.RGBA{255, 255, 0, 0}, 1)
	*/
	return sls
}

var notePositionLineMap = map[int]string{
	0: "f",
	1: "d",
	2: "b",
	3: "g",
	4: "e",
}

var notePositionSpaceMap = map[int]string{
	0: "e",
	1: "c",
	2: "a",
	3: "f",
}

func createStaffLines(rects []image.Rectangle) staffLines {
	var sls staffLines
	if len(rects)%5 == 0 {
		// TODO
	}

	// create staff representations for lines
	for i, rect := range rects {
		fmt.Println("i", i, "rect.Min.Y", rect.Min.Y, "rect.Max.Y", rect.Max.Y, "notePositionLineMap[i]", notePositionLineMap[i])
		sls = append(sls, staffLine{rect, notePositionLineMap[i]})
	}

	// create staff representations for clear line rects to represent the notes marked by the space between lines
	for i := 0; i < len(rects)-1; i++ {
		j := i + 1
		// if this ect and the next don't overlap, represenet that space as a staffLine (FACE)
		if rects[i].Intersect(rects[j]).Empty() {
			fmt.Println("appending...", image.Rectangle{Min: rects[i].Min, Max: image.Point{rects[j].Max.X, rects[j].Min.Y}})
			sls = append(sls, staffLine{image.Rectangle{Min: rects[i].Min,
				Max: image.Point{rects[j].Max.X, rects[j].Min.Y}}, notePositionSpaceMap[i]})

		} else {
			fmt.Println("Not appending. Intersection found between", rects[i], rects[j])
		}

	}

	for _, sl := range sls {
		sl.print()
	}
	return sls
}

// a region in the bar associated with a particulat pitch
type staffLine struct {
	rect  image.Rectangle
	pitch string
}

func (sl staffLine) print() {
	fmt.Println("Min", sl.rect.Min.String(), "Max", sl.rect.Max.String(), "pitch", sl.pitch)
}

type staffLines []staffLine

func (sls staffLines) contains(notePosition circle) string {
	for _, sl := range sls {
		if notePosition.center.In(sl.rect) {

			fmt.Printf("center: %s ", notePosition.center.String())
			sl.print()
			return sl.pitch
		}
	}
	return "x"

}

func createRectangle(line1, line2 line) image.Rectangle {
	// need to choose the longest line
	xLen := line1.pt2.X
	if line2.pt2.X > xLen {
		xLen = line2.pt2.X
	}
	// add padding
	minPt := image.Point{line1.pt1.X, line1.pt1.Y - 1}
	maxPt := image.Point{xLen, line2.pt2.Y + 1}
	return image.Rectangle{minPt, maxPt}
}

type lines []line

func (lines lines) prettyPrint() {
	for _, l := range lines {
		l.prettyPrint()
	}
}

type line struct {
	pt1    image.Point
	pt2    image.Point
	length float64
}

func createLine(pt1, pt2 image.Point) line {
	lineLength := math.Sqrt(math.Pow(float64(pt2.Y-pt1.Y), 2) + math.Pow(float64(pt2.X-pt1.X), 2))
	return line{pt1, pt2, lineLength}
}

func (l line) prettyPrint() {
	fmt.Printf("%s, %s\tLength=%f\n", l.pt1.String(), l.pt2.String(), l.length)
}

// returns lines in an asending Y order (i.e. moving down the image)
func sortMatLines(matLines gocv.Mat) lines {
	var l lines
	for i := 0; i < matLines.Rows(); i++ {
		pt1 := image.Pt(int(matLines.GetVeciAt(i, 0)[0]), int(matLines.GetVeciAt(i, 0)[1]))
		pt2 := image.Pt(int(matLines.GetVeciAt(i, 0)[2]), int(matLines.GetVeciAt(i, 0)[3]))
		line := createLine(pt1, pt2)

		l = append(l, line)
	}

	sort.Slice(l, func(i, j int) bool {
		return l[i].pt1.Y < l[j].pt2.Y
	})
	return l
}
