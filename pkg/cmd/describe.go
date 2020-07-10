package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/kubectl/pkg/describe"

	"github.com/d-kuro/kubectl-fuzzy/pkg/fuzzyfinder"
)

const (
	exampleDescribe = `
	# Selecting a Object with the fuzzy finder and view the log and show details
	kubectl fuzzy describe TYPE [flags]
`
)

// NewCmdDescribe provides a cobra command wrapping DescribeOptions.
func NewCmdDescribe(streams genericclioptions.IOStreams) *cobra.Command {
	o := NewDescribeOptions(streams)

	cmd := &cobra.Command{
		Use:           "describe",
		Short:         "Selecting a object with the fuzzy finder and show details",
		Example:       exampleDescribe,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := o.Complete(c, args); err != nil {
				return err
			}

			if err := o.Validate(); err != nil {
				return err
			}

			if err := o.Run(c.Context(), args); err != nil {
				return err
			}

			return nil
		},
	}

	flags := cmd.Flags()
	o.configFlags.AddFlags(flags)
	o.AddFlags(flags)

	return cmd
}

// DescribeOptions provides information required to update
// the current context on a user's KUBECONFIG.
type DescribeOptions struct {
	configFlags *genericclioptions.ConfigFlags
	printFlags  *genericclioptions.JSONYamlPrintFlags
	genericclioptions.IOStreams

	DescriberSettings *describe.DescriberSettings

	Builder   *resource.Builder
	Describer func(*meta.RESTMapping) (describe.ResourceDescriber, error)

	AllNamespaces bool
	Namespace     string
	Selector      string
	BuilderArgs   []string

	Preview       bool
	PreviewFormat string
	RawPreview    bool
}

// AddFlags adds a flag to the flag set.
func (o *DescribeOptions) AddFlags(flags *pflag.FlagSet) {
	// kubectl flags
	flags.BoolVarP(&o.AllNamespaces, "all-namespaces", "A", false,
		"If present, list the requested object(s) across all namespaces."+
			"Namespace in current context is ignored even if specified with --namespace.")
	flags.StringVarP(&o.Selector, "selector", "l", "",
		"Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2)")
	flags.BoolVar(&o.DescriberSettings.ShowEvents, "show-events", true,
		"If true, display events related to the described object.")

	// original flags
	flags.BoolVarP(&o.Preview, "preview", "P", false,
		"If true, display the object YAML|JSON by preview window for fuzzy finder selector.")
	flags.StringVar(&o.PreviewFormat, "preview-format", "yaml",
		"Preview window output format. One of json|yaml.")
	flags.BoolVar(&o.RawPreview, "raw-preview", false,
		"If true, display the unsimplified object in the preview window. (default is simplified)")
}

// NewDescribeOptions provides an instance of DescribeOptions with default values.
func NewDescribeOptions(streams genericclioptions.IOStreams) *DescribeOptions {
	return &DescribeOptions{
		configFlags: genericclioptions.NewConfigFlags(true),
		printFlags:  genericclioptions.NewJSONYamlPrintFlags(),
		IOStreams:   streams,
		DescriberSettings: &describe.DescriberSettings{
			ShowEvents: true,
		},
	}
}

// Complete sets all information required for show details.
func (o *DescribeOptions) Complete(cmd *cobra.Command, args []string) error {
	o.Builder = resource.NewBuilder(o.configFlags)

	o.Describer = func(mapping *meta.RESTMapping) (describe.ResourceDescriber, error) {
		return describe.DescriberFn(o.configFlags, mapping)
	}

	o.BuilderArgs = args

	if !o.AllNamespaces {
		kubeConfig := o.configFlags.ToRawKubeConfigLoader()

		namespace, _, err := kubeConfig.Namespace()
		if err != nil {
			return fmt.Errorf("faild to get namespace from kube config: %w", err)
		}

		o.Namespace = namespace
	}

	return nil
}

// Validate ensures that all required arguments and flag values are provided.
func (DescribeOptions) Validate() error {
	return nil
}

// Run execute fizzy finder and show details.
func (o *DescribeOptions) Run(ctx context.Context, args []string) error {
	r := o.Builder.
		Unstructured().
		ContinueOnError().
		NamespaceParam(o.Namespace).DefaultNamespace().AllNamespaces(o.AllNamespaces).
		LabelSelectorParam(o.Selector).
		ResourceTypeOrNameArgs(true, o.BuilderArgs...).
		Flatten().
		Do()

	if err := r.Err(); err != nil {
		return fmt.Errorf("failed to request: %w", err)
	}

	infos, err := r.Infos()
	if err != nil {
		return fmt.Errorf("failed to get infos: %w", err)
	}

	var printer printers.ResourcePrinter
	if o.Preview {
		printer, err = o.printFlags.ToPrinter(o.PreviewFormat)
		if err != nil {
			return fmt.Errorf("failed to get printer: %w", err)
		}
	}

	info, err := fuzzyfinder.Infos(infos, printer, o.AllNamespaces, o.RawPreview)
	if err != nil {
		return fmt.Errorf("failed to fuzzyfinder execute: %w", err)
	}

	mapping := info.ResourceMapping()

	describer, err := o.Describer(mapping)
	if err != nil {
		return fmt.Errorf("failed to get describer: %w", err)
	}

	s, err := describer.Describe(info.Namespace, info.Name, *o.DescriberSettings)
	if err != nil {
		return fmt.Errorf("failed to generates output: %w", err)
	}

	fmt.Fprintf(o.Out, "%s", s)

	return nil
}
