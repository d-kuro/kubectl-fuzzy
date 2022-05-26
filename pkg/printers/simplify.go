package printers

import (
	"io"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/printers"
)

// Simplify wraps an existing printer and omits the metadata and status fields from the object before printing it.
// Implements the printers.ResourcePrinter interface.
type Simplify struct {
	Delegate printers.ResourcePrinter
}

var _ printers.ResourcePrinter = (*Simplify)(nil)

// omitMetadata omits the metadata from Object.
// The metadata to omit is as follows:
//
// * .metadata.generateName
// * .metadata.uid
// * .metadata.resourceVersion
// * .metadata.generation
// * .metadata.selfLink
// * .metadata.creationTimestamp
// * .metadata.deletionTimestamp
// * .metadata.deletionGracePeriodSeconds
// * .metadata.finalizers
// * .metadata.ownerReferences
// * .metadata.clusterName
// * .metadata.managedFields
// * .metadata.annotations.kubectl\.kubernetes\.io/last-applied-configuration.
func omitMetadata(o runtime.Object) runtime.Object {
	a, err := meta.Accessor(o)
	if err != nil {
		// The object is not a `metav1.Object`, ignore it.
		return o
	}

	a.SetGenerateName("")
	a.SetUID("")
	a.SetResourceVersion("")
	a.SetGeneration(0)
	a.SetSelfLink("")
	a.SetCreationTimestamp(metav1.Time{})
	a.SetDeletionTimestamp(nil)
	a.SetDeletionGracePeriodSeconds(nil)
	a.SetFinalizers(nil)
	a.SetOwnerReferences(nil)
	a.SetManagedFields(nil)

	omitLastAppliedConfigurationAnnotation(a)

	return o
}

// editLastAppliedConfigurationAnnotation omits "kubectl.kubernetes.io/last-applied-configuration" from the annotation.
func omitLastAppliedConfigurationAnnotation(o metav1.Object) {
	annotations := o.GetAnnotations()

	if _, ok := annotations[corev1.LastAppliedConfigAnnotation]; ok {
		delete(annotations, corev1.LastAppliedConfigAnnotation)

		o.SetAnnotations(annotations)
	}
}

// omitStatus omits the status from the Object.
func omitStatus(o runtime.Object) runtime.Object {
	unstructured, ok := o.(*unstructured.Unstructured)
	if !ok {
		return o
	}

	delete(unstructured.Object, "status")

	o = unstructured

	return o
}

// PrintObj copies the object and omits the managed fields from the copied object before printing it.
func (p *Simplify) PrintObj(obj runtime.Object, w io.Writer) error {
	if obj == nil {
		return p.Delegate.PrintObj(obj, w)
	}

	if meta.IsListType(obj) {
		obj = obj.DeepCopyObject()
		_ = meta.EachListItem(obj, func(item runtime.Object) error {
			omitMetadata(item)
			omitStatus(item)

			return nil
		})
	} else if _, err := meta.Accessor(obj); err == nil {
		obj = obj.DeepCopyObject()

		obj = omitMetadata(obj)
		obj = omitStatus(obj)
	}

	return p.Delegate.PrintObj(obj, w)
}
