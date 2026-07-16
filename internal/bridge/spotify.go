package bridge

import (
	"bufio"
	"errors"
	"io"
	"os"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
)

type AuthLineMsg string
type AuthDoneMsg struct {
	Err error
}

func SpotifyAuth(lineChan chan string) tea.Cmd {
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
		close(lineChan)

		if err != nil {
			lineChan <- "script python finalizou com erro: " + err.Error()
		}
		return AuthDoneMsg{Err: err}
	}
}

func ListenForAuthLines(sub chan string) tea.Cmd {
	return func() tea.Msg {
		if line, ok := <-sub; ok {
			return AuthLineMsg(line)
		}

		return nil
	}
}

func HasCredentials() bool {
	if _, err := os.Stat("credentials.json"); err == nil {
		return true
	} else if errors.Is(err, os.ErrNotExist) {
		return false
	}

	return false
}
