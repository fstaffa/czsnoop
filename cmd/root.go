package cmd

import (
	"context"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

var verboseFlag bool
var logger *slog.Logger

var rootCmd = &cobra.Command{
	Use:   "czsnoop",
	Short: "Search OSINT data specific for the Czech Republic",
	Long: `Search OSINT data specific for the Czech Republic. Uses:
https://www.rzp.cz
https://www.justice.cz`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {

		level := slog.LevelInfo
		if verboseFlag {
			level = slog.LevelDebug
		}
		logger = slog.New(slog.NewTextHandler(cmd.ErrOrStderr(), &slog.HandlerOptions{Level: level}))
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.ExecuteContext(context.Background())
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&verboseFlag, "debug", false, "Enable verbose mode")
}
