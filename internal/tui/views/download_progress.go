package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/aiomayo/aiodl/internal/tui"
)

type DownloadState int

const (
	StateInitializing DownloadState = iota
	StateDownloading
	StateComplete
	StateError
	StateCancelled
)

type DownloadProgressModel struct {
	filename   string
	downloaded int64
	total      int64
	startTime  time.Time
	lastUpdate time.Time
	lastBytes  int64
	speed      float64
	state      DownloadState
	err        error

	progress progress.Model
	spinner  spinner.Model
	help     help.Model
	keys     progressKeyMap
	width    int
	height   int
}

type progressKeyMap struct {
	Quit key.Binding
}

func (k progressKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Quit}
}

func (k progressKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Quit}}
}

var defaultProgressKeys = progressKeyMap{
	Quit: key.NewBinding(key.WithKeys("q", "esc", "ctrl+c"), key.WithHelp("q", "cancel")),
}

type tickMsg time.Time

func NewDownloadProgress(filename string) DownloadProgressModel {
	p := progress.New(progress.WithDefaultGradient(), progress.WithWidth(50))

	s := spinner.New()
	s.Spinner = spinner.Dot

	return DownloadProgressModel{
		filename:  filename,
		state:     StateInitializing,
		progress:  p,
		spinner:   s,
		help:      help.New(),
		keys:      defaultProgressKeys,
		startTime: time.Now(),
	}
}

func (m DownloadProgressModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.tickCmd())
}

func (m DownloadProgressModel) tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m DownloadProgressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width
		m.progress.Width = msg.Width - 10
		if m.progress.Width > 80 {
			m.progress.Width = 80
		}

	case tickMsg:
		if m.state == StateComplete || m.state == StateError || m.state == StateCancelled {
			return m, nil
		}
		return m, m.tickCmd()

	case spinner.TickMsg:
		if m.state == StateInitializing {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil

	case progress.FrameMsg:
		model, cmd := m.progress.Update(msg)
		m.progress = model.(progress.Model)
		return m, cmd

	case tea.KeyMsg:
		if key.Matches(msg, m.keys.Quit) {
			m.state = StateCancelled
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m *DownloadProgressModel) updateSpeed() {
	now := time.Now()
	elapsed := now.Sub(m.lastUpdate).Seconds()
	if elapsed > 0.5 {
		bytesTransferred := m.downloaded - m.lastBytes
		m.speed = float64(bytesTransferred) / elapsed
		m.lastUpdate = now
		m.lastBytes = m.downloaded
	}
}

func (m DownloadProgressModel) View() string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true)
	faintStyle := lipgloss.NewStyle().Faint(true)

	b.WriteString(titleStyle.Render(m.filename) + "\n\n")

	switch m.state {
	case StateInitializing:
		b.WriteString(m.spinner.View() + " Initializing download...\n")

	case StateDownloading:
		percent := 0.0
		if m.total > 0 {
			percent = float64(m.downloaded) / float64(m.total)
		}
		b.WriteString(m.progress.ViewAs(percent) + "\n\n")

		parts := []string{}
		if m.total > 0 {
			parts = append(parts, fmt.Sprintf("%s / %s", tui.FormatBytes(m.downloaded), tui.FormatBytes(m.total)))
		} else {
			parts = append(parts, tui.FormatBytes(m.downloaded))
		}
		if m.speed > 0 {
			parts = append(parts, tui.FormatBytes(int64(m.speed))+"/s")
		}
		if m.speed > 0 && m.total > 0 {
			remaining := float64(m.total-m.downloaded) / m.speed
			if remaining > 0 && remaining < 86400 {
				parts = append(parts, "ETA: "+tui.FormatDuration(int(remaining)))
			}
		}
		b.WriteString(faintStyle.Render(strings.Join(parts, " | ")))
		b.WriteString(fmt.Sprintf("  %.1f%%\n", percent*100))

	case StateComplete:
		b.WriteString(m.progress.ViewAs(1.0) + "\n\n")
		elapsed := time.Since(m.startTime)
		avgSpeed := float64(m.total) / elapsed.Seconds()
		b.WriteString(fmt.Sprintf("Complete! %s in %s (avg: %s/s)\n",
			tui.FormatBytes(m.total), tui.FormatDurationTime(elapsed), tui.FormatBytes(int64(avgSpeed))))

	case StateError:
		percent := 0.0
		if m.total > 0 {
			percent = float64(m.downloaded) / float64(m.total)
		}
		b.WriteString(m.progress.ViewAs(percent) + "\n\n")
		errMsg := "Download failed"
		if m.err != nil {
			errMsg = fmt.Sprintf("Failed: %v", m.err)
		}
		b.WriteString(errMsg + "\n")

	case StateCancelled:
		percent := 0.0
		if m.total > 0 {
			percent = float64(m.downloaded) / float64(m.total)
		}
		b.WriteString(m.progress.ViewAs(percent) + "\n\n")
		b.WriteString(fmt.Sprintf("Cancelled. Downloaded %s of %s\n",
			tui.FormatBytes(m.downloaded), tui.FormatBytes(m.total)))
	}

	b.WriteString("\n")
	b.WriteString(m.help.View(m.keys))

	return b.String()
}

func (m *DownloadProgressModel) SetProgress(downloaded, total int64) {
	m.downloaded = downloaded
	if total > 0 {
		m.total = total
	}
	if m.state == StateInitializing {
		m.state = StateDownloading
	}
	m.updateSpeed()
}

func (m *DownloadProgressModel) SetComplete() {
	m.state = StateComplete
	m.downloaded = m.total
}

func (m *DownloadProgressModel) SetError(err error) {
	m.state = StateError
	m.err = err
}

func (m DownloadProgressModel) IsComplete() bool {
	return m.state == StateComplete
}

func (m DownloadProgressModel) IsError() bool {
	return m.state == StateError
}

func (m DownloadProgressModel) IsCancelled() bool {
	return m.state == StateCancelled
}
