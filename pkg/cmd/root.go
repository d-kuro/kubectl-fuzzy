package cmd

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// NewCmdRoot return a cobra root command.
func NewCmdRoot(streams genericclioptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "kubectl-fuzzy",
		Short:        "Fuzzy Finder kubectl",
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := c.Usage(); err != nil {
				return err
			}
			return nil
		},
	}

	return AddSubCmd(cmd, streams)
}

// AddSubCmd to be added sub command for root command.
func AddSubCmd(cmd *cobra.Command, streams genericclioptions.IOStreams) *cobra.Command {
	cmd.AddCommand(NewCmdLogs(streams))
	cmd.AddCommand(NewCmdExec(streams))
	cmd.AddCommand(NewCmdDescribe(streams))
	cmd.AddCommand(NewCmdVersion())
	cmd.AddCommand(NewCmdCreateJobFrom(streams))

	return cmd
}
