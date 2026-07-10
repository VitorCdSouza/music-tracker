package ui

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/vitorcds/music-tracker/internal/downloader"
)

var progressRegex = regexp.MustCompile(`Total Query Progress:\s*(\d+)/(\d+)`)

type DownloadModel struct {
	lines   []string
	done    bool
	progBar progress.Model
	percent float64
	err     error
}

func NewDownloadModel() DownloadModel {
	return DownloadModel{
		lines:   []string{},
		progBar: progress.New(progress.WithDefaultGradient()),
	}
}

func (model DownloadModel) Update(msg tea.Msg) (DownloadModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case progress.FrameMsg:
		progressModel, cmd := model.progBar.Update(msg)
		model.progBar = progressModel.(progress.Model)
		cmds = append(cmds, cmd)

	case downloader.LineMsg:
		line := string(msg)
		model.lines = append(model.lines, string(msg))
		if len(model.lines) > 20 {
			model.lines = model.lines[len(model.lines)-20:]
		}

		matches := progressRegex.FindStringSubmatch(line)
		if len(matches) == 3 {
			current, err1 := strconv.ParseFloat(matches[1], 64)
			total, err2 := strconv.ParseFloat(matches[2], 64)

			if err1 == nil && err2 == nil && total > 0 {
				model.percent = current / total
				cmds = append(cmds, model.progBar.SetPercent(model.percent))
			}

		}

	case downloader.DownloadDoneMsg:
		model.err = msg.Err
		model.done = true

		if model.err == nil {
			model.percent = 1.0
			cmds = append(cmds, model.progBar.SetPercent(model.percent))
		}
	}

	return model, tea.Batch(cmds...)
}

func (model DownloadModel) View() string {
	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString(model.progBar.View())
	sb.WriteString("\n")

	start := 0
	maxLines := 5

	if len(model.lines) > maxLines {
		start = len(model.lines) - maxLines
	}
	for _, line := range model.lines[start:] {
		sb.WriteString(" ")
		sb.WriteString(line)
		sb.WriteString("\n")
	}

	if model.done {
		if model.err != nil {
			return "erro ao baixar: " + sb.String() + "\n\nenter para voltar"
		}
		return "finalizado: \n" + sb.String() + "\n\nenter para voltar"

	}

	return "baixando: \n\n" + sb.String() + "\n"
}

func extractPercent(line string) float64 {
	matches := progressRegex.FindStringSubmatch(line)
	if len(matches) > 1 {
		value, err := strconv.ParseFloat(matches[1], 64)
		if err != nil {
			return value / 100
		}
	}
	return -1.0
}
