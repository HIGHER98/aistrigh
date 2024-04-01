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
func findStaff(bar gocv.Mat) staffLines {

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
	0:  "a3",
	1:  "b3",
	2:  "c3",
	3:  "d3",
	4:  "e3",
	5:  "f3",
	6:  "g3",
	7:  "a4",
	8:  "b4",
	9:  "c4",
	10: "d4",
	11: "e4",
	12: "f4",
	13: "g4",
	14: "a5",
	15: "b5",
	16: "c5",
	17: "d5",
	18: "e5",
	19: "f5",
	20: "g5",
	21: "a6",
	22: "b6",
	23: "c6",
	24: "d6",
	25: "e6",
	26: "f6",
	27: "g6",
	28: "a7",
	29: "b7",
	30: "c7",
	31: "d7",
	32: "e7",
	33: "f7",
	34: "g7",
	35: "a8",
	36: "b8",
	37: "c8",
	38: "d8",
	39: "e8",
	40: "f8",
	41: "g8",
}

func createStaffLines(rects []image.Rectangle) staffLines {
	var sls staffLines

	// create staff representations for lines EGBDF
	for i, rect := range rects {
		// 19 is e5
		sls = append(sls, staffLine{rect, notes[19-(i*2)]})
	}

	// create staff representations for clear line rects to represent the notes marked by the space between lines FACE
	for i := 0; i < len(rects)-1; i++ {
		j := i + 1
		// if this rect and the next don't overlap, represenet that space as a staffLine (Notes: FACE)
		if rects[i].Intersect(rects[j]).Empty() {
			sls = append(sls, staffLine{image.Rectangle{Min: image.Point{rects[i].Min.X, rects[i].Max.Y},
				Max: image.Point{rects[j].Max.X, rects[j].Min.Y}},
				notes[18-(i*2)]})
		} else {
			fmt.Println("Not appending. Intersection found between", rects[i], rects[j])
		}
	}

	sls = createGhostLines(sls)
	sls.sort()
	sls.print()
	return sls
}

// create a ghost staff above and below the "main" staff to represent even lower and higher pitches
func createGhostLines(sls staffLines) staffLines {
	// sort ascending, sls[0] = 'f'
	sls.sort()
	if sls.isEmpty() {
		// TODO: return nil, err
		return nil
	}

	bottomLine := sls[len(sls)-1] // has the highest Y-value. i.e. furthest down on the image
	// go 8 notes down
	notesDown := 9
	for i := 0; i < notesDown; i++ {
		r := image.Rectangle{
			Min: image.Point{X: bottomLine.rect.Min.X, Y: (bottomLine.rect.Max.Y + (bottomLine.rect.Max.Y - sls[i].rect.Max.Y))},
			Max: image.Point{X: bottomLine.rect.Max.X, Y: (bottomLine.rect.Max.Y + (bottomLine.rect.Max.Y - sls[i].rect.Min.Y))}}
		sls = append(sls, staffLine{r, notes[10+i-notesDown]})
	}

	// go 10 notes up
	notesUp := 9
	for i := 1; i < notesUp; i++ {
		r := image.Rectangle{
			Min: image.Point{X: sls[0].rect.Min.X, Y: (sls[0].rect.Min.Y - (sls[i].rect.Min.Y - sls[0].rect.Min.Y))},
			Max: image.Point{X: sls[0].rect.Max.X, Y: (sls[0].rect.Max.Y - (sls[i].rect.Min.Y - sls[0].rect.Min.Y))}}
		sls = append(sls, staffLine{r, notes[26+i]})
	}

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
	minPt := image.Point{0, line1.pt1.Y - 1}
	maxPt := image.Point{1000, line2.pt2.Y + 1}
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
