package amharc

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"math"
	"math/rand"
	"sort"

	"gocv.io/x/gocv"
)

// Given a bar, this will create stafflines; a representational mapping of a space on the staff to a pitch
// EGBDF lines are found using Hough lines. A rectangle is constructed around these lines to denote the aforementioned pitches.
// The space in between the found rectangles is used to find the FACE lines.
// These set of rectangles are mapped above and below as 'Ghost lines' to represent even lower and higher pitches
func FindStaff(bar gocv.Mat) staffLines {

	img := bar.Clone()

	// Edge detection
	matLines := gocv.NewMat()
	defer matLines.Close()
	gocv.Canny(img, &img, 50, 200)
	gocv.HoughLinesPWithParams(img, &matLines, 1, math.Pi/180, 80, 200, 10)
	// These lines should contain EGBDF
	lines := sortMatLines(matLines)
	fmt.Println("Number of lines found", len(lines))

	var rects []image.Rectangle
	for i := 0; i < len(lines)-1; i++ {
		if (lines[i+1].pt1.Y - lines[i].pt1.Y) <= 2 {
			// This group of lines, as a rectangle, represents a single pitch
			rect := createRectangle(lines[i], lines[i+1])
			rects = append(rects, rect)
		}

	}
	rects, _ = stripOutliar(rects, 5)
	sls := createStaffLines(rects)
	return sls
}

const UndefinedPitch = "x"

var notes = map[int]string{
	0:  "a2",
	1:  "b2",
	2:  "c2",
	3:  "d2",
	4:  "e2",
	5:  "f2",
	6:  "g2",
	7:  "a3",
	8:  "b3",
	9:  "c3",
	10: "d3",
	11: "e3",
	12: "f3",
	13: "g3",
	14: "a4",
	15: "b4",
	16: "c4",
	17: "d4",
	18: "e4",
	19: "f4",
	20: "g4",
	21: "a5",
	22: "b5",
	23: "c5",
	24: "d5",
	25: "e5",
	26: "f5",
	27: "g5",
	28: "a6",
	29: "b6",
	30: "c6",
	31: "d6",
	32: "e6",
	33: "f6",
	34: "g6",
	35: "a7",
	36: "b7",
	37: "c7",
	38: "d7",
	39: "e7",
	40: "f7",
	41: "g7",
}

func createStaffLines(rects []image.Rectangle) staffLines {
	var sls staffLines

	// how much space we are typically allocating for a pitch on a line/on space between lines
	var avgLineAlloc int
	var avgSpaceAlloc int

	var lineAllocSum int
	var spaceAllocSum int

	// create staff representations for lines EGBDF
	for i, rect := range rects {
		// 19 is e4
		sls = append(sls, staffLine{rect, notes[19-(i*2)]})
		lineAllocSum += rect.Max.Y - rect.Min.Y
	}
	avgLineAlloc = lineAllocSum / 5

	// create staff representations for clear line rects to represent the notes marked by the space between lines FACE
	for i := 0; i < len(rects)-1; i++ {
		j := i + 1
		// if this rect and the next don't overlap, represent that space as a staffLine (Notes: FACE)
		if rects[i].Intersect(rects[j]).Empty() {
			rect := image.Rectangle{Min: image.Point{rects[i].Min.X, rects[i].Max.Y},
				Max: image.Point{rects[j].Max.X, rects[j].Min.Y}}
			sls = append(sls, staffLine{rect, notes[18-(i*2)]})
			spaceAllocSum += rect.Max.Y - rect.Min.Y
		} else {
			fmt.Println("Not appending. Intersection found between", rects[i], rects[j])
		}
	}
	avgSpaceAlloc = spaceAllocSum / 4

	fmt.Println("Average line alloc:", avgLineAlloc, "Average space alloc:", avgSpaceAlloc)

	sls = createGhostLines(sls, avgLineAlloc, avgSpaceAlloc)
	sls.sort()
	sls.print()
	return sls
}

// TODO: Rename createLedgerLines
// create a ghost staff above and below the "main" staff to represent even lower and higher pitches
func createGhostLines(sls staffLines, lineAlloc, spaceAlloc int) staffLines {
	// sort ascending, sls[0] = 'f4'
	// sls[len(sls)-1] = 'e3'
	sls.sort()
	if sls.isEmpty() {
		// TODO: return nil, err
		return nil
	}

	var r image.Rectangle

	fmt.Println("lineAlloc", lineAlloc, "spaceAlloc", spaceAlloc)

	topLine := sls[0]             // f4
	bottomLine := sls[len(sls)-1] // e3

	l := len(sls)

	// going down
	for i := l - 2; i >= 0; i-- {
		if i%2 == 1 {
			// This is a space
			distanceToNextLine := sls[i].rect.Max.Y - sls[i].rect.Min.Y
			distanceToBottomLine := bottomLine.rect.Min.Y - sls[i].rect.Max.Y

			r = image.Rectangle{
				Min: image.Point{X: topLine.rect.Min.X, Y: bottomLine.rect.Max.Y + distanceToBottomLine},
				Max: image.Point{X: topLine.rect.Max.X, Y: bottomLine.rect.Max.Y + distanceToBottomLine + distanceToNextLine},
			}
		} else {
			// This is a line
			distanceToNextLine := sls[i].rect.Max.Y - sls[i].rect.Min.Y
			distanceToBottomLine := bottomLine.rect.Min.Y - sls[i].rect.Max.Y

			r = image.Rectangle{
				Min: image.Point{X: topLine.rect.Min.X, Y: bottomLine.rect.Max.Y + distanceToBottomLine},
				Max: image.Point{X: topLine.rect.Max.X, Y: bottomLine.rect.Max.Y + distanceToBottomLine + distanceToNextLine},
			}

		}

		sls = append(sls, staffLine{r, notes[3+i]})
	}

	// going up
	for i := 1; i < l; i++ {
		j := i + 1

		if i%2 == 1 {
			// This is a space
			distanceToNextLine := sls[j].rect.Min.Y - sls[i].rect.Min.Y
			distanceToTopLine := sls[i].rect.Min.Y - topLine.rect.Max.Y

			r = image.Rectangle{
				Min: image.Point{X: topLine.rect.Min.X, Y: topLine.rect.Min.Y - distanceToTopLine - distanceToNextLine},
				Max: image.Point{X: topLine.rect.Max.X, Y: topLine.rect.Min.Y - distanceToTopLine},
			}
		} else {
			// This is a line
			distanceToNextLine := sls[i].rect.Max.Y - sls[i].rect.Min.Y
			distanceToTopLine := sls[i].rect.Min.Y - topLine.rect.Max.Y

			r = image.Rectangle{
				Min: image.Point{X: topLine.rect.Min.X, Y: topLine.rect.Min.Y - distanceToTopLine - distanceToNextLine},
				Max: image.Point{X: topLine.rect.Max.X, Y: topLine.rect.Min.Y - distanceToTopLine},
			}

		}

		sls = append(sls, staffLine{r, notes[19+i]})
	}

	/* This works with the averages
	// d3, b3, g3, e2, ...
	for i := 0; i < 8; i += 2 {
		r = image.Rectangle{
			Min: image.Point{X: sls[len(sls)-1].rect.Min.X, Y: sls[len(sls)-1].rect.Max.Y},
			Max: image.Point{X: sls[len(sls)-1].rect.Max.X, Y: sls[len(sls)-1].rect.Max.Y + spaceAlloc},
		}

		sls = append(sls, staffLine{r, notes[10-i]})

		r = image.Rectangle{
			Min: image.Point{X: sls[len(sls)-1].rect.Min.X, Y: sls[len(sls)-1].rect.Max.Y},
			Max: image.Point{X: sls[len(sls)-1].rect.Max.X, Y: sls[len(sls)-1].rect.Max.Y + lineAlloc},
		}

		sls = append(sls, staffLine{r, notes[10-i-1]})
	}
	*/
	/*
		// c3, a3, f2, d2, ...
		for i := 1; i < 6; i += 2 {
			r = image.Rectangle{
				Min: image.Point{X: sls[len(sls)-1].rect.Min.X, Y: bottomLine.Max.Y + spaceAlloc + lineAlloc*i},
				Max: image.Point{X: sls[len(sls)-1].rect.Max.X, Y: bottomLine.Max.Y + spaceAlloc + lineAlloc*i + lineAlloc},
			}

			sls = append(sls, staffLine{r, notes[10-i]})
		}
	*/

	return sls
	///////

	/*
		for i := 0; i <= 7; i++ {
			if i%2 == 1 {
				r = image.Rectangle{
					Min: image.Point{X: sls[len(sls)-1].rect.Min.X, Y: sls[len(sls)-1].rect.Max.Y},
					Max: image.Point{X: sls[len(sls)-1].rect.Max.X, Y: sls[len(sls)-1].rect.Max.Y + spaceAlloc},
				}
			} else {
				r = image.Rectangle{
					Min: image.Point{X: sls[len(sls)-1].rect.Min.X, Y: sls[len(sls)-1].rect.Max.Y},
					Max: image.Point{X: sls[len(sls)-1].rect.Max.X, Y: sls[len(sls)-1].rect.Max.Y + lineAlloc},
				}
			}
			sls = append(sls, staffLine{r, notes[10-i]})
		}

		for i := 1; i < 9; i++ {
			r = image.Rectangle{
				Min: image.Point{X: sls[0].rect.Min.X, Y: sls[0].rect.Min.Y - space},
				Max: image.Point{X: sls[0].rect.Max.X, Y: sls[0].rect.Min.Y},
			}

			r = image.Rectangle{
				Min: image.Point{X: sls[0].rect.Min.X, Y: sls[0].rect.Min.Y - space - line},
				Max: image.Point{X: sls[0].rect.Max.X, Y: sls[0].rect.Min.Y - space},
			}

			r = image.Rectangle{
				Min: image.Point{X: sls[0].rect.Min.X, Y: sls[0].rect.Min.Y - space - line - space},
				Max: image.Point{X: sls[0].rect.Max.X, Y: sls[0].rect.Min.Y - space - line},
			}

			sls = append(sls, staffLine{r, notes[26+i]})
		}

		for i := 1; i < 9; i++ {
			r = image.Rectangle{
				Min: image.Point{X: sls[0].rect.Min.X, Y: (sls[0].rect.Min.Y - (sls[i].rect.Min.Y - sls[0].rect.Min.Y))},
				Max: image.Point{X: sls[0].rect.Max.X, Y: (sls[0].rect.Max.Y - (sls[i].rect.Min.Y - sls[0].rect.Min.Y))},
			}
			sls = append(sls, staffLine{r, notes[26+i]})
		}

	*/
	return sls
}

// sorts stafflines ascending, i.e. moving down the sheet, e.g. g7 -> f7 -> ... -> e4
func (sls staffLines) sort() {
	sort.Slice(sls, func(i, j int) bool {
		return sls[i].rect.Min.Y < sls[j].rect.Min.Y
	})
}

func (sls staffLines) isEmpty() bool {
	return len(sls) == 0
}

func (sls staffLines) draw(img gocv.Mat) {
	for _, sl := range sls {
		sl.draw(img)
	}
}

func (sls staffLines) print() {
	for _, sl := range sls {
		sl.print()
	}
}

// Strip the biggest outliar(s) of the y-sorted rectangles based on their y-values, leaving only `num` closest rects
func stripOutliar(rects []image.Rectangle, num int) ([]image.Rectangle, error) {

	if len(rects) < num {
		return nil, errors.New("We want a range larger than the number of available points")
	}

	p1 := 0
	p2 := num - 1
	diff := math.MaxInt
	start := 0
	end := num - 1

	for p2 < len(rects) {
		if rects[p2].Min.Y-rects[p1].Min.Y < diff {
			diff = rects[p2].Min.Y - rects[p1].Min.Y
			start = p1
			end = p2
		}
		p1++
		p2++
	}

	result := rects[start : end+1]
	return result, nil
}

// a region in the bar associated with a particulat pitch
type staffLine struct {
	rect  image.Rectangle
	pitch string
}

func (sl staffLine) print() {
	fmt.Println("--- Min", sl.rect.Min.String(), "Max", sl.rect.Max.String(), "pitch", sl.pitch)
}

func (sl staffLine) draw(img gocv.Mat) {
	colorR := uint8(rand.Intn(255))
	colorG := uint8(rand.Intn(255))
	colorB := uint8(rand.Intn(255))
	//gocv.Line(&bar, line.pt1, line.pt2, color.RGBA{colorR, colorG, colorB, 0}, 1)

	text := fmt.Sprintf("%s", sl.pitch)
	colour := color.RGBA{colorR, colorG, colorB, 0}
	gocv.PutText(&img, text, sl.rect.Min, gocv.FontHersheyPlain, 1, colour, 1)
	gocv.Rectangle(&img, sl.rect, colour, 1)
}

type staffLines []staffLine

func (sls staffLines) contains(notePosition circle) string {
	for _, sl := range sls {
		if notePosition.center.In(sl.rect) {
			return sl.pitch
		}
	}
	return UndefinedPitch
}

func createRectangle(line1, line2 line) image.Rectangle {
	// need to choose the longest line
	xLen := line1.pt2.X
	if line2.pt2.X > xLen {
		xLen = line2.pt2.X
	}
	// add padding
	//minPt := image.Point{line1.pt1.X, line1.pt1.Y - 1}
	//maxPt := image.Point{xLen, line2.pt2.Y + 1}
	minPt := image.Point{0, line1.pt1.Y}
	maxPt := image.Point{1000, line2.pt2.Y}
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
