package cmd

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	dockerterm "github.com/moby/term"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/printers"
	coreclient "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/kubectl/pkg/scheme"
	"k8s.io/kubectl/pkg/util/interrupt"
	"k8s.io/kubectl/pkg/util/term"

	"github.com/d-kuro/kubectl-fuzzy/pkg/fuzzyfinder"
	"github.com/d-kuro/kubectl-fuzzy/pkg/kubernetes"
)

const (
	exampleExec = `
	# Selecting a Pod with the fuzzy finder and execute a command in a container
	kubectl fuzzy exec [flags] -- COMMAND [args...]
`
)

// NewCmdExec provides a cobra command wrapping ExecOptions.
func NewCmdExec(streams genericclioptions.IOStreams) *cobra.Command {
	o := NewExecOptions(streams)

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

	flags := cmd.Flags()
	o.configFlags.AddFlags(flags)
	o.AddFlags(flags)

	return cmd
}

// ExecOptions provides information required to update
// the current context on a user's KUBECONFIG.
type ExecOptions struct {
	configFlags *genericclioptions.ConfigFlags
	printFlags  *genericclioptions.JSONYamlPrintFlags
	StreamOptions

	Client coreclient.CoreV1Interface

	AllNamespaces bool
	Command       []string
	Namespace     string

	Preview       bool
	PreviewFormat string
}

// StreamOptions holds information pertaining to the streaming session.
type StreamOptions struct {
	genericclioptions.IOStreams

	Stdin bool
	TTY   bool
	// InterruptParent, if set, is used to handle interrupts while attached
	InterruptParent *interrupt.Handler
}

// AddFlags adds a flag to the flag set.
func (o *ExecOptions) AddFlags(flags *pflag.FlagSet) {
	// kubectl flags
	flags.BoolVarP(&o.AllNamespaces, "all-namespaces", "A", false,
		"If present, list the requested object(s) across all namespaces."+
			"Namespace in current context is ignored even if specified with --namespace.")
	flags.BoolVarP(&o.Stdin, "stdin", "i", false,
		"Pass stdin to the container")
	flags.BoolVarP(&o.TTY, "tty", "t", false,
		"Stdin is a TTY")

	// original flags
	flags.BoolVar(&o.Preview, "preview", false,
		"If true, display the object YAML|JSON by preview window for fuzzy finder selector.")
	flags.StringVar(&o.PreviewFormat, "preview-format", "yaml",
		"Preview window output format. One of json|yaml.")
}

// NewExecOptions provides an instance of ExecOptions with default values.
func NewExecOptions(streams genericclioptions.IOStreams) *ExecOptions {
	return &ExecOptions{
		StreamOptions: StreamOptions{
			IOStreams: streams,
		},
		configFlags: genericclioptions.NewConfigFlags(true),
		printFlags:  genericclioptions.NewJSONYamlPrintFlags(),
	}
}

// Complete sets all information required for execute a command in a container.
func (o *ExecOptions) Complete(cmd *cobra.Command, args []string, argsLenAtDash int) error {
	switch {
	case argsLenAtDash > -1:
		o.Command = args[argsLenAtDash:]
	case len(args) > 0:
		fmt.Fprint(o.IOStreams.ErrOut,
			"kubectl exec fzf [COMMAND] is DEPRECATED and will be removed in a future version."+
				"Use kubectl exec fzf -- [COMMAND] instead.\n")

		o.Command = args
	}

	client, err := kubernetes.NewClient(o.configFlags)
	if err != nil {
		return fmt.Errorf("failed to new Kubernetes client: %w", err)
	}

	o.Client = client.CoreV1()

	if !o.AllNamespaces {
		kubeConfig := o.configFlags.ToRawKubeConfigLoader()

		namespace, _, err := kubeConfig.Namespace()
		if err != nil {
			return fmt.Errorf("faild to get namespace from kube config: %w", err)
		}

		o.Namespace = namespace
	}

	o.PreviewFormat = strings.ToLower(o.PreviewFormat)

	return nil
}

// Validate ensures that all required arguments and flag values are provided.
func (o *ExecOptions) Validate() error {
	if len(o.Command) == 0 {
		return fmt.Errorf("you must specify at least one command for the container")
	}

	return nil
}

// Run execute fizzy finder and execute a command in a container.
func (o *ExecOptions) Run(ctx context.Context) error {
	pods, err := o.Client.Pods(o.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list pods: %w", err)
	}

	var printer printers.ResourcePrinter
	if o.Preview {
		printer, err = o.printFlags.ToPrinter(o.PreviewFormat)
		if err != nil {
			return err
		}
	}

	pod, err := fuzzyfinder.Pods(pods.Items, o.AllNamespaces, printer)
	if err != nil {
		return fmt.Errorf("failed to fuzzyfinder execute: %w", err)
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
func (o *ExecOptions) ExecFunc(pod corev1.Pod, containerName string,
	tty term.TTY, sizeQueue remotecommand.TerminalSizeQueue) func() error {
	fn := func() error {
		req := o.Client.RESTClient().
			Post().
			Resource("pods").
			Name(pod.Name).
			Namespace(pod.Namespace).
			SubResource("exec").
			VersionedParams(&corev1.PodExecOptions{
				Container: containerName,
				Command:   o.Command,
				Stdin:     o.Stdin,
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

func (o *StreamOptions) SetupTTY() term.TTY {
	t := term.TTY{
		Parent: o.InterruptParent,
		Out:    o.Out,
	}

	if !o.Stdin {
		// need to nil out o.In to make sure we don't create a stream for stdin
		o.In = nil
		o.TTY = false

		return t
	}

	t.In = o.In

	if !o.TTY {
		return t
	}

	if !t.IsTerminalIn() {
		o.TTY = false

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
