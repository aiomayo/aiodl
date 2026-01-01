package cmd

import (
	"context"
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"

	"github.com/aiomayo/aiodl/internal/adapter"
	"github.com/aiomayo/aiodl/internal/tui/views"
)

func newListCmd() *cobra.Command {
	var interactive bool

	cmd := &cobra.Command{
		Use:   "list [URL]",
		Short: "List available formats",
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
				return err
			}

			if len(info.Formats) == 0 {
				log.Warn("No formats available")
				return nil
			}

			if interactive && ui.IsInteractive() {
				selector := views.NewFormatSelector(
					fmt.Sprintf("Formats for: %s", info.Title),
					info.Formats,
				)
				_, err := ui.Run(selector)
				return err
			}

			_, _ = fmt.Print(views.RenderFormats(
				fmt.Sprintf("Formats for: %s", info.Title),
				info.Formats,
			))
			return nil
		},
	}

	cmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "interactive format browser")

	return cmd
}
