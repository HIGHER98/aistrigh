package amharc

import (
	"image"
	"testing"
)

func TestStripOutliar(t *testing.T) {
	rects := []image.Rectangle{
		image.Rect(0, 1, 0, 2),
		image.Rect(0, 2, 0, 3),
		image.Rect(0, 3, 0, 4),
		image.Rect(0, 25, 0, 26),
		image.Rect(0, 46, 0, 47), // rects[4]
		image.Rect(0, 47, 0, 48),
		image.Rect(0, 48, 0, 49),
		image.Rect(0, 49, 0, 50),
		image.Rect(0, 50, 0, 51), // rects[8]
		image.Rect(0, 75, 0, 75),
		image.Rect(0, 76, 0, 77),
		image.Rect(0, 77, 0, 78),
	}
	strippedRects, err := stripOutliar(rects, 5)
	if err != nil {
		t.Error(err)
	}
	if len(strippedRects) != 5 {
		t.Errorf("Expected 5 elements, %d found", len(strippedRects))
	}

	_, err = stripOutliar(rects, 20)
	if err == nil {
		t.Error("Expected an error for providing a number greater than the number of rects")
	}

}
