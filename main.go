package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fr-str/httpea/internal/config"
	"github.com/fr-str/httpea/pkg/tui"
)

func main() {
	if config.Debug {
		f, err := tea.LogToFile("debug.log", "debug")
		if err != nil {
			fmt.Println("fatal:", err)
			os.Exit(1)
		}
		defer f.Close()
		config.DebugLogFile = f
	}

	startTUI()
}

func startTUI() {
	// m := tui.InitialModel()
	// mo, _ := m.Update(tea.WindowSizeMsg{
	// 	Width:  120,
	// 	Height: 50,
	// })
	// fmt.Println(mo.View())
	// return
	p := tea.NewProgram(tui.InitialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("there's been an error: %v", err)
		os.Exit(1)
	}
}
