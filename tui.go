package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	. "giv/printer"
	"image"
	"log"
	"os"
)

type keymap struct {
	zoom key.Binding
	quit key.Binding
}

type window struct {
	width  int
	height int
}

type model struct {
	path       string
	help       help.Model
	keymap     keymap
	window     window
	img        image.Image
	imgString  string
	renderData RenderData
	zoom       float32
	offset     image.Point
}

func wholeImageRender(window *window, viewport *ViewportSize) bool {
	return window.width >= viewport.Width && (window.height-2)*2 >= viewport.Height
}

func newModel(path string) model {
	imgString := ""
	img, err := ReadImageFile(path)

	if err != nil {
		imgString = err.Error()
	}

	return model{
		path: path,
		help: help.New(),
		keymap: keymap{
			zoom: key.NewBinding(key.WithKeys("-", "+"), key.WithHelp("-+", "zoom")),
			quit: key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
		},
		img:       img,
		imgString: imgString,
		zoom:      0,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	printImage := false

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "=":
			m.zoom += 0.25
			printImage = true
		case "+":
			m.zoom += 0.01
			printImage = true
		case "-":
			m.zoom = max(m.zoom-0.25, min(m.zoom, 0.25))
			printImage = true
		case "_":
			m.zoom = max(m.zoom-0.01, 0.01)
			printImage = true
		case "left":
			m.offset.X -= 1
			printImage = true
		case "right":
			m.offset.X += 1
			printImage = true
		case "up":
			m.offset.Y -= 1
			printImage = true
		case "down":
			m.offset.Y += 1
			printImage = true
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.window.width = msg.Width
		m.window.height = msg.Height

		printImage = true
	}

	if m.img != nil && printImage {
		windowWidth, windowHeight := m.window.width, m.window.height-2
		bounds := m.img.Bounds()
		imageHeight, imageWidth := bounds.Dy(), bounds.Dx()

		if m.zoom == 0 {
			_, scale := CalculateScale(imageHeight, imageWidth, windowHeight*2, windowWidth)
			if scale < 1 {
				scale = 1
			}
			scale = 1 / scale
			m.zoom = float32(scale)
		}

		viewportWidth := int(float32(imageWidth) * m.zoom)
		viewportHeight := int(float32(imageHeight) * m.zoom)

		viewport := ViewportSize{Width: viewportWidth, Height: viewportHeight}

		if wholeImageRender(&m.window, &viewport) {
			m.offset = image.Point{}
		}

		renderData := PrintImage(m.img, viewport, true, m.offset)
		m.imgString = renderData.ImageString
		m.renderData = renderData
	}

	return m, nil
}

func (m model) View() string {
	h := m.help.ShortHelpView([]key.Binding{
		m.keymap.zoom,
		m.keymap.quit,
	})

	topBarStyle := lipgloss.NewStyle().Background(lipgloss.Color("#fff")).Foreground(lipgloss.Color("#000"))
	zoomText := fmt.Sprintf(" Zoom: %d%%", int(m.zoom*100))

	imagePosition := lipgloss.Center
	if !wholeImageRender(&m.window, &m.renderData.Viewport) {
		imagePosition = lipgloss.Top
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			lipgloss.PlaceHorizontal(m.window.width-len(zoomText), lipgloss.Left, topBarStyle.Render(m.path), lipgloss.WithWhitespaceBackground(lipgloss.Color("#fff"))),
			topBarStyle.Render(zoomText),
		),
		lipgloss.Place(m.window.width, m.window.height-2, imagePosition, imagePosition, lipgloss.NewStyle().MaxWidth(m.window.width).MaxHeight(m.window.height-2).Render(m.imgString)),
		lipgloss.PlaceHorizontal(m.window.width, lipgloss.Center, lipgloss.NewStyle().Render(h)),
	)
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		log.Fatal("At least one argument is required")
	}

	m := newModel(args[0])

	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
