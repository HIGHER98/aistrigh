package amharc

import (
	"image"

	"gocv.io/x/gocv"
)

const YPadding = 30

// Extract all the bars into different regions represented by an image.Rectangle for each bar
func ExtractBars(img gocv.Mat) ([]image.Rectangle, error) {
	clefs, err := findClefs(img)
	if err != nil {
		return nil, err
	}

	// With these treble clefs, we can use the information to extract a bar by adding Y padding and drawing a rect from x = 0 to x = img.Cols()

	var rects []image.Rectangle
	for _, clef := range clefs {
		rects = append(rects, image.Rectangle{
			Min: image.Point{0, clef.Min.Y - YPadding},
			Max: image.Point{img.Cols(), clef.Max.Y + YPadding},
		})

	}

	return rects, nil
}
