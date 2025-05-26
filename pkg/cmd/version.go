package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Version indicates the semver of the kubectl-fuzzy command.
const Version = "v1.9.0"

// Revision indicates the revision of git where kubectl-fuzzy was built.
// This is dynamically embedded with the -ldflags option when the binary is built.
// The default display when not embedding is "development".
var Revision = "development" //nolint:gochecknoglobals

// NewCmdVersion is return version command.
func NewCmdVersion() *cobra.Command {
	return &cobra.Command{
		Use:                   "version",
		Short:                 "Show version",
		SilenceUsage:          true,
		DisableFlagsInUseLine: true,
		Run: func(cmd *cobra.Command, args []string) {
			_, _ = fmt.Fprintf(os.Stdout, "version: %s (rev: %s)\n", Version, Revision)
		},
	}
}
