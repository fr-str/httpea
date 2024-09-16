package tui

import (
	"bytes"
	"fmt"
	"regexp"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/fr-str/httpea/internal/log"
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
	} else if m.ReqView.reqDuration < time.Millisecond && m.ReqView.reqDuration != 0 {
		// ¯\_(ツ)_/¯
		lenDur = len(duration) - 3
	}

	// colour JSON
	if m.ReqView.header != nil && m.ReqView.header.Get("Content-Type") == "application/json" {
		m.ReqView.SetContent(FixedJQ(m.ReqView.Body, m.ReqView.Width-1))
		// m.ReqView.SetContent(hackedJsonColorizer(m.ReqView.Body, m.ReqView.Width-1))
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
	resetColor         = "\033[0m"
	regANSIIColor      = util.Must(regexp.Compile(`\033\[[0-9];[0-9]+m`))
	regANSIIResetColor = util.Must(regexp.Compile(`\033\[0m`))
)

func FixedJQ(s string, limit int) string {
	b, _ := util.PrettyJSON([]byte(s))

	ret := bytes.NewBuffer(make([]byte, 0, len(b)))
	for _, l := range bytes.Split(b, []byte("\n")) {
		f := regANSIIColor.FindAllIndex(l, -1)
		if len(f) == 0 {
			continue
		}
		colors := regANSIIColor.FindAll(l, -1)
		color := []byte{}
		if len(colors) == 1 {
			color = colors[0]
		} else {
			color = colors[len(colors)-2]
		}

		lastIdx := f[len(f)-1][1]
		// ignore escape codes in line wreapping
		tmp := regANSIIColor.ReplaceAll(l, []byte{})
		tmp = regANSIIResetColor.ReplaceAll(tmp, []byte{})
		if len(tmp) > limit {
			writeWithLimit(ret, string(l[:lastIdx]), string(color), limit)
		} else {
			ret.WriteString(fmt.Sprintf("%s%s%s", color, l[:lastIdx], resetColor))
		}
		ret.Write(l[lastIdx:])
		ret.Write([]byte(resetColor))
		ret.WriteString("\n")
	}

	// return ""
	return ret.String()

}

func writeWithLimit(w *bytes.Buffer, s string, color string, limit int) {
	tmp := s
	space := spceLen(s)
	first := true
	for len(removeANSII(tmp)) > limit {
		ansiL := ansiiLen(tmp)
		log.Debug("ansiL", ansiL)
		line := tmp[:limit+ansiL]
		if !first {
			line = space + line
		}
		first = false
		w.WriteString(fmt.Sprintf("%s%s%s\n", color, line, resetColor))
		tmp = tmp[len(line):]
		log.Debug("[dupa] tmp: ", tmp)
	}
	w.WriteString(fmt.Sprintf("%s%s%s", color, space+tmp, resetColor))
}

func removeANSII(s string) string {
	s = string(regANSIIColor.ReplaceAll([]byte(s), []byte{}))
	return string(regANSIIResetColor.ReplaceAll([]byte(s), []byte{}))
}

func ansiiLen(s string) int {
	f1 := regANSIIColor.FindAllString(s, -1)
	f2 := regANSIIResetColor.FindAllString(s, -1)
	count := 0
	for _, v := range f1 {
		count += len(v)
	}
	for _, v := range f2 {
		count += len(v)
	}
	return count
}

func spceLen(s string) string {
	count := 0
	for i := 0; s[i] == ' '; i++ {
		count++
	}
	return s[:count]
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
