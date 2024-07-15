package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		log.Fatal("At least one argument is required")
	}

	path := args[0]

	img, err := readImageFile(path)
	if err != nil {
		log.Fatal(err)
	}

	termSize, err := getTerminalSize()
	if err != nil {
		log.Fatal(err)
	}

	// Take terminal's input line into account, otherwise 2 topmost pixel rows will not be visible without scrolling up.
	termSize.height -= 1

	bounds := img.Bounds()
	windowSize := calculateScale(bounds.Max.Y-bounds.Min.Y, bounds.Max.X-bounds.Min.X, termSize.height*2, termSize.width)

	for y := bounds.Min.Y; y < bounds.Max.Y; y += windowSize * 2 {
		for x := bounds.Min.X; x < bounds.Max.X; x += windowSize {
			fr, fg, fb := interpolatePixelValue(&img, x, y, windowSize)
			fmt.Printf("\x1b[38;2;%d;%d;%dm", fr, fg, fb)

			if y+windowSize < bounds.Max.Y {
				br, bg, bb := interpolatePixelValue(&img, x, y+windowSize, windowSize)
				fmt.Printf("\x1b[48;2;%d;%d;%dm", br, bg, bb)
			}

			fmt.Printf("\u2580")
		}
		fmt.Println("\x1b[0m")
	}
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

func readImageFile(path string) (image.Image, error) {
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

type terminalSize struct {
	width  int
	height int
}

func getTerminalSize() (terminalSize, error) {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	sizeOutput, err := cmd.Output()
	if err != nil {
		return terminalSize{}, fmt.Errorf("error calling \"stty size\": %w", err)
	}

	sizes := strings.Split(strings.TrimSpace(string(sizeOutput)), " ")

	height, err := strconv.Atoi(sizes[0])
	if err != nil {
		return terminalSize{}, err
	}

	width, err := strconv.Atoi(sizes[1])
	if err != nil {
		return terminalSize{}, err
	}

	return terminalSize{width, height}, nil
}
