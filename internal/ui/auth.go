package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/vitorcds/music-tracker/internal/bridge"
	"github.com/vitorcds/music-tracker/internal/config"
)

type playerOption struct {
	ID    string
	Label string
}

type AuthModel struct {
	cursor         int
	players        []playerOption
	cfg            config.AppConfig
	bridgeFeedback chan string

	state int
	logs  []string
}

func NewAuthModel(cfg config.AppConfig, authChan chan string) AuthModel {
	return AuthModel{
		players: []playerOption{
			{ID: "spotify", Label: "spotify"},
			{ID: "youtube", Label: "youtube (em desenvolvimento)"},
		},
		cfg:            cfg,
		bridgeFeedback: authChan,
		state:          0,
		logs:           []string{},
	}
}

func (model AuthModel) Init() tea.Cmd {
	return nil
}

func (model AuthModel) Update(msg tea.Msg) (AuthModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case bridge.LineMsg:
		model.logs = append(model.logs, string(msg))
		if len(model.logs) > 15 {
			model.logs = model.logs[len(model.logs)-15:]
		}
		return model, nil

	case tea.KeyMsg:
		if model.state == 0 {
			switch msg.String() {
			case "up":
				if model.cursor > 0 {
					model.cursor--
				}
			case "down":
				if model.cursor < len(model.players)-1 {
					model.cursor++
				}

			case "enter":
				model.state = 1
				model.cfg.DownloadFrom = model.players[model.cursor].ID

				err := config.SaveConfig(model.cfg)
				if err == nil {
					cmds = append(cmds, func() tea.Msg {
						return ConfigSavedMsg{NewConfig: model.cfg}
					})
				}

				if model.cfg.DownloadFrom == "spotify" {
					provider := bridge.SpotifyProvider{}
					cmds = append(cmds,
						provider.Auth(model.bridgeFeedback),
						provider.ListenForLines(model.bridgeFeedback),
					)
				} else { // TODO ytb auth, dk if it needs one though
					cmds = append(cmds, func() tea.Msg {
						return bridge.AuthDoneMsg{Err: nil}
					})
				}

				return model, tea.Batch(cmds...)
			}
		}
	}
	return model, tea.Batch(cmds...)
}

func (model AuthModel) View() string {
	var sb strings.Builder

	if model.state == 1 {
		sb.WriteString("autenticando... \n\n")
		for _, log := range model.logs {
			sb.WriteString(log)
			sb.WriteString("\n")
		}
		return sb.String()
	}

	sb.WriteString("player:\n")

	for i := 0; i < len(model.players); i++ {
		player := model.players[i]
		cursorStr := " "
		if i == model.cursor {
			cursorStr = "> "
		}
		fmt.Fprintf(&sb, "%s%s\n", cursorStr, player.Label)
	}

	sb.WriteString("\n[j/k] navegar")

	return sb.String()
}
