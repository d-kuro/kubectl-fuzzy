package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const Version = "v1.7.0"

var Revision = "development" //nolint:gochecknoglobals

// NewCmdVersion is return version command.
func NewCmdVersion() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(os.Stdout, "version: %s (rev: %s)\n", Version, Revision)
		},
	}
}
