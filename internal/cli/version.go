package cli

import (
	"fmt"

	"github.com/MonteChristo46/glitch-hunt-cli/assets"

	"github.com/spf13/cobra"
)

func NewVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("huntcli version %s\n", assets.Version())
		},
	}
}
