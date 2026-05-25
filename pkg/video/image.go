package video

import (
	"image"
	"reflect"
)

// ValidImage reports whether img holds a non-nil concrete value.
// Go interfaces can be non-nil while holding a typed nil pointer.
func ValidImage(img image.Image) bool {
	if img == nil {
		return false
	}
	v := reflect.ValueOf(img)
	return v.Kind() != reflect.Ptr || !v.IsNil()
}
