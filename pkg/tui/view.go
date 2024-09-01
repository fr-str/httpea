package tui

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/fr-str/httpea/internal/util"
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
	}

	// colour JSON
	if m.ReqView.header != nil && m.ReqView.header.Get("Content-Type") == "application/json" {
		m.ReqView.SetContent(hackedJsonColorizer(m.ReqView.Body, m.ReqView.Width-1))
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

var (
	resetColor = "\033[0m"
	keyBlue    = "\033[0;34m"
	strYellow  = "\033[38;2;175;215;95m"
	boolOrange = "\033[38;5;214m"
	regKey     = util.Must(regexp.Compile(`^(\s+)("[A-Za-z]+"):`))
	regValStr  = util.Must(regexp.Compile(`: (".*")`))
	regValBool = util.Must(regexp.Compile(`: (true|false)`))
	// includes float
	regValNum = util.Must(regexp.Compile(`: ([0-9]+(\.[0-9]+)?)`))
)

func hackedJsonColorizer(s string, limit int) string {
	ret := ""
	for _, l := range strings.Split(s, "\n") {
		hasComma := false
		if len(l) > 0 {
			hasComma = l[len(l)-1] == ','
		}
		lineLen := 0
		key := regKey.FindStringSubmatch(l)
		if len(key) == 0 {
			ret += l + "\n"
			continue
		}

		ret += fmt.Sprintf("%s%s%s%s: ", key[1], keyBlue, key[2], resetColor)
		lineLen += len(key[0])

		if valStr := regValStr.FindStringSubmatch(l); len(valStr) > 0 {
			ret += format(valStr, limit, lineLen, hasComma)
			continue
		}
		if valBool := regValBool.FindStringSubmatch(l); len(valBool) > 0 {
			ret += format(valBool, limit, lineLen, hasComma)
			continue
		}
		if valNum := regValNum.FindStringSubmatch(l); len(valNum) > 0 {
			ret += format(valNum, limit, lineLen, hasComma)
			continue
		}
		// if no match no color and remove key
		ret += l[len(key[0]):]
		if hasComma {
			ret += ","
		}
		ret += "\n"

	}
	return ret
}

func format(val []string, limit int, lineLen int, hasComma bool) string {
	ret := ""
	if len(val[1]) > limit {
		ret += writeWithLimit(val[1], strYellow, limit-lineLen)
		if hasComma {
			ret += ","
		}
		return ret + "\n"

	}
	ret += fmt.Sprintf("%s%s%s", strYellow, val[1], resetColor)
	if hasComma {
		ret += ","
	}
	return ret + "\n"
}

func writeWithLimit(s string, color string, limit int) string {
	if len(s) > limit {
		ret := fmt.Sprintf("%s%s%s\n", color, s[:limit], resetColor)
		ret += fmt.Sprintf("%s%s%s\n", color, s[limit:], resetColor)
		return ret
	}
	return fmt.Sprintf("%s%s%s\n", color, s, resetColor)
}

// lol works
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
