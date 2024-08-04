package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fr-str/httpea/internal/config"
	"github.com/fr-str/httpea/pkg/tui"
	"github.com/google/shlex"
)

func main() {
	curl := flag.String("curl", "", "transform curl command to pea format")
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

	if *curl != "" {
		parseCurl(*curl)
		return
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

func parseCurl(curl string) {
	pc, err := Parse(curl)
	if err != nil {
		panic(err)
	}
	transformToPea(pc)
}

func transformToPea(pc *parsedCurl) {
	buf := strings.Builder{}
	buf.WriteString("[Host]\n")
	buf.WriteString(pc.Url)
	buf.WriteString("\n")
	buf.WriteString("\n")
	buf.WriteString("[Method]\n")
	buf.WriteString(pc.Method)
	buf.WriteString("\n")
	buf.WriteString("\n")
	buf.WriteString("[Headers]\n")
	for k, v := range pc.Headers {
		buf.WriteString(k)
		buf.WriteString(": ")
		buf.WriteString(v)
		buf.WriteString("\n")
	}
	buf.WriteString("\n")
	buf.WriteString("[Body]\n")
	buf.WriteString(pc.Body)
	fmt.Println(buf.String())
}

type parsedCurl struct {
	Method  string
	Url     string
	Body    string
	Headers map[string]string
}

func Parse(curl string) (*parsedCurl, error) {
	pc := &parsedCurl{Headers: map[string]string{}}

	parts, err := shlex.Split(curl)
	if err != nil {
		return nil, err
	}

	var trimmedParts []string
	for _, p := range parts {
		part := strings.TrimSpace(p)
		if part != "" {
			trimmedParts = append(trimmedParts, part)
		}
	}

	var currentPart, nextPart string
	for i := 1; i < len(trimmedParts); {
		currentPart = trimmedParts[i-1]
		nextPart = trimmedParts[i]
		if currentPart != "" {
			switch currentPart {
			case "-X":
				pc.Method = strings.ToUpper(nextPart)
				i++
			case "-H":
				k, v, err := parseHeader(nextPart)
				if err != nil {
					return nil, err
				}
				pc.Headers[strings.ToLower(k)] = v
				i++
			case "-d":
				pc.Body = nextPart
				i++
			case "--data-raw":
				pc.Body = nextPart
				i++
			case "--abstract-unix-socket":
				i++
			case "--alt-svc":
				i++
			case "--aws-sigv4":
				i++
			case "-a":
			case "--append":
			case "--anyauth":
			case "--basic":
			case "curl":
			case "-k":
			case "-v":
			case "-V":
			default:
				if !strings.HasPrefix(currentPart, "-") {
					pc.Url = currentPart
				}
			}
		}
		i++
	}

	if pc.Body != "" {
		if pc.Method == "" {
			pc.Method = "POST"
		}
	} else {
		if pc.Method == "" {
			pc.Method = "GET"
		}
	}

	return pc, nil
}

func parseHeader(h string) (string, string, error) {
	parts := strings.SplitN(h, ":", 2)
	if len(parts) == 2 {
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), nil
	}
	return "", "", errors.New(fmt.Sprintf(`wrong header format: %v`, h))
}
