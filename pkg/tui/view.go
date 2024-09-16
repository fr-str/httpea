package tui

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var (
	lightYellow   = lipgloss.Color("229")
	durationStyle = lipgloss.NewStyle().Bold(true).Foreground(lightYellow)
	fileStyle     = lipgloss.NewStyle().Bold(true).Foreground(lightYellow)
	baseStyle     = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder())
)

// TODO: create a zone layout
func (m model) View() string {
	defer timeIt(time.Now(), "View")

	round := time.Microsecond
	if m.ReqView.reqDuration > time.Second {
		round = time.Millisecond
	}

	duration := m.ReqView.reqDuration.Round(round).String()
	lenDur := len(duration) - 2
	if m.Spinner.running {
		duration = m.Spinner.View()
		// ¯\_(ツ)_/¯
		lenDur = len(duration) - 8
	} else if m.ReqView.reqDuration < time.Millisecond && m.ReqView.reqDuration != 0 {
		// ¯\_(ツ)_/¯
		lenDur = len(duration) - 3
	}

	if m.ReqView.header != nil && m.ReqView.header.Get("Content-Type") == "application/json" {
		b, err := json.MarshalIndent(json.RawMessage(m.ReqView.Body), "", "  ")
		if err != nil {
			m.ReqView.SetContent(fmt.Sprintf("Error: %s", err))
		} else {
			m.ReqView.SetContent(string(b))
		}
	} else {
		m.ReqView.SetContent(m.ReqView.Body)
	}

	reqViewBorderFormat := "Output %s"
	reqView := nameBorder(m.ReqView.View(), fmt.Sprintf(reqViewBorderFormat,
		durationStyle.Render(duration)), len(reqViewBorderFormat)+lenDur)

	fileView := nameBorder(m.FileView.View(), "File Content")
	table := nameBorder(baseStyle.Render(m.FileTable.View()), "Files")
	views := []string{}
	views = append(views, lipgloss.JoinVertical(lipgloss.Top, table, fileView))
	views = append(views, reqView)

	help := m.help.View(m.keys)

	return lipgloss.JoinHorizontal(lipgloss.Top, views...) + "\n" + help
}

// lol works
// sometimes...
func nameBorder(s string, name string, forceLen ...int) string {
	text := []rune(name)
	b := []rune(s)
	st := 2
	end := len(text)
	if len(forceLen) > 0 {
		end = forceLen[0]
	}
	end += 2
	b = append(b[:st], append(text, b[end:]...)...)
	return string(b)
}
