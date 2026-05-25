package video

import (
	"image"
	"testing"
)

func TestValidImage(t *testing.T) {
	if ValidImage(nil) {
		t.Fatal("nil should be invalid")
	}
	var y *image.YCbCr
	if ValidImage(y) {
		t.Fatal("typed nil *YCbCr should be invalid")
	}
	img := image.NewYCbCr(image.Rect(0, 0, 2, 2), image.YCbCrSubsampleRatio420)
	if !ValidImage(img) {
		t.Fatal("real YCbCr should be valid")
	}
}
