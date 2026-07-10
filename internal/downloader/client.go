package downloader

import (
	"bufio"
	"os/exec"
	"regexp"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/vitorcds/music-tracker/internal/config"
)

type LineMsg string
type DownloadDoneMsg struct {
	Err error
}

var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

func StartDownload(url string, lineChan chan string, cfg config.AppConfig) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command(
			"python3", "-u", "-m", "zotify",
			"--root-path", cfg.DownloadPath,
			"--download-format", cfg.AudioFormat,
			"--download-quality", cfg.AudioQuality,
			"--standard-interface", "true",
			url,
		)

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return DownloadDoneMsg{Err: err}
		}

		if err := cmd.Start(); err != nil {
			return DownloadDoneMsg{Err: err}
		}

		go func() {
			scanner := bufio.NewScanner(stdout)
			for scanner.Scan() {
				rawLine := scanner.Text()

				cleanLine := ansiRegex.ReplaceAllString(rawLine, "")
				cleanLine = strings.ReplaceAll(cleanLine, "\r", "")
				cleanLine = strings.TrimSpace(cleanLine)

				if cleanLine != "" {
					lineChan <- cleanLine
				}
			}
			if err := scanner.Err(); err != nil {
				lineChan <- "Erro: " + err.Error()
			}
		}()

		err = cmd.Wait()
		return DownloadDoneMsg{Err: err}
	}
}

func ListenForLines(sub chan string) tea.Cmd {
	return func() tea.Msg {
		return LineMsg(<-sub)
	}
}
