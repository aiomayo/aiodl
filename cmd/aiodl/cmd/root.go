package cmd

import (
	"os"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/aiomayo/aiodl/internal/adapter"
	"github.com/aiomayo/aiodl/internal/config"
	"github.com/aiomayo/aiodl/internal/tui"
)

var (
	v  = viper.New()
	ui *tui.UI
)

var rootCmd = &cobra.Command{
	Use:           "aiodl",
	Short:         "A unified media downloader",
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return initConfig()
	},
}

var (
	cfgFile       string
	noInteractive bool
)

var version string

func SetVersion(v string) {
	version = v
}

func init() {
	log.SetReportTimestamp(false)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
	rootCmd.PersistentFlags().Bool("verbose", false, "verbose output")
	rootCmd.PersistentFlags().BoolVar(&noInteractive, "no-interactive", false, "disable interactive mode")

	_ = v.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))

	rootCmd.AddCommand(newDownloadCmd())
	rootCmd.AddCommand(newInfoCmd())
	rootCmd.AddCommand(newListCmd())
	rootCmd.AddCommand(newConfigCmd())
}

func initConfig() error {
	if noInteractive {
		ui = tui.NewWithMode(tui.ModeNonInteractive)
	} else {
		ui = tui.New()
	}

	if cfgFile != "" {
		v.SetConfigFile(cfgFile)
	}

	if _, err := config.Load(v); err != nil {
		return err
	}

	_ = adapter.Register(adapter.NewYouTubeAdapter())
	return nil
}

func Execute() {
	rootCmd.Version = version

	if err := rootCmd.Execute(); err != nil {
		log.Error("Command failed", "err", err)
		os.Exit(1)
	}
}
