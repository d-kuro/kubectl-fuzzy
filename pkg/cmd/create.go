package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/kubectl/pkg/scheme"
)

// NewCmdCreate provides a cobra command wrapping CreateOptions.
func NewCmdCreate(streams genericclioptions.IOStreams) *cobra.Command {
	o := NewCreateOptions(streams)

	cmd := &cobra.Command{
		Use:           "create",
		Short:         "Create a resource",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(c *cobra.Command, args []string) error {
			return fmt.Errorf("must specify resource, only supported job")
		},
	}

	flags := cmd.Flags()
	o.configFlags.AddFlags(flags)
	o.AddFlags(flags)

	cmd.AddCommand(NewCmdCreateJob(streams))

	return cmd
}

// CreateOptions provides information required to update
// the current context on a user's KUBECONFIG.
type CreateOptions struct {
	configFlags *genericclioptions.ConfigFlags
	printFlags  *genericclioptions.PrintFlags
	genericclioptions.IOStreams
}

// NewCreateOptions provides an instance of CreateOptions with default values.
func NewCreateOptions(streams genericclioptions.IOStreams) *CreateOptions {
	return &CreateOptions{
		configFlags: genericclioptions.NewConfigFlags(true),
		printFlags:  genericclioptions.NewPrintFlags("created").WithTypeSetter(scheme.Scheme),
		IOStreams:   streams,
	}
}

// AddFlags adds a flag to the flag set.
func (o *CreateOptions) AddFlags(flags *pflag.FlagSet) {
}

// Complete sets all information required for create.
func (o *CreateOptions) Complete(cmd *cobra.Command, args []string) error {
	return nil
}

// Validate ensures that all required arguments and flag values are provided.
func (o *CreateOptions) Validate() error {
	return nil
}

// Run execute no action.
func (o *CreateOptions) Run(ctx context.Context) error {
	return nil
}
