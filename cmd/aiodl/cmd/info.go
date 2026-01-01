package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/aiomayo/aiodl/internal/adapter"
	"github.com/aiomayo/aiodl/internal/tui/views"
)

func newInfoCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "info [URL]",
		Short: "Display information about media",
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

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(info)
			}

			_, _ = fmt.Print(views.RenderInfo(info))

			if info.Type == adapter.MediaTypePlaylist && len(info.Items) > 0 {
				_, _ = fmt.Printf("\nPlaylist videos:\n")
				_, _ = fmt.Print(views.RenderPlaylistItems(info.Items, 5))
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output as JSON")

	return cmd
}
