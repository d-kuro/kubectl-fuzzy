package fuzzyfinder

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/ktr0731/go-fuzzyfinder"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/cli-runtime/pkg/resource"
	"sigs.k8s.io/yaml"

	"github.com/d-kuro/kubectl-fuzzy/pkg/kubernetes/simplify"
)

// Option represents available fuzzy-finding options.
type Option func(*opt)

type opt struct {
	allNamespaces bool
	printer       printers.ResourcePrinter
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
func WithPreview(printer printers.ResourcePrinter) Option {
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
		if opt.rawPreview {
			finderOpts = append(finderOpts, rawInfoPreviewWindow(infos, opt.printer))
		} else {
			finderOpts = append(finderOpts, infoPreviewWindow(infos, opt.printer))
		}
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

func rawInfoPreviewWindow(infos []*resource.Info, printer printers.ResourcePrinter) fuzzyfinder.Option {
	return fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
		if i >= 0 {
			buf := &bytes.Buffer{}
			if err := printer.PrintObj(infos[i].Object, buf); err != nil {
				return fmt.Sprintf("error: %s", err)
			}

			return strings.TrimPrefix(buf.String(), "---\n")
		}

		return ""
	})
}

func infoPreviewWindow(infos []*resource.Info, printer printers.ResourcePrinter) fuzzyfinder.Option {
	jsonPrinter := &printers.JSONPrinter{}

	return fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
		if i >= 0 {
			buf := &bytes.Buffer{}
			if err := jsonPrinter.PrintObj(infos[i].Object, buf); err != nil {
				return fmt.Sprintf("error: %s", err)
			}

			simplified, err := simplifyObject(buf.String(), printer)
			if err != nil {
				return fmt.Sprintf("error: %s", err)
			}

			return simplified
		}

		return ""
	})
}

func simplifyObject(jsonObj string, printer printers.ResourcePrinter) (string, error) {
	simplified, err := simplify.Transform(jsonObj)
	if err != nil {
		return "", err
	}

	return convert([]byte(simplified), printer)
}

func convert(jsonObj []byte, printer printers.ResourcePrinter) (string, error) {
	switch printer.(type) {
	case *printers.JSONPrinter:
		return string(jsonObj), nil
	case *printers.YAMLPrinter:
		y, err := yaml.JSONToYAML(jsonObj)
		if err != nil {
			return "", fmt.Errorf("failed to convert JSON to YAML: %w", err)
		}

		return string(y), nil
	default:
		return "", fmt.Errorf("unsupported printer type: %T", printer)
	}
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
