package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type SettingsModel struct {
	Tabs []string
}

func NewSettingsModel() SettingsModel {
	return SettingsModel{Tabs: []string{"busca", "arquivos", "config"}}
}

func (model SettingsModel) Init() tea.Cmd {
	return textinput.Blink
}

func (model SettingsModel) Update(msg tea.Msg) (SettingsModel, tea.Cmd) {
	var cmd tea.Cmd
	return model, cmd
}

func (model SettingsModel) View() string {
	var sb strings.Builder

	sb.WriteString("arquivos:")

	return sb.String()
}
