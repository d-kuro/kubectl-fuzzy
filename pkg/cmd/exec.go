package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/d-kuro/kubectl-fuzzy/pkg/fuzzyfinder"
	"github.com/d-kuro/kubectl-fuzzy/pkg/kubernetes"
	dockerterm "github.com/moby/term"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/cli-runtime/pkg/resource"
	coreclient "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/kubectl/pkg/scheme"
	"k8s.io/kubectl/pkg/util/interrupt"
	"k8s.io/kubectl/pkg/util/term"
)

const (
	exampleExec = `
	# Selecting a Pod with the fuzzy finder and execute a command in a container
	kubectl fuzzy exec [flags] -- COMMAND [args...]
`
)

// NewCmdExec provides a cobra command wrapping ExecOptions.
func NewCmdExec(config *genericclioptions.ConfigFlags, streams genericclioptions.IOStreams) *cobra.Command {
	o := NewExecOptions(config, streams)

	cmd := &cobra.Command{
		Use:           "exec",
		Short:         "Selecting a Pod with the fuzzy finder and execute a command in a container",
		Example:       exampleExec,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(c *cobra.Command, args []string) error {
			argsLenAtDash := c.ArgsLenAtDash()

			if err := o.Complete(c, args, argsLenAtDash); err != nil {
				return err
			}

			if err := o.Validate(); err != nil {
				return err
			}

			if err := o.Run(c.Context()); err != nil {
				return err
			}

			return nil
		},
	}

	o.AddFlags(cmd.Flags())

	return cmd
}

// ExecOptions provides information required to update
// the current context on a user's KUBECONFIG.
type ExecOptions struct {
	configFlags *genericclioptions.ConfigFlags
	printFlags  *genericclioptions.JSONYamlPrintFlags
	streamOptions

	client  coreclient.CoreV1Interface
	builder *resource.Builder

	allNamespaces bool
	namespace     string
	selector      string
	command       []string

	preview       bool
	previewFormat string
	rawPreview    bool
}

// streamOptions holds information pertaining to the streaming session.
type streamOptions struct {
	genericclioptions.IOStreams

	stdin bool
	tty   bool
	// interruptParent, if set, is used to handle interrupts while attached
	interruptParent *interrupt.Handler
}

// AddFlags adds a flag to the flag set.
func (o *ExecOptions) AddFlags(flags *pflag.FlagSet) {
	// kubectl flags
	flags.BoolVarP(&o.allNamespaces, "all-namespaces", "A", false,
		"If present, list the requested object(s) across all namespaces. "+
			"Namespace in current context is ignored even if specified with --namespace.")
	flags.StringVarP(&o.selector, "selector", "l", "",
		"Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2)")
	flags.BoolVarP(&o.stdin, "stdin", "i", false,
		"Pass stdin to the container")
	flags.BoolVarP(&o.tty, "tty", "t", false,
		"Stdin is a TTY")

	// original flags
	flags.BoolVarP(&o.preview, "preview", "P", false,
		"If true, display the object YAML|JSON by preview window for fuzzy finder selector.")
	flags.StringVar(&o.previewFormat, "preview-format", "yaml",
		"Preview window output format. One of json|yaml.")
	flags.BoolVar(&o.rawPreview, "raw-preview", false,
		"If true, display the unsimplified object in the preview window. (default is simplified)")
}

// NewExecOptions provides an instance of ExecOptions with default values.
func NewExecOptions(config *genericclioptions.ConfigFlags, streams genericclioptions.IOStreams) *ExecOptions {
	return &ExecOptions{
		streamOptions: streamOptions{
			IOStreams: streams,
		},
		configFlags: config,
		printFlags:  genericclioptions.NewJSONYamlPrintFlags(),
	}
}

// Complete sets all information required for execute a command in a container.
func (o *ExecOptions) Complete(cmd *cobra.Command, args []string, argsLenAtDash int) error {
	switch {
	case argsLenAtDash > -1:
		o.command = args[argsLenAtDash:]
	case len(args) > 0:
		fmt.Fprint(o.IOStreams.ErrOut,
			"kubectl exec fzf [COMMAND] is DEPRECATED and will be removed in a future version."+
				"Use kubectl exec fzf -- [COMMAND] instead.\n")

		o.command = args
	}

	client, err := kubernetes.NewClient(o.configFlags)
	if err != nil {
		return fmt.Errorf("failed to new Kubernetes client: %w", err)
	}

	o.client = client.CoreV1()
	o.builder = resource.NewBuilder(o.configFlags)

	if !o.preview {
		o.preview, _ = strconv.ParseBool(os.Getenv(previewEnabledEnvVar))
	}

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
func (o *ExecOptions) Validate() error {
	if len(o.command) == 0 {
		return fmt.Errorf("you must specify at least one command for the container")
	}

	return nil
}

// Run execute fizzy finder and execute a command in a container.
func (o *ExecOptions) Run(ctx context.Context) error {
	r := o.builder.
		Unstructured().
		ContinueOnError().
		NamespaceParam(o.namespace).DefaultNamespace().AllNamespaces(o.allNamespaces).
		LabelSelectorParam(o.selector).
		ResourceTypeOrNameArgs(true, "pods").
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
		return fmt.Errorf("resource not found")
	}

	var printer printers.ResourcePrinter
	if o.preview {
		printer, err = o.printFlags.ToPrinter(o.previewFormat)
		if err != nil {
			return err
		}
	}

	info, err := fuzzyfinder.Infos(infos,
		fuzzyfinder.WithAllNamespaces(o.allNamespaces),
		fuzzyfinder.WithPreview(printer),
		fuzzyfinder.WithRawPreview(o.rawPreview))
	if err != nil {
		return fmt.Errorf("failed to fuzzyfinder execute: %w", err)
	}

	uncastVersionedObj, err := scheme.Scheme.ConvertToVersion(info.Object, corev1.SchemeGroupVersion)
	if err != nil {
		return fmt.Errorf("from must be an existing cronjob: %v", err)
	}

	pod, ok := uncastVersionedObj.(*corev1.Pod)
	if !ok {
		return fmt.Errorf("illegal types that are not pod")
	}

	var containerName string

	if len(pod.Spec.Containers) > 1 {
		container, err := fuzzyfinder.Containers(pod.Spec.Containers)
		if err != nil {
			return fmt.Errorf("failed to fuzzyfinder execute: %w", err)
		}

		containerName = container.Name
	} else {
		containerName = pod.Spec.Containers[0].Name
	}

	// ensure we can recover the terminal while attached
	t := o.SetupTTY()

	var sizeQueue remotecommand.TerminalSizeQueue
	if t.Raw {
		// this call spawns a goroutine to monitor/update the terminal size
		sizeQueue = t.MonitorSize(t.GetSize())

		// unset p.Err if it was previously set because both stdout and stderr go over p.Out when tty is
		// true
		o.ErrOut = nil
	}

	if err := t.Safe(o.ExecFunc(pod, containerName, t, sizeQueue)); err != nil {
		return err
	}

	return nil
}

// ExecFunc returns a function for executing the execute a command in a container.
func (o *ExecOptions) ExecFunc(pod *corev1.Pod, containerName string,
	tty term.TTY, sizeQueue remotecommand.TerminalSizeQueue) func() error {
	fn := func() error {
		req := o.client.RESTClient().
			Post().
			Resource("pods").
			Name(pod.Name).
			Namespace(pod.Namespace).
			SubResource("exec").
			VersionedParams(&corev1.PodExecOptions{
				Container: containerName,
				Command:   o.command,
				Stdin:     o.stdin,
				Stdout:    o.Out != nil,
				Stderr:    o.ErrOut != nil,
				TTY:       tty.Raw,
			}, scheme.ParameterCodec)

		restConfig, err := o.configFlags.ToRESTConfig()
		if err != nil {
			return fmt.Errorf("faild to get REST client config: %w", err)
		}

		exec, err := remotecommand.NewSPDYExecutor(restConfig, http.MethodPost, req.URL())
		if err != nil {
			return err
		}

		return exec.Stream(remotecommand.StreamOptions{
			Stdin:             o.In,
			Stdout:            o.Out,
			Stderr:            o.ErrOut,
			Tty:               tty.Raw,
			TerminalSizeQueue: sizeQueue,
		})
	}

	return fn
}

func (o *streamOptions) SetupTTY() term.TTY {
	t := term.TTY{
		Parent: o.interruptParent,
		Out:    o.Out,
	}

	if !o.stdin {
		// need to nil out o.In to make sure we don't create a stream for stdin
		o.In = nil
		o.tty = false

		return t
	}

	t.In = o.In

	if !o.tty {
		return t
	}

	if !t.IsTerminalIn() {
		o.tty = false

		if o.ErrOut != nil {
			fmt.Fprintln(o.ErrOut, "Unable to use a TTY - input is not a terminal or the right kind of file")
		}

		return t
	}

	// if we get to here, the user wants to attach stdin, wants a TTY, and o.In is a terminal, so we
	// can safely set t.Raw to true
	t.Raw = true

	stdin, stdout, _ := dockerterm.StdStreams()
	o.In = stdin
	t.In = stdin

	if o.Out != nil {
		o.Out = stdout
		t.Out = stdout
	}

	return t
}
