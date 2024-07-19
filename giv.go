package main

import (
	"fmt"
	. "giv/printer"
	"log"
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

	termSize, err := getTerminalSize()
	if err != nil {
		log.Fatal(err)
	}

	// Take terminal's input line into account, otherwise 2 topmost pixel rows will not be visible without scrolling up.
	termSize.Height -= 1

	for _, path := range args {
		imageStr, err := PrintImageFile(path, termSize)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(imageStr)
		}
	}
}

func getTerminalSize() (ViewportSize, error) {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	sizeOutput, err := cmd.Output()
	if err != nil {
		return ViewportSize{}, fmt.Errorf("error calling \"stty size\": %w", err)
	}

	sizes := strings.Split(strings.TrimSpace(string(sizeOutput)), " ")

	height, err := strconv.Atoi(sizes[0])
	if err != nil {
		return ViewportSize{}, err
	}

	width, err := strconv.Atoi(sizes[1])
	if err != nil {
		return ViewportSize{}, err
	}

	return ViewportSize{Width: width, Height: height}, nil
}
