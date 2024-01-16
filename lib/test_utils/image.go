package test_utils

import (
	"image"
	"image/color"
)

// CreateTestImage creates a test image of a smooth gradient color for the
// given width and height.
func CreateTestImage(width, height int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Adjust the color based on your desired gradient or pattern
			r := uint8((float64(x) / float64(width)) * 255)
			g := uint8((float64(y) / float64(height)) * 255)
			b := uint8((float64(x+y) / float64(width+height)) * 255)

			img.SetRGBA(x, y, color.RGBA{r, g, b, 255})
		}
	}
	return img
}
