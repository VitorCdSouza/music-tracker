package bridge

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/vitorcds/music-tracker/internal/config"
)

type Provider interface {
	Auth(line chan string) tea.Cmd
	HasCredentials() bool
	ScrapOnline(url string, line chan string) tea.Cmd
	Download(ids []string, line chan string, cfg config.AppConfig) tea.Cmd
	ListenForLines(line chan string) tea.Cmd
}
