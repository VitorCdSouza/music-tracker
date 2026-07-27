package bridge

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/vitorcds/music-tracker/internal/config"
)

type MsgToPython struct {
	Action       string   `json:"action"`
	Url          string   `json:"url"`
	PlaylistName string   `json:"playlistName"`
	Ids          []string `json:"ids"`
	Quality      string   `json:"quality"`
	DownloadPath string   `json:"downloadPath"`
}

type MsgFromPython struct {
	Event        string `json:"event"`
	Message      string `json:"message"`
	PlaylistName string `json:"playlistName"`
}

type Provider interface {
	StartWorker(channel chan MsgFromPython) error

	Auth(line chan string) tea.Cmd
	HasCredentials() bool
	Scrap(url string, line chan string, config config.AppConfig) tea.Cmd
	Download(playlistName string, ids []string, line chan string, cfg config.AppConfig) tea.Cmd
}

func ListenForEvents(sub chan MsgFromPython) tea.Cmd {
	return func() tea.Msg {
		if line, ok := <-sub; ok {
			return line
		}

		return nil
	}
}

func ListenForLines(sub chan string) tea.Cmd {
	return func() tea.Msg {
		if line, ok := <-sub; ok {
			return LineMsg(line)
		}

		return nil
	}
}
