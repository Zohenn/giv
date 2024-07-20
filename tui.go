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
	path string
	help help.Model
	keymap
	window    window
	img       image.Image
	imgString string
	zoom      int
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
		zoom:      100,
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
			m.zoom += 25
			printImage = true
		case "-":
			m.zoom = max(m.zoom-25, 25)
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
		viewportWidth, viewportHeight := m.window.width, m.window.height-2

		viewportWidth = int(float32(viewportWidth) * float32(m.zoom) / 100)
		viewportHeight = int(float32(viewportHeight) * float32(m.zoom) / 100)

		renderData := PrintImage(m.img, ViewportSize{Width: viewportWidth, Height: viewportHeight}, true)
		m.imgString = renderData.ImageString

		f, err := tea.LogToFile("/tmp/giv.log", "debug")

		log.Printf("%dx%d %d %f\n", viewportWidth, viewportHeight, renderData.Scale, renderData.ActualScale)

		if err != nil {
			m.imgString = err.Error()
		}

		f.Close()
	}

	return m, nil
}

func (m model) View() string {
	h := m.help.ShortHelpView([]key.Binding{
		m.keymap.zoom,
		m.keymap.quit,
	})

	topBarStyle := lipgloss.NewStyle().Background(lipgloss.Color("#fff")).Foreground(lipgloss.Color("#000"))
	zoomText := fmt.Sprintf(" Zoom: %d%%", m.zoom)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			lipgloss.PlaceHorizontal(m.window.width-len(zoomText), lipgloss.Left, topBarStyle.Render(m.path), lipgloss.WithWhitespaceBackground(lipgloss.Color("#fff"))),
			topBarStyle.Render(zoomText),
		),
		lipgloss.Place(m.window.width, m.window.height-2, lipgloss.Center, lipgloss.Center, lipgloss.NewStyle().MaxWidth(m.window.width).MaxHeight(m.window.height-2).Render(m.imgString)),
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
