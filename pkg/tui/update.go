package tui

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fr-str/httpea/internal/config"
	"github.com/fr-str/httpea/internal/log"
	"github.com/fr-str/httpea/internal/util"
	"github.com/fr-str/httpea/internal/util/env"
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
	log.Debug("----------------------new update----------------------")
	defer timeIt(time.Now(), "Update")

	var cmd tea.Cmd
	switch msg := msg.(type) {
	case editDone:
		if msg.err != nil {
			m.ReqView.SetContent(msg.err.Error())
		}

	case tea.WindowSizeMsg:
		m.help.Width = msg.Width
		m.handleSizes(msg)

	case spinner.TickMsg:
		if m.Spinner.running {
			m.Spinner.Model, cmd = m.Spinner.Update(msg)
			return m, cmd
		}

	case reqError:
		m.Spinner.running = false
		m.ReqView.reqDuration = 0
		m.ReqView.Body = msg.Error()
		m.ReqView.SetContent(m.ReqView.Body)

	case *pea.Response:
		m.Spinner.running = false

		m.ReqView.reqDuration = msg.Duration
		m.ReqView.Body = getBody(msg.Response)
		m.ReqView.header = msg.Header
		return m, cmd

	case tea.KeyMsg:
		log.Debug(msg.String())
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.keys.Tab):
			m.focus++
			return m, cmd

		case key.Matches(msg, m.keys.Up):
		case key.Matches(msg, m.keys.Down):
		case key.Matches(msg, m.keys.C):
			return m, openEditor(filepath.Join(config.FileFolder, "config.pea"))

		case key.Matches(msg, m.keys.E):
			f := m.FileTable.SelectedRow()[2]
			return m, openEditor(util.GetPeaFilePath(f))

		case key.Matches(msg, m.keys.Space):
			m.handleSelecting()
			return m, cmd

		case key.Matches(msg, m.keys.Enter):
			file := m.FileTable.SelectedRow()[2]
			m.Spinner.running = true
			return m, tea.Batch(handleRequest(m.client, file), m.Spinner.Tick)

		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
			return m, cmd
		}
	}

	m, cmd = m.handleFocus(msg)
	if len(m.FileTable.SelectedRow()) > 0 {
		fileContent := util.ReadPeaFile(m.FileTable.SelectedRow()[2])
		fileContent = pea.ResolveEnvVars(fileContent, m.client.Env)
		m.FileView.SetContent(pea.RegCategory.ReplaceAllStringFunc(fileContent, func(b string) string {
			ss := fileStyle.Render(string(b))
			return ss
		}))
	}
	return m, cmd
}

func (m *model) handleSelecting() {
	// table has 3 columns,
	// number | method | filename
	number := m.FileTable.SelectedRow()[0]
	if number == "" {
		//select file
		m.selected = append(m.selected, m.FileTable.SelectedRow()[2])
	} else {
		// deselect a file
		idx := slices.Index(m.selected, m.FileTable.SelectedRow()[2])
		log.Debug("m.selected: ", m.selected, len(m.selected))
		log.Debug("idx: ", idx)
		if idx == -1 {
			return
		}
		m.selected = slices.Delete(m.selected, idx, idx+1)
		log.Debug("m.selected: ", m.selected, len(m.selected))
	}

	// number selected files in view
	for _, r := range m.FileTable.Rows() {
		idx := slices.Index(m.selected, r[2])
		if idx != -1 {
			r[0] = fmt.Sprint(idx + 1)
			continue
		}
		r[0] = ""
	}

	m.FileTable.UpdateViewport()
}

type response *pea.Response
type reqError error

func handleRequest(c *pea.Client, file string) tea.Cmd {
	return func() tea.Msg {
		resp, err := c.Request(file)
		if err != nil {
			return reqError(err)
		}
		return resp
	}
}

func getBody(resp *http.Response) string {
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	ct := strings.Split(resp.Header.Get("Content-Type"), ";")[0]
	if ct == "application/json" {
		var err error
		out, err := json.MarshalIndent(json.RawMessage(b), "", "  ")
		if err != nil {
			return err.Error()
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
			d, _ := pea.GetRequestDataFromFile(f, m.client.Env)
			rows = append(rows, components.Row{m.reSelect(f), d.Method, f})
		}
		m.FileTable.SetRows(rows)
		m.FileTable, cmd = m.FileTable.Update(msg)

	case 1:
		m.ReqView.Model, cmd = m.ReqView.Update(msg)

	}

	return m, cmd
}

func (m *model) reSelect(f string) string {
	idx := slices.Index(m.selected, f)
	if idx == -1 {
		return ""
	}
	return fmt.Sprint(idx + 1)
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
