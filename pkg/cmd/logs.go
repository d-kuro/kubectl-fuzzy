package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/d-kuro/kubectl-fuzzy/pkg/fuzzyfinder"
	"github.com/d-kuro/kubectl-fuzzy/pkg/kubernetes"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/cli-runtime/pkg/resource"
	coreclient "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/kubectl/pkg/scheme"
)

const (
	exampleLogs = `
	# Selecting a Pod with the fuzzy finder and view the log
	kubectl fuzzy logs [flags]
`
)

// NewCmdLogs provides a cobra command wrapping LogsOptions.
func NewCmdLogs(config *genericclioptions.ConfigFlags, streams genericclioptions.IOStreams) *cobra.Command {
	o := NewLogsOptions(config, streams)

	cmd := &cobra.Command{
		Use:           "logs",
		Short:         "Selecting a Pod with the fuzzy finder and view the log",
		Example:       exampleLogs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := o.Complete(c, args); err != nil {
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

// LogsOptions provides information required to update
// the current context on a user's KUBECONFIG.
type LogsOptions struct {
	configFlags *genericclioptions.ConfigFlags
	printFlags  *genericclioptions.JSONYamlPrintFlags
	genericclioptions.IOStreams

	allNamespaces bool
	namespace     string
	follow        bool
	previous      bool
	since         time.Duration
	sinceTime     string
	timestamps    bool
	tailLines     int64
	limitBytes    int64

	podClient coreclient.PodsGetter
	builder   *resource.Builder

	preview       bool
	previewFormat string
	rawPreview    bool
}

// AddFlags adds a flag to the flag set.
func (o *LogsOptions) AddFlags(flags *pflag.FlagSet) {
	// kubectl flags
	flags.BoolVarP(&o.allNamespaces, "all-namespaces", "A", false,
		"If present, list the requested object(s) across all namespaces. "+
			"Namespace in current context is ignored even if specified with --namespace.")
	flags.BoolVarP(&o.follow, "follow", "f", false,
		"Specify if the logs should be streamed.")
	flags.BoolVarP(&o.previous, "previous", "p", false,
		"If true, print the logs for the previous instance of the container in a pod if it exists.")
	flags.DurationVar(&o.since, "since", time.Second*0,
		"Only return logs newer than a relative duration like 5s, 2m, or 3h. Defaults to all logs. "+
			"Only one of since-time / since may be used.")
	flags.StringVar(&o.sinceTime, "since-time", "",
		"Only return logs after a specific date (RFC3339). Defaults to all logs. "+
			"Only one of since-time / since may be used.")
	flags.BoolVar(&o.timestamps, "timestamps", false,
		"Include timestamps on each line in the log output.")
	flags.Int64Var(&o.tailLines, "tail", -1,
		"Lines of recent log file to display. Defaults to -1 with no selector, "+
			"showing all log lines otherwise 10, if a selector is provided.")
	flags.Int64Var(&o.limitBytes, "limit-bytes", 0,
		"Maximum bytes of logs to return. Defaults to no limit.")

	// original flags
	flags.BoolVarP(&o.preview, "preview", "P", false,
		"If true, display the object YAML|JSON by preview window for fuzzy finder selector.")
	flags.StringVar(&o.previewFormat, "preview-format", "yaml",
		"Preview window output format. One of json|yaml.")
	flags.BoolVar(&o.rawPreview, "raw-preview", false,
		"If true, display the unsimplified object in the preview window. (default is simplified)")
}

// NewLogsOptions provides an instance of LogsOptions with default values.
func NewLogsOptions(flags *genericclioptions.ConfigFlags, streams genericclioptions.IOStreams) *LogsOptions {
	return &LogsOptions{
		configFlags: flags,
		printFlags:  genericclioptions.NewJSONYamlPrintFlags(),
		IOStreams:   streams,
	}
}

// Complete sets all information required for get logs.
func (o *LogsOptions) Complete(cmd *cobra.Command, args []string) error {
	client, err := kubernetes.NewClient(o.configFlags)
	if err != nil {
		return fmt.Errorf("failed to new Kubernetes client: %w", err)
	}

	if !o.preview {
		o.preview = os.Getenv(previewEnabledEnvVar) == "true"
	}

	o.podClient = client.CoreV1()
	o.builder = resource.NewBuilder(o.configFlags)

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
func (LogsOptions) Validate() error {
	return nil
}

// Run execute fizzy finder and view logs.
func (o *LogsOptions) Run(ctx context.Context) error {
	r := o.builder.
		Unstructured().
		ContinueOnError().
		NamespaceParam(o.namespace).DefaultNamespace().AllNamespaces(o.allNamespaces).
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

	req := o.podClient.Pods(pod.Namespace).GetLogs(pod.Name, &corev1.PodLogOptions{
		Container:    containerName,
		Follow:       o.follow,
		Previous:     o.previous,
		SinceSeconds: o.ConvertSinceSeconds(),
		SinceTime:    o.ConvertSinceTime(),
		Timestamps:   o.timestamps,
		TailLines:    o.ConvertTailLines(),
		LimitBytes:   o.ConvertLimitBytes(),
	})

	reader, err := req.Stream(ctx)
	if err != nil {
		return err
	}
	defer reader.Close()

	if _, err = io.Copy(o.Out, reader); err != nil {
		return err
	}

	return nil
}

func (o *LogsOptions) ConvertSinceSeconds() *int64 {
	i := int64(o.since)
	if i == 0 {
		return nil
	}

	return &i
}

func (o *LogsOptions) ConvertSinceTime() *metav1.Time {
	if len(o.sinceTime) == 0 {
		return nil
	}

	t, err := time.Parse(o.sinceTime, time.RFC3339)
	if err != nil {
		return nil
	}

	return &metav1.Time{Time: t}
}

func (o *LogsOptions) ConvertTailLines() *int64 {
	if o.tailLines == -1 {
		return nil
	}

	return &o.tailLines
}

func (o *LogsOptions) ConvertLimitBytes() *int64 {
	if o.limitBytes == 0 {
		return nil
	}

	return &o.limitBytes
}
