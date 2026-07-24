package bridge

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/vitorcds/music-tracker/internal/config"
	"github.com/vitorcds/music-tracker/internal/downloader"
)

type SpotifyProvider struct{}

type LineMsg string
type AuthDoneMsg struct {
	Err error
}
type ScrapDoneMsg struct {
	PlaylistName string
	IDs          []string
	Err          error
}

func (sp SpotifyProvider) Auth(lineChan chan string) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command(
			"python3", "-u", "../../internal/scripts/login.py",
		)

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return AuthDoneMsg{Err: err}
		}

		stderr, err := cmd.StderrPipe()
		if err != nil {
			return AuthDoneMsg{Err: err}
		}

		if err := cmd.Start(); err != nil {
			lineChan <- "erro ao iniciar script: " + err.Error()
			return AuthDoneMsg{Err: err}
		}

		multi := io.MultiReader(stdout, stderr)
		scanner := bufio.NewScanner(multi)

		for scanner.Scan() {
			line := scanner.Text()

			if line != "" {
				lineChan <- line
			}
		}

		if err := scanner.Err(); err != nil {
			lineChan <- "erro: " + err.Error()
		}
		err = cmd.Wait()
		if err != nil {
			lineChan <- "script python finalizou com erro: " + err.Error()
		}

		close(lineChan)
		return AuthDoneMsg{Err: err}
	}
}

func (s SpotifyProvider) ListenForLines(sub chan string) tea.Cmd {
	return func() tea.Msg {
		if line, ok := <-sub; ok {
			return LineMsg(line)
		}

		return nil
	}
}

func (s SpotifyProvider) HasCredentials() bool {
	if _, err := os.Stat("credentials.json"); err == nil {
		return true
	} else if errors.Is(err, os.ErrNotExist) {
		return false
	}

	return false
}

func (s SpotifyProvider) ScrapOnline(url string, lineChan chan string) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command(
			"python3", "-u", "../../internal/scripts/scraper.py",
			"credentials.json", url,
		)

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return ScrapDoneMsg{Err: err}
		}

		stderr, err := cmd.StderrPipe()
		if err != nil {
			return ScrapDoneMsg{Err: err}
		}

		if err := cmd.Start(); err != nil {
			lineChan <- "erro ao iniciar script: " + err.Error()
			return ScrapDoneMsg{Err: err}
		}

		multi := io.MultiReader(stdout, stderr)
		scanner := bufio.NewScanner(multi)

		var jsonRaw strings.Builder
		recordingJson := false

		for scanner.Scan() {
			line := scanner.Text()

			if strings.TrimSpace(line) == "" {
				continue
			}

			if line == "json" {
				recordingJson = true
				continue
			}

			if recordingJson {
				jsonRaw.WriteString(line)
			} else {
				lineChan <- line
			}
		}
		if err := scanner.Err(); err != nil {
			lineChan <- "erro: " + err.Error()
		}

		type ScaperReturn struct {
			PlaylistName string `json:"playlist"`
			ID           string `json:"spotify_id"`
		}

		var items []ScaperReturn

		err = json.Unmarshal([]byte(jsonRaw.String()), &items)
		if err != nil {
			lineChan <- "erro ao ler json: " + err.Error()
			close(lineChan)
			return ScrapDoneMsg{Err: err}
		}

		var playlistName string
		var musicIds []string
		for _, item := range items {
			if item.PlaylistName != "" {
				playlistName = item.PlaylistName
			}

			if item.ID != "" {
				musicIds = append(musicIds, item.ID)
			}
		}

		err = cmd.Wait()
		if err != nil {
			lineChan <- "script python finalizou com erro: " + err.Error()
		}

		return ScrapDoneMsg{PlaylistName: playlistName, IDs: musicIds, Err: nil}
	}

}

func (s SpotifyProvider) Download(playlistName string, ids []string, lineChan chan string, cfg config.AppConfig) tea.Cmd {
	return func() tea.Msg {

		cmd := exec.Command(
			"python3", "-u", "../../internal/scripts/downloader.py",
			"credentials.json", cfg.DownloadPath, cfg.AudioQuality, playlistName,
		)

		idsUnified := strings.Join(ids, "\n")
		cmd.Stdin = strings.NewReader(idsUnified)

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return downloader.DownloadDoneMsg{Err: err}
		}

		stderr, err := cmd.StderrPipe()
		if err != nil {
			return downloader.DownloadDoneMsg{Err: err}
		}

		if err := cmd.Start(); err != nil {
			lineChan <- "erro ao iniciar script: " + err.Error()
			return downloader.DownloadDoneMsg{Err: err}
		}

		multi := io.MultiReader(stdout, stderr)
		scanner := bufio.NewScanner(multi)

		for scanner.Scan() {
			line := scanner.Text()

			if line != "" {
				lineChan <- line
			}
		}

		if err := scanner.Err(); err != nil {
			lineChan <- "erro: " + err.Error()
		}
		err = cmd.Wait()
		if err != nil {
			lineChan <- "script python finalizou com erro: " + err.Error()
		}

		close(lineChan)
		return downloader.DownloadDoneMsg{Err: err}
	}
}
