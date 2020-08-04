package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/d-kuro/kubectl-fuzzy/pkg/fuzzyfinder"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/kubectl/pkg/describe"
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

	describerSettings *describe.DescriberSettings

	builder   *resource.Builder
	describer func(*meta.RESTMapping) (describe.ResourceDescriber, error)

	allNamespaces bool
	namespace     string
	selector      string
	builderArgs   []string

	preview       bool
	previewFormat string
	rawPreview    bool
}

// AddFlags adds a flag to the flag set.
func (o *DescribeOptions) AddFlags(flags *pflag.FlagSet) {
	// kubectl flags
	flags.BoolVarP(&o.allNamespaces, "all-namespaces", "A", false,
		"If present, list the requested object(s) across all namespaces."+
			"Namespace in current context is ignored even if specified with --namespace.")
	flags.StringVarP(&o.selector, "selector", "l", "",
		"Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2)")
	flags.BoolVar(&o.describerSettings.ShowEvents, "show-events", true,
		"If true, display events related to the described object.")

	// original flags
	flags.BoolVarP(&o.preview, "preview", "P", false,
		"If true, display the object YAML|JSON by preview window for fuzzy finder selector.")
	flags.StringVar(&o.previewFormat, "preview-format", "yaml",
		"Preview window output format. One of json|yaml.")
	flags.BoolVar(&o.rawPreview, "raw-preview", false,
		"If true, display the unsimplified object in the preview window. (default is simplified)")
}

// NewDescribeOptions provides an instance of DescribeOptions with default values.
func NewDescribeOptions(streams genericclioptions.IOStreams) *DescribeOptions {
	return &DescribeOptions{
		configFlags: genericclioptions.NewConfigFlags(true),
		printFlags:  genericclioptions.NewJSONYamlPrintFlags(),
		IOStreams:   streams,
		describerSettings: &describe.DescriberSettings{
			ShowEvents: true,
		},
	}
}

// Complete sets all information required for show details.
func (o *DescribeOptions) Complete(cmd *cobra.Command, args []string) error {
	o.builder = resource.NewBuilder(o.configFlags)

	o.describer = func(mapping *meta.RESTMapping) (describe.ResourceDescriber, error) {
		return describe.DescriberFn(o.configFlags, mapping)
	}

	o.builderArgs = args

	if !o.allNamespaces {
		kubeConfig := o.configFlags.ToRawKubeConfigLoader()

		namespace, _, err := kubeConfig.Namespace()
		if err != nil {
			return fmt.Errorf("faild to get namespace from kube config: %w", err)
		}

		o.namespace = namespace
	}

	return nil
}

// Validate ensures that all required arguments and flag values are provided.
func (DescribeOptions) Validate() error {
	return nil
}

// Run execute fizzy finder and show details.
func (o *DescribeOptions) Run(ctx context.Context, args []string) error {
	r := o.builder.
		Unstructured().
		ContinueOnError().
		NamespaceParam(o.namespace).DefaultNamespace().AllNamespaces(o.allNamespaces).
		LabelSelectorParam(o.selector).
		ResourceTypeOrNameArgs(true, o.builderArgs...).
		Flatten().
		Do()

	if err := r.Err(); err != nil {
		return fmt.Errorf("failed to request: %w", err)
	}

	infos, err := r.Infos()
	if err != nil {
		return fmt.Errorf("failed to get infos: %w", err)
	}

	if len(infos) == 0 {
		return errors.New("no resources found")
	}

	var printer printers.ResourcePrinter
	if o.preview {
		printer, err = o.printFlags.ToPrinter(o.previewFormat)
		if err != nil {
			return fmt.Errorf("failed to get printer: %w", err)
		}
	}

	info, err := fuzzyfinder.Infos(infos,
		fuzzyfinder.WithAllNamespaces(o.allNamespaces),
		fuzzyfinder.WithPreview(printer),
		fuzzyfinder.WithRawPreview(o.rawPreview))
	if err != nil {
		return fmt.Errorf("failed to fuzzyfinder execute: %w", err)
	}

	mapping := info.ResourceMapping()

	describer, err := o.describer(mapping)
	if err != nil {
		return fmt.Errorf("failed to get describer: %w", err)
	}

	s, err := describer.Describe(info.Namespace, info.Name, *o.describerSettings)
	if err != nil {
		return fmt.Errorf("failed to generates output: %w", err)
	}

	fmt.Fprintf(o.Out, "%s", s)

	return nil
}
