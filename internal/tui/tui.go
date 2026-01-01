package tui

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type Mode int

const (
	ModeInteractive Mode = iota
	ModeNonInteractive
)

type UI struct {
	mode Mode
}

func New() *UI {
	return &UI{mode: detectMode()}
}

func NewWithMode(mode Mode) *UI {
	return &UI{mode: mode}
}

func detectMode() Mode {
	if !isTerminal(os.Stdin) || !isTerminal(os.Stdout) {
		return ModeNonInteractive
	}
	if os.Getenv("CI") != "" || os.Getenv("NO_COLOR") != "" {
		return ModeNonInteractive
	}
	return ModeInteractive
}

func isTerminal(f *os.File) bool {
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

func (u *UI) IsInteractive() bool {
	return u.mode == ModeInteractive
}

func (u *UI) Run(model tea.Model) (tea.Model, error) {
	if u.mode == ModeNonInteractive {
		return model, nil
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithContext(ctx))
	return p.Run()
}

func FormatSize(size int64) string {
	if size <= 0 {
		return "-"
	}
	return FormatBytes(size)
}

func FormatBitrate(bitrate int) string {
	if bitrate <= 0 {
		return "-"
	}
	return fmt.Sprintf("%d kbps", bitrate/1000)
}

func FormatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return strconv.FormatInt(b, 10) + " B"
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}

func FormatDuration(seconds int) string {
	if seconds < 60 {
		return strconv.Itoa(seconds) + "s"
	}
	if seconds < 3600 {
		m := seconds / 60
		s := seconds % 60
		if s == 0 {
			return strconv.Itoa(m) + "m"
		}
		return strconv.Itoa(m) + "m " + strconv.Itoa(s) + "s"
	}
	h := seconds / 3600
	m := (seconds % 3600) / 60
	if m == 0 {
		return strconv.Itoa(h) + "h"
	}
	return strconv.Itoa(h) + "h " + strconv.Itoa(m) + "m"
}

func FormatDurationTime(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	if d < time.Hour {
		m := int(d.Minutes())
		s := int(d.Seconds()) % 60
		return fmt.Sprintf("%dm %ds", m, s)
	}
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	return fmt.Sprintf("%dh %dm", h, m)
}
