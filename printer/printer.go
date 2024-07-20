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

type RenderData struct {
	ImageString string
	Scale       int
	ActualScale float64
}

func PrintImageFile(path string, viewportSize ViewportSize) (RenderData, error) {
	img, err := ReadImageFile(path)
	if err != nil {
		return RenderData{}, fmt.Errorf("error opening image file: %w", err)
	}

	return PrintImage(img, viewportSize, false), nil
}

func PrintImage(img image.Image, viewportSize ViewportSize, useActualScale bool) RenderData {
	if viewportSize.Height == 0 || viewportSize.Width == 0 {
		return RenderData{}
	}

	bounds := img.Bounds()
	windowSize, actualScale := calculateScale(bounds.Max.Y-bounds.Min.Y, bounds.Max.X-bounds.Min.X, viewportSize.Height*2, viewportSize.Width)
	outputString := strings.Builder{}

	for vy := 0; vy < viewportSize.Height; vy++ {
		var y int
		if useActualScale {
			y = int(math.Round(float64(vy)*actualScale*2)) + bounds.Min.Y
		} else {
			y = vy*windowSize*2 + bounds.Min.Y
		}

		if y >= bounds.Max.Y {
			continue
		}

		for vx := 0; vx < viewportSize.Width; vx++ {
			var x int
			if useActualScale {
				x = int(math.Round(float64(vx)*actualScale)) + bounds.Min.X
			} else {
				x = vx*windowSize + bounds.Min.X
			}

			if x >= bounds.Max.X {
				continue
			}

			var fr, fg, fb uint32
			if useActualScale {
				fr, fg, fb = interpolatePixelValueFloat(&img, x, y, actualScale)
			} else {
				fr, fg, fb = interpolatePixelValue(&img, x, y, windowSize)
			}
			outputString.WriteString(fmt.Sprintf("\x1b[38;2;%d;%d;%dm", fr, fg, fb))

			var bottomRowY int
			if useActualScale {
				bottomRowY = int(math.Round(float64(y) + actualScale))
			} else {
				bottomRowY = y + windowSize
			}

			if bottomRowY < bounds.Max.Y {
				var br, bg, bb uint32
				if useActualScale {
					br, bg, bb = interpolatePixelValueFloat(&img, x, bottomRowY, actualScale)
				} else {
					br, bg, bb = interpolatePixelValue(&img, x, bottomRowY, windowSize)
				}
				outputString.WriteString(fmt.Sprintf("\x1b[48;2;%d;%d;%dm", br, bg, bb))
			}

			outputString.WriteString(fmt.Sprintf("\u2580"))
		}

		if vy < viewportSize.Height-1 {
			outputString.WriteString(fmt.Sprintln("\x1b[0m"))
		} else {
			outputString.WriteString(fmt.Sprint("\x1b[0m"))
		}
	}

	return RenderData{
		ImageString: outputString.String(),
		Scale:       windowSize,
		ActualScale: actualScale,
	}
}

func calculateScale(imageHeight int, imageWidth int, viewportHeight int, viewportWidth int) (int, float64) {
	iHeight, iWidth, vHeight, vWidth := float64(imageHeight), float64(imageWidth), float64(viewportHeight), float64(viewportWidth)

	widthRatio := iWidth / vWidth
	heightRatio := iHeight / vHeight

	ratio := max(widthRatio, heightRatio)

	// Handle the case when image is smaller than viewport.
	if ratio < 1 {
		return int(math.Ceil(ratio)), ratio
	}

	return int(math.Round(ratio)), ratio
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

func interpolatePixelValueFloat(img *image.Image, startX int, startY int, scale float64) (uint32, uint32, uint32) {
	// Prevent single iteration loops.
	if scale <= 1 {
		r, g, b, _ := (*img).At(startX, startY).RGBA()
		return r >> 8, g >> 8, b >> 8
	}

	bounds := (*img).Bounds()

	var rSum, gSum, bSum uint32 = 0, 0, 0

	yUpperBound := min(bounds.Max.Y, int(math.Round(float64(startY)+scale)))
	xUpperBound := min(bounds.Max.X, int(math.Round(float64(startX)+scale)))

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

	defer func(reader *os.File) {
		_ = reader.Close()
	}(reader)

	img, _, err := image.Decode(reader)
	if err != nil {
		return nil, err
	}

	return img, nil
}
