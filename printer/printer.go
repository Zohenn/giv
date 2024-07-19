package printer

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"os"
	"strings"
)

type ViewportSize struct {
	Width  int
	Height int
}

func PrintImageFile(path string, viewportSize ViewportSize) (string, error) {
	img, err := ReadImageFile(path)
	if err != nil {
		return "", fmt.Errorf("error opening image file: %w", err)
	}

	return PrintImage(img, viewportSize)
}

func PrintImage(img image.Image, viewportSize ViewportSize) (string, error) {
	if viewportSize.Height == 0 || viewportSize.Width == 0 {
		return "", nil
	}

	bounds := img.Bounds()
	windowSize := calculateScale(bounds.Max.Y-bounds.Min.Y, bounds.Max.X-bounds.Min.X, viewportSize.Height*2, viewportSize.Width)
	outputString := strings.Builder{}

	for y := bounds.Min.Y; y < bounds.Max.Y; y += windowSize * 2 {
		for x := bounds.Min.X; x < bounds.Max.X; x += windowSize {
			fr, fg, fb := interpolatePixelValue(&img, x, y, windowSize)
			outputString.WriteString(fmt.Sprintf("\x1b[38;2;%d;%d;%dm", fr, fg, fb))

			if y+windowSize < bounds.Max.Y {
				br, bg, bb := interpolatePixelValue(&img, x, y+windowSize, windowSize)
				outputString.WriteString(fmt.Sprintf("\x1b[48;2;%d;%d;%dm", br, bg, bb))
			}

			outputString.WriteString(fmt.Sprintf("\u2580"))
		}

		if y+windowSize*2 < bounds.Max.Y {
			outputString.WriteString(fmt.Sprintln("\x1b[0m"))
		} else {
			outputString.WriteString(fmt.Sprint("\x1b[0m"))
		}
	}

	return outputString.String(), nil
}

func calculateScale(imageHeight int, imageWidth int, viewportHeight int, viewportWidth int) int {
	iHeight, iWidth, vHeight, vWidth := float64(imageHeight), float64(imageWidth), float64(viewportHeight), float64(viewportWidth)

	widthRatio := iWidth / vWidth
	heightRatio := iHeight / vHeight

	// This already handles the case when image is smaller than viewport.
	return int(math.Ceil(max(widthRatio, heightRatio)))
}

func interpolatePixelValue(img *image.Image, startX int, startY int, windowSize int) (uint32, uint32, uint32) {
	// Prevent single iteration loops.
	if windowSize == 1 {
		r, g, b, _ := (*img).At(startX, startY).RGBA()
		return r >> 8, g >> 8, b >> 8
	}

	bounds := (*img).Bounds()

	var rSum, gSum, bSum uint32 = 0, 0, 0

	yUpperBound := min(bounds.Max.Y, startY+windowSize)
	xUpperBound := min(bounds.Max.X, startX+windowSize)

	for y := startY; y < yUpperBound; y++ {
		for x := startX; x < xUpperBound; x++ {
			r, g, b, _ := (*img).At(x, y).RGBA()

			rSum += r
			gSum += g
			bSum += b
		}
	}

	count := uint32((yUpperBound - startY) * (xUpperBound - startX))

	return (rSum / count) >> 8, (gSum / count) >> 8, (bSum / count) >> 8
}

func ReadImageFile(path string) (image.Image, error) {
	reader, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer reader.Close()

	img, _, err := image.Decode(reader)
	if err != nil {
		return nil, err
	}

	return img, nil
}
