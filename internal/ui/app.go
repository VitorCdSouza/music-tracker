package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/vitorcds/music-tracker/internal/config"
	"github.com/vitorcds/music-tracker/internal/downloader"
)

type appMode int

const (
	modeNormal appMode = iota
	modeInput
)

type screen int

const (
	screenSearch screen = iota
	screenFiles
	screenConfig
	screenDownloading
)

type AppModel struct {
	mode     appMode
	current  screen
	navbar   NavbarModel
	search   SearchModel
	download DownloadModel
	files    FilesModel
	config   config.AppConfig
	lineChan chan string
}

func NewAppModel(cfg config.AppConfig) AppModel {
	return AppModel{
		mode:     modeNormal,
		current:  screenSearch,
		navbar:   NewNavbarModel(),
		search:   NewSearchModel(),
		download: NewDownloadModel(),
		files:    NewFilesModel(),
		config:   cfg,
		lineChan: make(chan string),
	}
}

func (model AppModel) Init() tea.Cmd {
	return model.search.Init()
}

func (model AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "ctrl+c":
			return model, tea.Quit
		case "esc":
			model.mode = modeNormal
			model.search.TextInput.Blur()
			return model, nil
		}

		if model.mode == modeNormal {
			switch keyMsg.String() {
			case "i":
				if model.current == screenSearch {
					model.mode = modeInput
					return model, model.search.TextInput.Focus()
				}

			case "L":
				if model.current < 2 {
					model.current = screen(int(model.current) + 1)
				}
				return model, nil
			case "H":
				if model.current > 0 {
					model.current = screen(int(model.current) - 1)
					return model, nil
				}
			}

			return model, nil
		}
	}

	switch msg := msg.(type) {
	case progress.FrameMsg:
		if model.current == screenDownloading {
			model.download, cmd = model.download.Update(msg)
			cmds = append(cmds, cmd)
			return model, tea.Batch(cmds...)
		}

	case downloader.LineMsg:
		model.download, cmd = model.download.Update(msg)
		cmds = append(cmds, cmd, downloader.ListenForLines(model.lineChan))
		return model, tea.Batch(cmds...)

	case downloader.DownloadDoneMsg:
		model.download, cmd = model.download.Update(msg)
		cmds = append(cmds, cmd)
		return model, tea.Batch(cmds...)
	}

	switch model.current {
	case screenSearch:
		if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "enter" {
			url := model.search.TextInput.Value()
			model.current = screenDownloading
			model.download = NewDownloadModel()

			cmds = append(cmds,
				downloader.StartDownload(url, model.lineChan, model.config),
				downloader.ListenForLines(model.lineChan),
			)
			return model, tea.Batch(cmds...)
		}
		model.search, cmd = model.search.Update(msg)
		cmds = append(cmds, cmd)

	case screenDownloading:
		if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "enter" {
			model.current = screenSearch
			model.search.TextInput.SetValue("")
			return model, model.search.Init()
		}

		model.download, cmd = model.download.Update(msg)
		cmds = append(cmds, cmd)
	}

	return model, tea.Batch(cmds...)
}

func (model AppModel) View() string {
	var sb strings.Builder

	activeTab := int(model.current)
	if model.current == screenDownloading {
		activeTab = 0
	}

	sb.WriteString(model.navbar.View(activeTab))
	sb.WriteString("\n")

	switch model.current {
	case screenSearch:
		sb.WriteString(model.search.View())
	case screenDownloading:
		sb.WriteString(model.download.View())
	case screenFiles:
		sb.WriteString(model.files.View())
	case screenConfig:
		sb.WriteString("Desenvolvimento")
	default:
		return ""
	}
	return sb.String()
}
