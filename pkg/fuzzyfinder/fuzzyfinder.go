package fuzzyfinder

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/d-kuro/kubectl-fuzzy/pkg/printers"
	"github.com/ktr0731/go-fuzzyfinder"
	corev1 "k8s.io/api/core/v1"
	kprinters "k8s.io/cli-runtime/pkg/printers"
	"k8s.io/cli-runtime/pkg/resource"
)

// Option represents available fuzzy-finding options.
type Option func(*opt)

type opt struct {
	allNamespaces bool
	printer       kprinters.ResourcePrinter
	rawPreview    bool
}

// WithAllNamespaces specifies whether to display the namespace during fuzzy-finding.
// Default is false.
func WithAllNamespaces(allNamespace bool) Option {
	return func(o *opt) {
		o.allNamespaces = allNamespace
	}
}

// WithPreview specifies whether to show a preview during fuzzy-finding.
// The output of the preview is done using the ResourcePrinter.
func WithPreview(printer kprinters.ResourcePrinter) Option {
	return func(o *opt) {
		o.printer = printer
	}
}

// WithRawPreview specifies whether an unsimplified object should be displayed in the fuzzy-finding preview.
// Default is false.
func WithRawPreview(rawPreview bool) Option {
	return func(o *opt) {
		o.rawPreview = rawPreview
	}
}

// Infos will start a fuzzy finder based on the received infos and returns the selected info.
func Infos(infos []*resource.Info, opts ...Option) (*resource.Info, error) {
	var opt opt

	for _, o := range opts {
		o(&opt)
	}

	var finderOpts []fuzzyfinder.Option

	if opt.printer != nil {
		if !opt.rawPreview {
			opt.printer = &printers.Simplify{Delegate: opt.printer}
		}

		finderOpts = append(finderOpts, infoPreviewWindow(infos, opt.printer))
	}

	printWithKind := multipleGVKsRequested(infos)

	idx, err := fuzzyfinder.Find(infos,
		func(i int) string {
			var b strings.Builder

			if printWithKind {
				fmt.Fprintf(&b, "%s/", strings.ToLower(infos[i].Mapping.GroupVersionKind.GroupKind().String()))
			}

			fmt.Fprintf(&b, infos[i].Name)

			if opt.allNamespaces && len(infos[i].Namespace) >= 1 {
				fmt.Fprintf(&b, " (%s)", infos[i].Namespace)
			}

			return b.String()
		},
		finderOpts...,
	)
	if err != nil {
		return nil, err
	}

	return infos[idx], nil
}

func Containers(containers []corev1.Container) (corev1.Container, error) {
	idx, err := fuzzyfinder.Find(containers,
		func(i int) string {
			return containers[i].Name
		})
	if err != nil {
		return corev1.Container{}, err
	}

	return containers[idx], nil
}

func infoPreviewWindow(infos []*resource.Info, printer kprinters.ResourcePrinter) fuzzyfinder.Option {
	return fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
		if i >= 0 {
			buf := &bytes.Buffer{}
			if err := printer.PrintObj(infos[i].Object, buf); err != nil {
				return fmt.Sprintf("error: %s", err)
			}

			// Remove the separator as it is added when using kprinters.YAMLPrinter repeatedly.
			return strings.TrimPrefix(buf.String(), "---\n")
		}

		return ""
	})
}

func multipleGVKsRequested(infos []*resource.Info) bool {
	if len(infos) < 2 { //nolint:gomnd
		return false
	}

	gvk := infos[0].Mapping.GroupVersionKind

	for _, info := range infos {
		if info.Mapping.GroupVersionKind != gvk {
			return true
		}
	}

	return false
}
