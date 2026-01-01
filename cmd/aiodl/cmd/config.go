package cmd

import (
	"fmt"
	"os"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/aiomayo/aiodl/internal/config"
)

func newConfigCmd() *cobra.Command {
	var pathOnly bool

	cmd := &cobra.Command{
		Use:   "config",
		Short: "Show current configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			if pathOnly {
				configPath := v.ConfigFileUsed()
				if configPath == "" {
					configPath = config.FilePath()
					log.Warn("No config loaded", "default", configPath)
				} else {
					_, _ = fmt.Fprintln(os.Stdout, configPath)
				}
				return nil
			}

			data, err := yaml.Marshal(v.AllSettings())
			if err != nil {
				return fmt.Errorf("failed to marshal config: %w", err)
			}
			_, _ = fmt.Fprint(os.Stdout, string(data))
			return nil
		},
	}

	cmd.Flags().BoolVar(&pathOnly, "path", false, "show config file path only")

	return cmd
}
