package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/d-kuro/kubectl-fuzzy/pkg/fuzzyfinder"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog/v2"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	cmdwait "k8s.io/kubectl/pkg/cmd/wait"
)

const (
	exampleDelete = `
	# Selecting an object with the fuzzy finder and delete
	kubectl fuzzy delete TYPE [flags]
`
)

// NewCmdDelete provides a cobra command wrapping DeleteOptions.
func NewCmdDelete(config *genericclioptions.ConfigFlags, streams genericclioptions.IOStreams) *cobra.Command {
	o := NewDeleteOptions(config, streams)

	cmd := &cobra.Command{
		Use:           "delete",
		Short:         "Selecting an object with the fuzzy finder and delete",
		Example:       exampleDelete,
		SilenceUsage:  true,
		SilenceErrors: true,
		SuggestFor:    []string{"rm"},
		RunE: func(c *cobra.Command, args []string) error {
			if err := o.Complete(c, args); err != nil {
				return err
			}

			if err := o.Validate(); err != nil {
				return err
			}

			return o.Run(c.Context(), args)
		},
	}

	o.AddFlags(cmd.Flags())
	cmdutil.AddDryRunFlag(cmd)

	return cmd
}

// DeleteOptions provides information required to update
// the current context on a user's KUBECONFIG.
type DeleteOptions struct {
	configFlags *genericclioptions.ConfigFlags
	printFlags  *genericclioptions.JSONYamlPrintFlags
	genericclioptions.IOStreams

	allNamespaces    bool
	labelSelector    string
	fieldSelector    string
	cascade          bool
	deleteNow        bool
	forceDeletion    bool
	waitForDeletion  bool
	warnClusterScope bool

	gracePeriod int
	timeout     time.Duration

	dryRunStrategy cmdutil.DryRunStrategy
	dryRunVerifier *resource.QueryParamVerifier

	output string

	dynamicClient dynamic.Interface
	namespace     string

	preview       bool
	previewFormat string
	rawPreview    bool
}

// AddFlags adds a flag to the flag set.
func (o *DeleteOptions) AddFlags(flags *pflag.FlagSet) {
	// kubectl flags
	flags.BoolVarP(&o.allNamespaces, "all-namespaces", "A", false,
		"If present, list the requested object(s) across all namespaces. "+
			"Namespace in current context is ignored even if specified with --namespace.")
	flags.BoolVar(&o.cascade, "cascade", true,
		"If true, cascade the deletion of the resources managed by this resource "+
			"(e.g. Pods created by a ReplicationController). Default true.")
	flags.StringVar(&o.fieldSelector, "field-selector", "",
		"Selector (field query) to filter on, supports '=', '==', and '!='."+
			"(e.g. --field-selector key1=value1,key2=value2)."+
			"The server only supports a limited number of field queries per type.")
	flags.BoolVar(&o.forceDeletion, "force", false,
		"If true, immediately remove resources from API and bypass graceful deletion. "+
			"Note that immediate deletion of some resources may result in inconsistency or data loss and requires confirmation.")
	flags.IntVar(&o.gracePeriod, "grace-period", -1,
		"Period of time in seconds given to the resource to terminate gracefully. Ignored if negative. "+
			"Set to 1 for immediate shutdown. Can only be set to 0 when --force is true (force deletion).")
	flags.BoolVar(&o.deleteNow, "now", false,
		"If true, resources are signaled for immediate shutdown (same as --grace-period=1).")
	flags.StringVarP(&o.output, "output", "o", "",
		"Output mode. Use \"-o name\" for shorter output (resource/name).")
	flags.StringVarP(&o.labelSelector, "selector", "l", "",
		"Selector (label query) to filter on, not including uninitialized ones.")
	flags.DurationVar(&o.timeout, "timeout", 0*time.Second,
		"The length of time to wait before giving up on a delete, zero means determine a timeout from the size of the object")
	flags.BoolVar(&o.waitForDeletion, "wait", true,
		"If true, wait for resources to be gone before returning. This waits for finalizers.")

	// original flags
	flags.BoolVarP(&o.preview, "preview", "P", false,
		"If true, display the object YAML|JSON by preview window for fuzzy finder selector.")
	flags.StringVar(&o.previewFormat, "preview-format", "yaml",
		"Preview window output format. One of json|yaml.")
	flags.BoolVar(&o.rawPreview, "raw-preview", false,
		"If true, display the unsimplified object in the preview window. (default is simplified)")
}

// NewDeleteOptions provides an instance of DeleteOptions with default values.
func NewDeleteOptions(config *genericclioptions.ConfigFlags, streams genericclioptions.IOStreams) *DeleteOptions {
	return &DeleteOptions{
		configFlags: config,
		printFlags:  genericclioptions.NewJSONYamlPrintFlags(),
		IOStreams:   streams,
	}
}

// Complete sets all information required for show details.
func (o *DeleteOptions) Complete(cmd *cobra.Command, args []string) error {
	cmdNamespace, enforceNamespace, err := o.configFlags.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return fmt.Errorf("faild to get namespace from kube config: %w", err)
	}

	if !o.preview {
		o.preview, _ = strconv.ParseBool(os.Getenv(previewEnabledEnvVar))
	}

	o.warnClusterScope = enforceNamespace && !o.allNamespaces

	if o.deleteNow {
		if o.gracePeriod != -1 {
			return fmt.Errorf("--now and --grace-period cannot be specified together")
		}

		o.gracePeriod = 1
	}

	if o.gracePeriod == 0 && !o.forceDeletion {
		// To preserve backwards compatibility, but prevent accidental data loss, we convert --grace-period=0
		// into --grace-period=1. Users may provide --force to bypass this conversion.
		o.gracePeriod = 1
	}

	if o.forceDeletion && o.gracePeriod < 0 {
		o.gracePeriod = 0
	}

	o.dryRunStrategy, err = cmdutil.GetDryRunStrategy(cmd)
	if err != nil {
		return fmt.Errorf("faild to get dry-run strategy: %w", err)
	}

	restConfig, err := o.configFlags.ToRESTConfig()
	if err != nil {
		return fmt.Errorf("faild to get REST config: %w", err)
	}

	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("faild to create dynamic client: %w", err)
	}

	discoveryClient, err := o.configFlags.ToDiscoveryClient()
	if err != nil {
		return fmt.Errorf("failed to create discovery client: %w", err)
	}

	o.dryRunVerifier = resource.NewQueryParamVerifier(dynamicClient, discoveryClient, resource.QueryParamDryRun)
	o.dynamicClient = dynamicClient
	o.namespace = cmdNamespace

	return nil
}

// Validate ensures that all required arguments and flag values are provided.
func (o *DeleteOptions) Validate() error {
	switch {
	case o.gracePeriod == 0 && o.forceDeletion:
		_, _ = fmt.Fprintln(o.ErrOut,
			"warning: Immediate deletion does not wait for confirmation that the running resource has been terminated. "+
				"The resource may continue to run on the cluster indefinitely.")
	case o.gracePeriod > 0 && o.forceDeletion:
		return fmt.Errorf("--force and --grace-period greater than 0 cannot be specified together")
	}

	return nil
}

// Run execute fizzy finder and delete object.
func (o *DeleteOptions) Run(ctx context.Context, args []string) error {
	r := resource.NewBuilder(o.configFlags).
		Unstructured().
		ContinueOnError().
		NamespaceParam(o.namespace).DefaultNamespace().
		LabelSelectorParam(o.labelSelector).
		FieldSelectorParam(o.fieldSelector).
		AllNamespaces(o.allNamespaces).
		ResourceTypeOrNameArgs(true, args...).RequireObject(false).
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

	uidMap := cmdwait.UIDMap{}

	options := &metav1.DeleteOptions{}
	if o.gracePeriod >= 0 {
		options = metav1.NewDeleteOptions(int64(o.gracePeriod))
	}

	policy := metav1.DeletePropagationBackground
	if !o.cascade {
		policy = metav1.DeletePropagationOrphan
	}

	options.PropagationPolicy = &policy

	if o.warnClusterScope && info.Mapping.Scope.Name() == meta.RESTScopeNameRoot {
		_, _ = fmt.Fprintf(o.ErrOut, "warning: deleting cluster-scoped resources, not scoped to the provided namespace\n")
		o.warnClusterScope = false
	}

	if o.dryRunStrategy == cmdutil.DryRunClient {
		o.PrintObj(info)

		return nil
	}

	if o.dryRunStrategy == cmdutil.DryRunServer {
		if err := o.dryRunVerifier.HasSupport(info.Mapping.GroupVersionKind); err != nil {
			return err
		}
	}

	response, err := o.deleteResource(info, options)
	if err != nil {
		return err
	}

	resourceLocation := cmdwait.ResourceLocation{
		GroupResource: info.Mapping.Resource.GroupResource(),
		Namespace:     info.Namespace,
		Name:          info.Name,
	}

	if status, ok := response.(*metav1.Status); ok && status.Details != nil {
		uidMap[resourceLocation] = status.Details.UID

		return nil
	}

	responseMetadata, err := meta.Accessor(response)
	if err != nil {
		// we don't have UID, but we didn't fail the delete, next best thing is just skipping the UID
		klog.V(1).Info(err)

		return nil
	}

	uidMap[resourceLocation] = responseMetadata.GetUID()

	if !o.waitForDeletion {
		return nil
	}

	// if we don't have a dynamic client, we don't want to wait.  Eventually when delete is cleaned up, this will likely
	// drop out.
	if o.dynamicClient == nil {
		return nil
	}

	// If we are dry-running, then we don't want to wait
	if o.dryRunStrategy != cmdutil.DryRunNone {
		return nil
	}

	const defaultEffectiveTimeout = 168 * time.Hour

	effectiveTimeout := o.timeout
	if effectiveTimeout == 0 {
		// if we requested to wait forever, set it to a week.
		effectiveTimeout = defaultEffectiveTimeout
	}

	waitOptions := cmdwait.WaitOptions{
		ResourceFinder: genericclioptions.ResourceFinderForResult(
			resource.InfoListVisitor([]*resource.Info{info})),
		UIDMap:        uidMap,
		DynamicClient: o.dynamicClient,
		Timeout:       effectiveTimeout,

		Printer:     printers.NewDiscardingPrinter(),
		ConditionFn: cmdwait.IsDeleted,
		IOStreams:   o.IOStreams,
	}

	err = waitOptions.RunWait()
	if errors.IsForbidden(err) || errors.IsMethodNotSupported(err) {
		// if we're forbidden from waiting, we shouldn't fail.
		// if the resource doesn't support a verb we need, we shouldn't fail.
		klog.V(1).Info(err)

		return nil
	}

	return err
}

func (o *DeleteOptions) deleteResource(info *resource.Info, options *metav1.DeleteOptions) (runtime.Object, error) {
	deleteResponse, err := resource.
		NewHelper(info.Client, info.Mapping).
		DryRun(o.dryRunStrategy == cmdutil.DryRunServer).
		DeleteWithOptions(info.Namespace, info.Name, options)
	if err != nil {
		return nil, cmdutil.AddSourceToErr("deleting", info.Source, err)
	}

	o.PrintObj(info)

	return deleteResponse, nil
}

// PrintObj for deleted objects is special because we do not have an object to print.
// This mirrors name printer behavior.
func (o *DeleteOptions) PrintObj(info *resource.Info) {
	operation := "deleted"
	groupKind := info.Mapping.GroupVersionKind
	kindString := fmt.Sprintf("%s.%s", strings.ToLower(groupKind.Kind), groupKind.Group)

	if len(groupKind.Group) == 0 {
		kindString = strings.ToLower(groupKind.Kind)
	}

	if o.gracePeriod == 0 {
		operation = "force deleted"
	}

	switch o.dryRunStrategy {
	case cmdutil.DryRunClient:
		operation = fmt.Sprintf("%s (dry run)", operation)
	case cmdutil.DryRunServer:
		operation = fmt.Sprintf("%s (server dry run)", operation)
	case cmdutil.DryRunNone:
		break
	}

	if o.output == "name" {
		// -o name: prints resource/name
		_ = fmt.Fprintf(o.Out, "%s/%s\n", kindString, info.Name)

		return
	}

	// understandable output by default
	_ = fmt.Fprintf(o.Out, "%s \"%s\" %s\n", kindString, info.Name, operation)
}
