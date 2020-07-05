package cmd

import (
	"context"
	"fmt"

	"github.com/d-kuro/kubectl-fuzzy/pkg/fuzzyfinder"
	"github.com/d-kuro/kubectl-fuzzy/pkg/kubernetes"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	batchv1client "k8s.io/client-go/kubernetes/typed/batch/v1"
	batchv1beta1client "k8s.io/client-go/kubernetes/typed/batch/v1beta1"
	"k8s.io/kubectl/pkg/scheme"
)

// NewCmdCreateJobFrom provides a cobra command wrapping CreateJobFromOptions.
func NewCmdCreateJobFrom(streams genericclioptions.IOStreams) *cobra.Command {
	o := NewCreateJobFromOptions(streams)

	cmd := &cobra.Command{
		Use:           "create-job-from",
		Short:         "Selecting a Pod with the fuzzy finder and view the log",
		Example:       "TODO",
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

// CreateJobFromOptions provides information required to update
// the current context on a user's KUBECONFIG.
type CreateJobFromOptions struct {
	configFlags *genericclioptions.ConfigFlags
	printFlags  *genericclioptions.PrintFlags
	genericclioptions.IOStreams

	printObj func(obj runtime.Object) error

	resource string
	name     string

	cronJobClient batchv1beta1client.CronJobsGetter
	jobClient     batchv1client.JobsGetter
	namespace     string
}

// NewCreateJobFromOptions provides an instance of CreateJobFromOptions with default values.
func NewCreateJobFromOptions(streams genericclioptions.IOStreams) *CreateJobFromOptions {
	return &CreateJobFromOptions{
		configFlags: genericclioptions.NewConfigFlags(true),
		printFlags:  genericclioptions.NewPrintFlags("created").WithTypeSetter(scheme.Scheme),
		IOStreams:   streams,
	}
}

// AddFlags adds a flag to the flag set.
func (o *CreateJobFromOptions) AddFlags(flags *pflag.FlagSet) {
}

// Complete sets all information required for get logs.
func (o *CreateJobFromOptions) Complete(cmd *cobra.Command, args []string) error {

	client, err := kubernetes.NewClient(o.configFlags)
	if err != nil {
		return fmt.Errorf("failed to new Kubernetes client: %w", err)
	}

	if len(args) == 0 {
		return fmt.Errorf("resource must specify and only supported cronjob")
	}

	o.resource = args[0]
	if len(args) >= 2 {
		o.name = args[1]
	}

	o.cronJobClient = client.BatchV1beta1()
	o.jobClient = client.BatchV1()

	kubeConfig := o.configFlags.ToRawKubeConfigLoader()

	namespace, _, err := kubeConfig.Namespace()
	if err != nil {
		return fmt.Errorf("faild to get namespace from kube config: %w", err)
	}

	o.namespace = namespace

	printer, err := o.printFlags.ToPrinter()
	if err != nil {
		return err
	}
	o.printObj = func(obj runtime.Object) error {
		return printer.PrintObj(obj, o.Out)
	}

	return nil
}

// Validate ensures that all required arguments and flag values are provided.
func (o *CreateJobFromOptions) Validate() error {
	if o.resource != "cronjob" && o.resource != "cj" {
		return fmt.Errorf("resource must be cronjob and only supported cronjob")
	}
	return nil
}

// Run execute fizzy finder and create job from cronJob.
func (o *CreateJobFromOptions) Run(ctx context.Context) error {

	cronJobs, err := o.cronJobClient.CronJobs(o.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list cronJobs: %w", err)
	}

	cronJob, err := fuzzyfinder.CronJobs(cronJobs.Items)
	if err != nil {
		return fmt.Errorf("failed to fuzzyfinder execute: %w", err)
	}

	job := o.createJobFromCronJob(&cronJob, &o.name)

	createOptions := metav1.CreateOptions{}
	res, err := o.jobClient.Jobs(cronJob.Namespace).Create(context.Background(), job, createOptions)
	if err != nil {
		return fmt.Errorf("failed to create job: %v", err)
	}

	return o.printObj(res)
}

func (o *CreateJobFromOptions) createJobFromCronJob(cronJob *batchv1beta1.CronJob, name *string) *batchv1.Job {
	annotations := make(map[string]string)
	annotations["cronjob.kubernetes.io/instantiate"] = "manual"
	for k, v := range cronJob.Spec.JobTemplate.Annotations {
		annotations[k] = v
	}

	job := &batchv1.Job{
		// this is ok because we know exactly how we want to be serialized
		TypeMeta: metav1.TypeMeta{APIVersion: batchv1.SchemeGroupVersion.String(), Kind: "Job"},
		ObjectMeta: metav1.ObjectMeta{
			Annotations:  annotations,
			Labels:       cronJob.Spec.JobTemplate.Labels,
			GenerateName: fmt.Sprintf("%s-", cronJob.Name),
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: batchv1beta1.SchemeGroupVersion.String(),
					Kind:       "CronJob",
					Name:       cronJob.GetName(),
					UID:        cronJob.GetUID(),
				},
			},
		},
		Spec: cronJob.Spec.JobTemplate.Spec,
	}
	if name != nil {
		job.ObjectMeta.Name = *name
	}
	return job
}
