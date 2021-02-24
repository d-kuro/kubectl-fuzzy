package cmd

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

const (
	previewEnabledEnvVar = "KUBE_FUZZY_PREVIEW_ENABLED"
)

// NewCmdRoot return a cobra root command.
func NewCmdRoot(streams genericclioptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "kubectl-fuzzy",
		Short:                 "Fuzzy Finder kubectl",
		SilenceUsage:          true,
		SilenceErrors:         true,
		DisableFlagsInUseLine: true,
		RunE: func(c *cobra.Command, args []string) error {
			return c.Usage()
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
