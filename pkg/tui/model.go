package tui

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/fr-str/httpea/internal/config"
	"github.com/fr-str/httpea/internal/log"
	"github.com/fr-str/httpea/pkg/components"
	"github.com/fr-str/httpea/pkg/pea"
)

type model struct {
	focus  uint
	client *pea.Client
	// TODO: not implemented
	selected  []string
	keys      keyMap
	help      help.Model
	FileTable components.Table
	FileView  viewport.Model
	ReqView   reqView
	Spinner   spinnerDupa
	NumBuf    string
}

type reqView struct {
	viewport.Model
	reqDuration   time.Duration
	totalDuration time.Duration
	Body          string
	header        http.Header
	ShowHeaders   bool
}

type spinnerDupa struct {
	spinner.Model
	running bool
}

func InitialModel() model {
	m := model{
		keys:    keys,
		help:    help.New(),
		ReqView: reqView{Model: viewport.New(10, 20)},
		client:  pea.NewClient(),
		Spinner: spinnerDupa{
			spinner.New(spinner.WithSpinner(spinner.Points)),
			false,
		},
	}

	m.FileTable = initTable()
	m.FileView = viewport.New(m.FileTable.Width(), m.FileTable.Height())
	m.FileView.Style = baseStyle
	m.ReqView.Style = baseStyle
	return m
}

func (m model) Init() tea.Cmd {
	return tea.SetWindowTitle("httpea")
}

func initTable() components.Table {
	rows := []components.Row{}
	for _, f := range listFiles() {
		d, _ := pea.GetRequestDataFromFile(f, map[string]string{})
		rows = append(rows, components.Row{"", d.Method, f})
	}
	col := []components.Column{
		{Width: 1}, {Title: "", Width: 7}, {Title: "Files", Width: 18},
	}

	t := components.NewTable(
		components.WithColumns(col),
		components.WithRows(rows),
		components.WithFocused(true))

	return t

}

func listFiles() []string {
	fsd, err := os.ReadDir(config.FileFolder)
	if err != nil {
		log.Debug(err.Error())
	}
	out := make([]string, 0, len(fsd))
	for _, f := range fsd {
		if filepath.Ext(f.Name()) != ".pea" {
			continue
		}
		b, _, _ := strings.Cut(f.Name(), ".")
		out = append(out, b)
	}

	return out
}
