package cmd

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	coreclient "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/d-kuro/kubectl-fuzzy/pkg/fuzzyfinder"
	"github.com/d-kuro/kubectl-fuzzy/pkg/kubernetes"
)

const (
	exampleLogs = `
	# Selecting a Pod with the fuzzy finder and view the log
	kubectl fuzzy logs [flags]
`
)

// NewCmdLogs provides a cobra command wrapping LogsOptions.
func NewCmdLogs(streams genericclioptions.IOStreams) *cobra.Command {
	o := NewLogsOptions(streams)

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

	flags := cmd.Flags()
	o.configFlags.AddFlags(flags)
	o.AddFlags(flags)

	return cmd
}

// LogsOptions provides information required to update
// the current context on a user's KUBECONFIG.
type LogsOptions struct {
	configFlags *genericclioptions.ConfigFlags
	genericclioptions.IOStreams

	AllNamespaces bool
	Follow        bool
	Previous      bool
	Since         time.Duration
	SinceTime     string
	Timestamps    bool
	TailLines     int64
	LimitBytes    int64

	PodClient coreclient.PodsGetter
	Namespace string
}

// AddFlags adds a flag to the flag set.
func (o *LogsOptions) AddFlags(flags *pflag.FlagSet) {
	flags.BoolVarP(&o.AllNamespaces, "all-namespaces", "A", false,
		"If present, list the requested object(s) across all namespaces."+
			"Namespace in current context is ignored even if specified with --namespace.")
	flags.BoolVarP(&o.Follow, "follow", "f", false,
		"Specify if the logs should be streamed.")
	flags.BoolVarP(&o.Previous, "previous", "p", false,
		"If true, print the logs for the previous instance of the container in a pod if it exists.")
	flags.DurationVar(&o.Since, "since", time.Second*0,
		"Only return logs newer than a relative duration like 5s, 2m, or 3h. Defaults to all logs."+
			"Only one of since-time / since may be used.")
	flags.StringVar(&o.SinceTime, "since-time", "",
		"Only return logs after a specific date (RFC3339). Defaults to all logs."+
			"Only one of since-time / since may be used.")
	flags.BoolVar(&o.Timestamps, "timestamps", false,
		"Include timestamps on each line in the log output.")
	flags.Int64Var(&o.TailLines, "tail", -1,
		"Lines of recent log file to display. Defaults to -1 with no selector,"+
			"showing all log lines otherwise 10, if a selector is provided.")
	flags.Int64Var(&o.LimitBytes, "limit-bytes", 0,
		"Maximum bytes of logs to return. Defaults to no limit.")
}

// NewLogsOptions provides an instance of LogsOptions with default values.
func NewLogsOptions(streams genericclioptions.IOStreams) *LogsOptions {
	return &LogsOptions{
		configFlags: genericclioptions.NewConfigFlags(true),
		IOStreams:   streams,
	}
}

// Complete sets all information required for get logs.
func (o *LogsOptions) Complete(cmd *cobra.Command, args []string) error {
	client, err := kubernetes.NewClient(o.configFlags)
	if err != nil {
		return fmt.Errorf("failed to new Kubernetes client: %w", err)
	}

	o.PodClient = client.CoreV1()

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
func (LogsOptions) Validate() error {
	return nil
}

// Run execute fizzy finder and view logs.
func (o *LogsOptions) Run(ctx context.Context) error {
	pods, err := o.PodClient.Pods(o.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list pods: %w", err)
	}

	pod, err := fuzzyfinder.Pods(pods.Items, o.AllNamespaces)
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

	req := o.PodClient.Pods(pod.Namespace).GetLogs(pod.Name, &corev1.PodLogOptions{
		Container:    containerName,
		Follow:       o.Follow,
		Previous:     o.Previous,
		SinceSeconds: o.ConvertSinceSeconds(),
		SinceTime:    o.ConvertSinceTime(),
		Timestamps:   o.Timestamps,
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
	i := int64(o.Since)
	if i == 0 {
		return nil
	}

	return &i
}

func (o *LogsOptions) ConvertSinceTime() *metav1.Time {
	if len(o.SinceTime) == 0 {
		return nil
	}

	t, err := time.Parse(o.SinceTime, time.RFC3339)
	if err != nil {
		return nil
	}

	return &metav1.Time{Time: t}
}

func (o *LogsOptions) ConvertTailLines() *int64 {
	if o.TailLines == -1 {
		return nil
	}

	return &o.TailLines
}

func (o *LogsOptions) ConvertLimitBytes() *int64 {
	if o.LimitBytes == 0 {
		return nil
	}

	return &o.LimitBytes
}
