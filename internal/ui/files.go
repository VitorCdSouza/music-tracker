package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type FilesModel struct {
	Tabs []string
}

func NewFilesModel() FilesModel {
	return FilesModel{Tabs: []string{"busca", "arquivos", "config"}}
}

func (model FilesModel) Init() tea.Cmd {
	return textinput.Blink
}

func (model FilesModel) Update(msg tea.Msg) (FilesModel, tea.Cmd) {
	var cmd tea.Cmd
	return model, cmd
}

func (model FilesModel) View() string {
	var sb strings.Builder

	sb.WriteString("arquivos:")

	return sb.String()
}
