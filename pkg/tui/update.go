package tui

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fr-str/httpea/internal/config"
	"github.com/fr-str/httpea/internal/log"
	"github.com/fr-str/httpea/internal/util"
	"github.com/fr-str/httpea/internal/util/env"
	"github.com/fr-str/httpea/pkg/client"
	"github.com/fr-str/httpea/pkg/components"
	"github.com/fr-str/httpea/pkg/pea"
)

type editDone struct{ err error }

func openEditor(file string) tea.Cmd {
	cmd := exec.Command(env.Get("EDITOR", "vim"), file)
	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		return editDone{err}
	})
}

// timeIt function
func timeIt(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Debug(fmt.Sprintf("%s took %s", name, elapsed))
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	log.Debug("----------------------new loop----------------------")
	defer timeIt(time.Now(), "Update")

	var cmd tea.Cmd
	switch msg := msg.(type) {
	case editDone:
		if msg.err != nil {
			m.ReqView.SetContent(msg.err.Error())
		}
		m.client.LoadAuto(m.Env)

	case tea.WindowSizeMsg:
		m.help.Width = msg.Width
		m.handleSizes(msg)

	case tea.KeyMsg:
		log.Debug(msg.String())
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Tab):
			m.focus++

		case key.Matches(msg, m.keys.Up):
		case key.Matches(msg, m.keys.Down):
		case key.Matches(msg, m.keys.C):
			return m, openEditor(filepath.Join(config.FileFolder, "config.pea"))

		case key.Matches(msg, m.keys.E):
			f := m.FileTable.SelectedRow()[2]
			return m, openEditor(util.GetPeaFilePath(f))

		case key.Matches(msg, m.keys.Space):
			s := m.FileTable.SelectedRow()[0]
			if s == "" {
				m.selectFile(m.FileTable.SelectedRow()[2])
			} else {
				m.deSelectFile(m.FileTable.SelectedRow()[2])
			}
			for _, r := range m.FileTable.Rows() {
				idx := slices.Index(m.selected, r[2])
				if idx != -1 {
					r[0] = fmt.Sprint(idx + 1)
					continue
				}
				r[0] = ""
			}

			m.FileTable.UpdateViewport()

		case key.Matches(msg, m.keys.Enter):
			m.handleEnter()

		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
		}
	}
	m, cmd = m.handleFocus(msg)
	if len(m.FileTable.SelectedRow()) > 0 {
		fileContent := util.ReadPeaFile(m.FileTable.SelectedRow()[2])
		fileContent = pea.ResolveEnvVars(fileContent, m.Env)
		m.FileView.SetContent(colorFileContent(fileContent))
	}
	return m, cmd
}

func colorFileContent(s string) string {
	b := pea.RegCategory.ReplaceAllStringFunc(s, func(b string) string {
		ss := fileStyle.Render(string(b))
		return ss
	})
	return b
}

func (m *model) selectFile(f string) {
	m.selected = append(m.selected, f)
}

func (m *model) deSelectFile(f string) {
	idx := slices.Index(m.selected, f)
	log.Debug("m.selected: ", m.selected, len(m.selected))
	log.Debug("idx: ", idx)
	if idx == -1 {
		return
	}
	m.selected = slices.Delete(m.selected, idx, idx+1)
	log.Debug("m.selected: ", m.selected, len(m.selected))
}

func (m *model) reSelect(f string) string {
	idx := slices.Index(m.selected, f)
	if idx == -1 {
		return ""
	}
	return fmt.Sprint(idx + 1)
}

func (m *model) handleEnter() {
	// if len(m.selected) == 0 {
	// }
	file := m.FileTable.SelectedRow()[2]
	d, err := pea.GetRequestDataFromFile(file, m.Env)
	if err != nil {
		m.ReqView.SetContent(err.Error())
		return
	}
	resp, err := m.client.Request(d)
	if err != nil {
		m.ReqView.SetContent(err.Error())
		return
	}

	if resp != nil {
		log.Debug("[dupa] m.client.AutoCode: ", m.client.AutoCode)
		if a, ok := m.client.AutoCode[strconv.Itoa(resp.StatusCode)]; ok {
			resp, err = a()
			if err != nil {
				m.ReqView.SetContent(err.Error())
				return
			}

			m.doExports(resp)
			d, err := pea.GetRequestDataFromFile(file, m.Env)
			if err != nil {
				m.ReqView.SetContent(err.Error())
				return
			}
			resp, err = m.client.Request(d)
			if err != nil {
				m.ReqView.SetContent(err.Error())
				return
			}
		}
	}

	m.ReqView.reqDuration = resp.Duration
	m.ReqView.body = getBody(resp.Response)
	m.ReqView.header = resp.Header
	m.ReqView.SetContent(m.ReqView.body)
	m.doExports(resp)
}

// func (m *model) doAuto() {
// 	for _, a := range m.client.AutoCode {
// 		resp, err := a()
// 		if err != nil {
// 			m.ReqView.SetContent(err.Error())
// 			continue
// 		}
// 		m.doExports(resp)
// 	}
// }

func (m *model) doExports(resp *client.Response) {
	body := m.ReqView.body
	errs := ""
	log.Debug("resp.BodyExports: ", len(resp.BodyExports))
	for _, expr := range resp.BodyExports {
		v, err := expr.Expr(body)
		if err != nil {
			errs += err.Error()
		}

		m.Env[expr.Name] = v
		log.Debug("expr.EnvName[]: ", expr.Name, v)
	}

	log.Debug("[dupa] resp.HeaderExports: ", len(resp.HeaderExports))
	for _, expr := range resp.HeaderExports {
		headerName, _ := expr.Expr(body)
		m.Env[expr.Name] = resp.Header.Get(headerName)
	}

	if len(errs) != 0 {
		m.ReqView.SetContent(errs + "\n" + body)
	}
}

func getBody(resp *http.Response) string {
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	ct := strings.Split(resp.Header.Get("Content-Type"), ";")[0]
	if ct == "application/json" {
		var err error
		out, err := util.PrettyJSON(b)
		log.Debug("err: ", err)
		if err != nil {
			out, err = json.MarshalIndent(json.RawMessage(b), "", " ")
			if err != nil {
				os.WriteFile(fmt.Sprintf("pea-%s.json", time.Now()), b, 0744)
				return err.Error()
			}
		}
		return string(out)
	}
	return string(b)
}

func (m model) handleFocus(msg tea.Msg) (model, tea.Cmd) {
	var cmd tea.Cmd
	if m.focus > 1 {
		m.focus = 0
	}

	m.FileTable.Blur()
	switch m.focus {
	case 0:
		m.FileTable.Focus()
		rows := []components.Row{}
		for _, f := range listFiles() {
			d, _ := pea.GetRequestDataFromFile(f, m.Env)
			rows = append(rows, components.Row{m.reSelect(f), d.Method, f})
		}
		m.FileTable.SetRows(rows)
		m.FileTable, cmd = m.FileTable.Update(msg)

	case 1:
		m.ReqView.Model, cmd = m.ReqView.Update(msg)

	}

	return m, cmd
}

// TODO: create a zone layout
func (m *model) handleSizes(msg tea.WindowSizeMsg) {
	magic := 2
	borderW := lipgloss.Width(baseStyle.Render(""))
	log.Debug("magicW: ", borderW)
	borderH := lipgloss.Height(baseStyle.Render(""))
	log.Debug("magicH: ", borderH)

	peaTableVi := m.FileTable.View()
	peaTableW := lipgloss.Width(peaTableVi) + borderW
	peaTableH := lipgloss.Height(peaTableVi) + borderH

	helpH := lipgloss.Height(m.help.View(m.keys))
	log.Debug("helpH: ", helpH)

	m.ReqView.Width = msg.Width - peaTableW - borderW
	log.Debug("peaTableW: ", peaTableW)
	log.Debug("msg.Width: ", msg.Width)
	log.Debug("msg.Height: ", msg.Height)
	log.Debug("m.ReqView.Width: ", m.ReqView.Width)
	m.ReqView.Height = msg.Height - helpH - borderH

	m.FileView.Width = msg.Width - lipgloss.Width(m.ReqView.View()) - magic
	m.FileView.Height = msg.Height - peaTableH - helpH - magic
}
