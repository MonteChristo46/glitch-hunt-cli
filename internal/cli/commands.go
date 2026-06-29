package cli

import (
	"fmt"
	"os"
"github.com/MonteChristo46/glitch-hunt-cli/assets"

	"github.com/MonteChristo46/glitch-hunt-cli/internal/config"

	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not load config: %v\n", err)
		cfg = config.Defaults()
	}

	var rootCmd = &cobra.Command{
		Use:     "huntcli",
		Short:   "Glitch Hunt Development Bridge CLI",
		Long:    assets.Banner() + "\n\nGlitch Hunt Development Bridge — locally receive and test cloud events before deploying.",
		Version: assets.Version(),
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if cmd.Use == "help" || cmd.Use == "huntcli" {
				return
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	rootCmd.AddCommand(
		InstallCmd(),
		LoginCmd(cfg),
		ListenCmd(cfg),
		TriggerCmd(cfg),
		NewVersionCmd(),
	)

	return rootCmd
}

func Execute() {
	rootCmd := NewRootCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
