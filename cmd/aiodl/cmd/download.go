package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"

	"github.com/aiomayo/aiodl/internal/adapter"
	"github.com/aiomayo/aiodl/internal/tui"
	"github.com/aiomayo/aiodl/internal/tui/views"
)

func newDownloadCmd() *cobra.Command {
	var (
		format      string
		output      string
		quality     string
		audioOnly   bool
		interactive bool
	)

	cmd := &cobra.Command{
		Use:   "download [URL]",
		Short: "Download media from a URL",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			url := args[0]
			ctx := context.Background()

			adp, found := adapter.Find(url)
			if !found {
				return fmt.Errorf("no adapter for URL: %s", url)
			}

			info, err := adp.GetInfo(ctx, url)
			if err != nil {
				return fmt.Errorf("failed to get media info: %w", err)
			}

			if info.Type == adapter.MediaTypePlaylist {
				return runPlaylistDownload(ctx, adp, info, quality, audioOnly)
			}

			var selectedFormat *adapter.Format
			if format == "" && (interactive || ui.IsInteractive()) && len(info.Formats) > 0 {
				selector := views.NewFormatSelector(
					fmt.Sprintf("Select format for: %s", info.Title),
					info.Formats,
				)

				model, err := ui.Run(selector)
				if err != nil {
					return fmt.Errorf("format selection failed: %w", err)
				}

				result := model.(views.FormatSelectorModel)
				if result.Cancelled() {
					log.Warn("Download cancelled")
					return nil
				}
				selectedFormat = result.Selected()
			}

			downloadOpts := adapter.DownloadOptions{
				Quality:   quality,
				FormatID:  format,
				AudioOnly: audioOnly,
			}

			if selectedFormat != nil {
				downloadOpts.FormatID = selectedFormat.ID
			}

			outputPath := output
			if outputPath == "" {
				ext := "mp4"
				if audioOnly {
					ext = "m4a"
				}
				if selectedFormat != nil && selectedFormat.Extension != "" {
					ext = selectedFormat.Extension
				}
				outputPath = sanitizeFilename(info.Title) + "." + ext
			}

			if ui.IsInteractive() {
				return runInteractiveDownload(ctx, adp, info, downloadOpts, outputPath)
			}
			return runNonInteractiveDownload(ctx, adp, info, downloadOpts, outputPath)
		},
	}

	cmd.Flags().StringVarP(&format, "format", "f", "", "format ID (itag)")
	cmd.Flags().StringVarP(&output, "output", "o", "", "output path")
	cmd.Flags().StringVarP(&quality, "quality", "q", "", "quality (e.g., 720p, 1080p)")
	cmd.Flags().BoolVarP(&audioOnly, "audio-only", "a", false, "audio only")
	cmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "interactively select format")

	return cmd
}

func runInteractiveDownload(ctx context.Context, adp adapter.Adapter,
	info *adapter.MediaInfo, opts adapter.DownloadOptions, outputPath string) error {

	progressView := views.NewDownloadProgress(outputPath)

	type downloadResult struct {
		reader io.ReadCloser
		err    error
	}
	resultCh := make(chan downloadResult, 1)

	go func() {
		reader, err := adp.Download(ctx, info, opts, func(p adapter.DownloadProgress) {
			progressView.SetProgress(p.Downloaded, p.Total)
		})
		resultCh <- downloadResult{reader, err}
	}()

	result := <-resultCh
	if result.err != nil {
		return fmt.Errorf("download failed: %w", result.err)
	}

	written, err := writeToFile(result.reader, outputPath)
	if err != nil {
		return fmt.Errorf("failed to write file %q: %w", outputPath, err)
	}

	log.Info("Downloaded", "file", outputPath, "size", tui.FormatBytes(written))
	return nil
}

func runNonInteractiveDownload(ctx context.Context, adp adapter.Adapter,
	info *adapter.MediaInfo, opts adapter.DownloadOptions, outputPath string) error {

	reader, err := adp.Download(ctx, info, opts, func(p adapter.DownloadProgress) {
		if p.Total > 0 {
			percent := float64(p.Downloaded) / float64(p.Total) * 100
			fmt.Printf("\rProgress: %.1f%% (%s / %s)",
				percent, tui.FormatBytes(p.Downloaded), tui.FormatBytes(p.Total))
		}
	})
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	written, err := writeToFile(reader, outputPath)
	fmt.Println()
	if err != nil {
		return fmt.Errorf("failed to write file %q: %w", outputPath, err)
	}

	log.Info("Downloaded", "file", outputPath, "size", tui.FormatBytes(written))
	return nil
}

func writeToFile(reader io.ReadCloser, outputPath string) (int64, error) {
	defer func() { _ = reader.Close() }()

	file, err := os.Create(outputPath)
	if err != nil {
		return 0, err
	}
	defer func() { _ = file.Close() }()

	written, err := io.Copy(file, reader)
	if err != nil {
		_ = os.Remove(outputPath)
		return 0, err
	}
	return written, nil
}

func sanitizeFilename(name string) string {
	replacer := strings.NewReplacer(
		"/", "-",
		"\\", "-",
		":", "-",
		"*", "",
		"?", "",
		"\"", "",
		"<", "",
		">", "",
		"|", "",
	)
	name = replacer.Replace(name)
	if len(name) > 200 {
		name = name[:200]
	}
	return strings.TrimSpace(name)
}

func runPlaylistDownload(ctx context.Context, adp adapter.Adapter,
	playlist *adapter.MediaInfo, quality string, audioOnly bool) error {
	if len(playlist.Items) == 0 {
		log.Warn("Playlist is empty")
		return nil
	}

	var itemsToDownload []adapter.MediaInfo

	if ui.IsInteractive() {
		browser := views.NewPlaylistBrowser(
			fmt.Sprintf("Select videos from: %s", playlist.Title),
			playlist.Items,
		)

		model, err := ui.Run(browser)
		if err != nil {
			return err
		}

		result := model.(views.PlaylistBrowserModel)
		if result.Cancelled() {
			log.Warn("Download cancelled")
			return nil
		}

		itemsToDownload = result.SelectedItems()
		if len(itemsToDownload) == 0 {
			log.Warn("No videos selected")
			return nil
		}
	} else {
		itemsToDownload = playlist.Items
	}

	total := len(itemsToDownload)
	for i, item := range itemsToDownload {
		log.Info("Downloading", "video", fmt.Sprintf("%d/%d", i+1, total), "title", item.Title)

		videoInfo, err := adp.GetInfo(ctx, item.URL)
		if err != nil {
			log.Warn("Skipping video, failed to get info", "title", item.Title, "err", err)
			continue
		}

		downloadOpts := adapter.DownloadOptions{
			Quality:   quality,
			AudioOnly: audioOnly,
		}

		ext := "mp4"
		if audioOnly {
			ext = "m4a"
		}
		outputPath := sanitizeFilename(item.Title) + "." + ext

		if err := downloadSingleVideo(ctx, adp, videoInfo, downloadOpts, outputPath); err != nil {
			log.Warn("Skipping video, download failed", "title", item.Title, "err", err)
			continue
		}

		log.Info("Downloaded", "file", outputPath)
	}

	log.Info("Playlist download complete", "downloaded", total)
	return nil
}

func downloadSingleVideo(ctx context.Context, adp adapter.Adapter,
	info *adapter.MediaInfo, opts adapter.DownloadOptions, outputPath string) error {
	reader, err := adp.Download(ctx, info, opts, func(p adapter.DownloadProgress) {
		if p.Total > 0 {
			percent := float64(p.Downloaded) / float64(p.Total) * 100
			fmt.Printf("\r  Progress: %.1f%% (%s / %s)",
				percent, tui.FormatBytes(p.Downloaded), tui.FormatBytes(p.Total))
		}
	})
	if err != nil {
		return err
	}

	_, err = writeToFile(reader, outputPath)
	fmt.Println()
	return err
}
