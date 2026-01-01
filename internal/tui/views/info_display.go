package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/aiomayo/aiodl/internal/adapter"
	"github.com/aiomayo/aiodl/internal/tui"
)

func RenderInfo(info *adapter.MediaInfo) string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true)
	labelStyle := lipgloss.NewStyle().Faint(true).Width(12)
	faintStyle := lipgloss.NewStyle().Faint(true)

	addRow := func(label, value string) {
		b.WriteString(labelStyle.Render(label+":") + " " + value + "\n")
	}

	b.WriteString(labelStyle.Render("Title:") + " " + titleStyle.Render(info.Title) + "\n")

	mediaType := "video"
	switch info.Type {
	case adapter.MediaTypeAudio:
		mediaType = "audio"
	case adapter.MediaTypePlaylist:
		mediaType = "playlist"
	}
	addRow("Type", mediaType)
	addRow("Platform", info.Platform)

	if info.Duration > 0 {
		addRow("Duration", tui.FormatDuration(info.Duration))
	}

	if info.Type == adapter.MediaTypePlaylist && len(info.Items) > 0 {
		addRow("Videos", fmt.Sprintf("%d items", len(info.Items)))
	}

	if len(info.Formats) > 0 {
		addRow("Formats", fmt.Sprintf("%d available", len(info.Formats)))

		bestVideo, bestAudio := findBestFormats(info.Formats)
		if bestVideo != nil {
			addRow("Best video", fmt.Sprintf("%s (%s, %s)",
				bestVideo.Quality, bestVideo.Extension, tui.FormatSize(bestVideo.FileSize)))
		}
		if bestAudio != nil {
			quality := tui.FormatBitrate(bestAudio.Bitrate)
			if bestAudio.Quality != "" {
				quality = bestAudio.Quality
			}
			addRow("Best audio", fmt.Sprintf("%s (%s, %s)",
				quality, bestAudio.Extension, tui.FormatSize(bestAudio.FileSize)))
		}
	}

	b.WriteString(labelStyle.Render("URL:") + " " + faintStyle.Render(info.URL) + "\n")

	return b.String()
}

func findBestFormats(formats []adapter.Format) (*adapter.Format, *adapter.Format) {
	var bestVideo, bestAudio *adapter.Format

	for i := range formats {
		f := &formats[i]
		if f.IsVideo() {
			if bestVideo == nil || f.Height > bestVideo.Height {
				bestVideo = f
			}
		} else {
			if bestAudio == nil || f.Bitrate > bestAudio.Bitrate {
				bestAudio = f
			}
		}
	}

	return bestVideo, bestAudio
}

func RenderFormats(title string, formats []adapter.Format) string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true)
	headerStyle := lipgloss.NewStyle().Faint(true)

	b.WriteString(titleStyle.Render(title) + "\n\n")
	b.WriteString(headerStyle.Render(
		fmt.Sprintf("%-8s %-8s %-12s %-6s %-10s %-10s",
			"ID", "Type", "Quality", "Ext", "Size", "Bitrate")) + "\n")
	b.WriteString(headerStyle.Render(strings.Repeat("─", 60)) + "\n")

	for _, f := range formats {
		ftype := "audio"
		if f.IsVideo() {
			ftype = "video"
		}
		b.WriteString(fmt.Sprintf("%-8s %-8s %-12s %-6s %-10s %-10s\n",
			f.ID, ftype, f.Quality, f.Extension,
			tui.FormatSize(f.FileSize), tui.FormatBitrate(f.Bitrate)))
	}

	return b.String()
}

func RenderPlaylistItems(items []adapter.MediaInfo, maxItems int) string {
	var b strings.Builder

	faintStyle := lipgloss.NewStyle().Faint(true)

	shown := items
	if maxItems > 0 && len(items) > maxItems {
		shown = items[:maxItems]
	}

	for i, item := range shown {
		idx := fmt.Sprintf("%3d.", i+1)
		duration := ""
		if item.Duration > 0 {
			duration = " " + faintStyle.Render(fmt.Sprintf("(%s)", tui.FormatDuration(item.Duration)))
		}
		b.WriteString(idx + " " + item.Title + duration + "\n")
	}

	if maxItems > 0 && len(items) > maxItems {
		remaining := len(items) - maxItems
		b.WriteString(faintStyle.Render(fmt.Sprintf("    ... and %d more", remaining)) + "\n")
	}

	return b.String()
}
