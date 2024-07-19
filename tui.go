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
}

func newModel(path string) model {
	img, _ := ReadImageFile(path)

	return model{
		path: path,
		help: help.New(),
		keymap: keymap{
			quit: key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
		},
		img: img,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		msgString := msg.String()
		if msgString == "ctrl+c" || msgString == "q" {
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.window.width = msg.Width
		m.window.height = msg.Height

		img, err := PrintImage(m.img, ViewportSize{Width: m.window.width, Height: m.window.height - 2})

		if err != nil {
			log.Fatal(err)
		}

		m.imgString = img
	}

	return m, nil
}

func (m model) View() string {
	h := m.help.ShortHelpView([]key.Binding{
		m.keymap.quit,
	})

	return lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.NewStyle().Width(m.window.width).Background(lipgloss.Color("#fff")).Foreground(lipgloss.Color("#000")).Render(m.path),
		lipgloss.Place(m.window.width, m.window.height-2, lipgloss.Center, lipgloss.Center, m.imgString),
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
