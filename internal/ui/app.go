package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/vitorcds/music-tracker/internal/bridge"
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
	screenAuth
)

type AppModel struct {
	mode    appMode
	current screen

	navbar   NavbarModel
	search   SearchModel
	download DownloadModel
	files    FilesModel
	settings SettingsModel
	auth     AuthModel

	config       config.AppConfig
	lineChan     chan string
	authLineChan chan string
}

func NewAppModel(cfg config.AppConfig) AppModel {
	initialScreen := screenSearch
	if cfg.DownloadFrom == "" || (cfg.DownloadFrom == "spotify" && !bridge.HasCredentials()) {
		initialScreen = screenAuth
	}

	authChan := make(chan string)

	return AppModel{
		mode:    modeNormal,
		current: initialScreen,

		navbar:   NewNavbarModel(),
		search:   NewSearchModel(),
		download: NewDownloadModel(),
		files:    NewFilesModel(cfg.DownloadPath),
		settings: NewSettingsModel(cfg),
		auth:     NewAuthModel(cfg, authChan),

		config:       cfg,
		lineChan:     make(chan string),
		authLineChan: authChan,
	}
}

func (model AppModel) Init() tea.Cmd {
	return nil
}

func (model AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	// key press
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "ctrl+c":
			return model, tea.Quit
		case "esc":
			model.mode = modeNormal
			model.search.TextInput.Cursor.SetMode(cursor.CursorStatic)
			for i := range model.settings.inputs {
				model.settings.inputs[i].Cursor.SetMode(cursor.CursorStatic)
			}
			return model, nil
		}

		if model.mode == modeNormal {
			switch keyMsg.String() {
			case "i":
				if model.current == screenFiles || model.current == screenAuth {
					return model, nil
				}
				model.mode = modeInput
				model.search.TextInput.Cursor.SetMode(cursor.CursorBlink)
				for i := range model.settings.inputs {
					model.settings.inputs[i].Cursor.SetMode(cursor.CursorBlink)
				}

				return model, textinput.Blink

			case "L":
				if model.current == screenAuth {
					return model, nil
				}

				if int(model.current) < len(model.navbar.Tabs)-1 {
					model.current = screen(int(model.current) + 1)
				}

				if model.current == screenFiles {
					model.files = model.files.Reload()
				}
				return model, nil
			case "H":
				if model.current == screenAuth {
					return model, nil
				}
				if model.current > 0 {
					model.current = screen(int(model.current) - 1)
				}
				return model, nil

			case "j":
				msg = tea.KeyMsg{Type: tea.KeyDown}
			case "k":
				msg = tea.KeyMsg{Type: tea.KeyUp}
			case "up", "down", "ctrl+s", "tab", "enter", "backspace":
			default:
				return model, nil
			}
		}
	}

	// comunication with other models/files
	switch msg := msg.(type) {
	case ConfigSavedMsg:
		model.config = msg.NewConfig
		model.mode = modeNormal
		model.files = NewFilesModel(model.config.DownloadPath)

		model.settings = NewSettingsModel(model.config)
		return model, nil

	case bridge.AuthLineMsg:
		model.auth, cmd = model.auth.Update(msg)
		cmds = append(cmds, cmd, bridge.ListenForAuthLines(model.authLineChan))
		return model, tea.Batch(cmds...)

	case bridge.AuthDoneMsg:
		if msg.Err != nil {
			return model, nil
		}
		model.current = screenSearch
		return model, nil

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
		model.files = model.files.Reload()
		return model, tea.Batch(cmds...)
	}

	// current screen handler
	switch model.current {
	case screenSearch:
		if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "enter" {
			url := model.search.TextInput.Value()
			model.current = screenDownloading
			model.download = NewDownloadModel()

			model.lineChan = make(chan string)

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

	case screenFiles:
		model.files, cmd = model.files.Update(msg)
		cmds = append(cmds, cmd)

	case screenConfig:
		model.settings, cmd = model.settings.Update(msg)
		cmds = append(cmds, cmd)

	case screenAuth:
		model.auth, cmd = model.auth.Update(msg)
		cmds = append(cmds, cmd)

	}

	return model, tea.Batch(cmds...)
}

func (model AppModel) View() string {
	var sb strings.Builder

	if model.current == screenAuth {
		return model.auth.View()
	}

	activeTab := int(model.current)
	if model.current == screenDownloading {
		activeTab = 0
	}

	sb.WriteString(model.navbar.View(activeTab))
	sb.WriteString("\n\n")

	switch model.current {
	case screenSearch:
		sb.WriteString(model.search.View())
	case screenDownloading:
		sb.WriteString(model.download.View())
	case screenFiles:
		sb.WriteString(model.files.View())
	case screenConfig:
		sb.WriteString(model.settings.View())
	default:
		return ""
	}

	sb.WriteString("\n\n")
	if model.mode == modeNormal {
		sb.WriteString("normal")
	} else {
		sb.WriteString("input")
	}
	return sb.String()
}
