package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/d-kuro/kubectl-fuzzy/pkg/fuzzyfinder"
	"github.com/d-kuro/kubectl-fuzzy/pkg/kubernetes"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/cli-runtime/pkg/resource"
	batchv1client "k8s.io/client-go/kubernetes/typed/batch/v1"
	"k8s.io/kubectl/pkg/scheme"
)

const (
	exampleCreateJob = `
	# Selecting a CronJob with the fuzzy finder and create job
	# Only supported cronjob
	# If a jobName is omitted, generated from cronJob name
	kubectl fuzzy create job [jobName] --from=cronjob [flags]
`
)

// NewCmdCreateJob provides a cobra command wrapping CreateJobOptions.
func NewCmdCreateJob(config *genericclioptions.ConfigFlags, streams genericclioptions.IOStreams) *cobra.Command {
	o := NewCreateJobOptions(config, streams)

	cmd := &cobra.Command{
		Use:           "job [NAME] --from=cronjob",
		Short:         "Selecting a CronJob with the fuzzy finder and create job",
		Example:       exampleCreateJob,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := o.Complete(c, args); err != nil {
				return err
			}

			if err := o.Validate(); err != nil {
				return err
			}

			return o.Run(c.Context())
		},
	}

	flags := cmd.Flags()
	o.configFlags.AddFlags(flags)
	o.AddFlags(flags)

	return cmd
}

// CreateJobOptions provides information required to update
// the current context on a user's KUBECONFIG.
type CreateJobOptions struct {
	configFlags       *genericclioptions.ConfigFlags
	previewPrintFlags *genericclioptions.JSONYamlPrintFlags
	printFlags        *genericclioptions.PrintFlags
	genericclioptions.IOStreams

	printObj func(obj runtime.Object) error

	name string
	from string

	builder   *resource.Builder
	jobClient batchv1client.JobsGetter
	namespace string

	preview       bool
	previewFormat string
	rawPreview    bool
}

// NewCreateJobOptions provides an instance of CreateJobOptions with default values.
func NewCreateJobOptions(config *genericclioptions.ConfigFlags, streams genericclioptions.IOStreams) *CreateJobOptions {
	return &CreateJobOptions{
		configFlags:       config,
		previewPrintFlags: genericclioptions.NewJSONYamlPrintFlags(),
		printFlags:        genericclioptions.NewPrintFlags("created").WithTypeSetter(scheme.Scheme),
		IOStreams:         streams,
	}
}

// AddFlags adds a flag to the flag set.
func (o *CreateJobOptions) AddFlags(flags *pflag.FlagSet) {
	flags.StringVar(&o.from, "from", o.from, "The name of the resource to create a Job from (only cronjob is supported).")

	// original flags
	flags.BoolVarP(&o.preview, "preview", "P", false,
		"If true, display the object YAML|JSON by preview window for fuzzy finder selector.")
	flags.StringVar(&o.previewFormat, "preview-format", "yaml",
		"Preview window output format. One of json|yaml.")
	flags.BoolVar(&o.rawPreview, "raw-preview", false,
		"If true, display the unsimplified object in the preview window. (default is simplified)")
}

// Complete sets all information required for get logs.
func (o *CreateJobOptions) Complete(cmd *cobra.Command, args []string) error {
	client, err := kubernetes.NewClient(o.configFlags)
	if err != nil {
		return fmt.Errorf("failed to new Kubernetes client: %w", err)
	}

	if o.from == "" {
		return fmt.Errorf("--from option is required, only supported job from cronjob")
	}

	if o.from != "cronjob" {
		return fmt.Errorf("must specify resource, only supported cronjob")
	}

	if !o.preview {
		o.preview, _ = strconv.ParseBool(os.Getenv(previewEnabledEnvVar))
	}

	if len(args) >= 1 {
		o.name = args[0]
	}

	o.jobClient = client.BatchV1()
	o.builder = resource.NewBuilder(o.configFlags)

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
func (o *CreateJobOptions) Validate() error {
	return nil
}

// Run execute fizzy finder and create job from cronJob.
func (o *CreateJobOptions) Run(ctx context.Context) error {
	infos, err := o.builder.
		Unstructured().
		NamespaceParam(o.namespace).DefaultNamespace().
		ResourceTypes(o.from).
		SelectAllParam(true).
		Flatten().
		Latest().
		Do().
		Infos()
	if err != nil {
		return fmt.Errorf("failed to list cronJobs: %w", err)
	}

	if len(infos) == 0 {
		return fmt.Errorf("resource not found")
	}

	var printer printers.ResourcePrinter
	if o.preview {
		printer, err = o.previewPrintFlags.ToPrinter(o.previewFormat)
		if err != nil {
			return err
		}
	}

	info, err := fuzzyfinder.Infos(infos,
		fuzzyfinder.WithAllNamespaces(false),
		fuzzyfinder.WithPreview(printer),
		fuzzyfinder.WithRawPreview(o.rawPreview))
	if err != nil {
		return fmt.Errorf("failed to fuzzyfinder execute: %w", err)
	}

	uncastVersionedObj, err := scheme.Scheme.ConvertToVersion(info.Object, batchv1beta1.SchemeGroupVersion)
	if err != nil {
		return fmt.Errorf("failed to convert resource into cronjob: %w", err)
	}

	cj, ok := uncastVersionedObj.(*batchv1beta1.CronJob)
	if !ok {
		return fmt.Errorf("failed to cast cronjob")
	}

	job := o.createJobFromCronJob(cj, &o.name)

	res, err := o.jobClient.Jobs(cj.Namespace).Create(context.Background(), job, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create job: %w", err)
	}

	return o.printObj(res)
}

func (o *CreateJobOptions) createJobFromCronJob(cronJob *batchv1beta1.CronJob, name *string) *batchv1.Job {
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
