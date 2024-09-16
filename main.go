package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fr-str/httpea/internal/config"
	"github.com/fr-str/httpea/pkg/curl"
	"github.com/fr-str/httpea/pkg/pea"
	"github.com/fr-str/httpea/pkg/tui"
)

func main() {
	curlIn := flag.String("c", "", "transform curl command to pea format")
	curlOut := flag.String("e", "", "transform pea file to curl command")
	flag.Parse()

	if config.Debug {
		f, err := tea.LogToFile("debug.log", "debug")
		if err != nil {
			fmt.Println("fatal:", err)
			os.Exit(1)
		}
		defer f.Close()
		config.DebugLogFile = f
		// config.DebugLogFile = os.Stdout
	}

	if *curlIn != "" {
		curl.ParseCurl(*curlIn)
		return
	}
	if *curlOut != "" {
		p, err := pea.GetRequestDataFromFile(*curlOut, nil)
		if err != nil {
			panic(err)
		}
		out, err := curl.PeaToCurl(p)
		if err != nil {
			panic(err)
		}
		fmt.Println(out)
		return
	}

	startTUI()
}

func startTUI() {
	// b, _ := os.ReadFile("example.json")
	// fmt.Println(tui.FixedJQ(string(b), 20))
	// return
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
