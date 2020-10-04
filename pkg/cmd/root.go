package cmd

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// NewCmdRoot return a cobra root command.
func NewCmdRoot(streams genericclioptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "kubectl-fuzzy",
		Short:                 "Fuzzy Finder kubectl",
		SilenceUsage:          true,
		DisableFlagsInUseLine: true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := c.Usage(); err != nil {
				return err
			}

			return nil
		},
	}

	globalConfig := &globalConfig{
		configFlags: genericclioptions.NewConfigFlags(true),
		streams:     streams,
	}

	globalConfig.configFlags.AddFlags(cmd.PersistentFlags())

	return AddSubCmd(cmd, globalConfig)
}

type globalConfig struct {
	configFlags *genericclioptions.ConfigFlags
	streams     genericclioptions.IOStreams
}

// AddSubCmd to be added sub command for root command.
func AddSubCmd(cmd *cobra.Command, config *globalConfig) *cobra.Command {
	cmd.AddCommand(NewCmdLogs(config.configFlags, config.streams))
	cmd.AddCommand(NewCmdExec(config.configFlags, config.streams))
	cmd.AddCommand(NewCmdDescribe(config.configFlags, config.streams))
	cmd.AddCommand(NewCmdCreate(config.configFlags, config.streams))
	cmd.AddCommand(NewCmdDelete(config.configFlags, config.streams))
	cmd.AddCommand(NewCmdVersion())

	return cmd
}
