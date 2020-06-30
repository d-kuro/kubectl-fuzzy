package fuzzyfinder

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/ktr0731/go-fuzzyfinder"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/cli-runtime/pkg/resource"
	"sigs.k8s.io/yaml"

	"github.com/d-kuro/kubectl-fuzzy/pkg/kubernetes/simplify"
)

func Pods(pods []corev1.Pod, printer printers.ResourcePrinter, allNs bool, raw bool) (corev1.Pod, error) {
	var opts []fuzzyfinder.Option

	if printer != nil {
		if raw {
			opts = append(opts, rawPodPreviewWindow(pods, printer))
		} else {
			opts = append(opts, podPreviewWindow(pods, printer))
		}
	}

	idx, err := fuzzyfinder.Find(pods,
		func(i int) string {
			if allNs {
				return fmt.Sprintf("%s (%s)", pods[i].Name, pods[i].Namespace)
			}
			return pods[i].Name
		},
		opts...,
	)
	if err != nil {
		return corev1.Pod{}, err
	}

	return pods[idx], nil
}

func rawPodPreviewWindow(pods []corev1.Pod, printer printers.ResourcePrinter) fuzzyfinder.Option {
	gvk := schema.GroupVersionKind{Kind: "Pod", Version: "v1"}

	return fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
		if i >= 0 {
			buf := &bytes.Buffer{}
			pods[i].GetObjectKind().SetGroupVersionKind(gvk)
			if err := printer.PrintObj(&pods[i], buf); err != nil {
				return fmt.Sprintf("error: %s", err)
			}
			return strings.TrimPrefix(buf.String(), "---\n")
		}
		return ""
	})
}

func podPreviewWindow(pods []corev1.Pod, printer printers.ResourcePrinter) fuzzyfinder.Option {
	gvk := schema.GroupVersionKind{Kind: "Pod", Version: "v1"}
	jsonPrinter := &printers.JSONPrinter{}

	return fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
		if i >= 0 {
			buf := &bytes.Buffer{}
			pods[i].GetObjectKind().SetGroupVersionKind(gvk)
			if err := jsonPrinter.PrintObj(&pods[i], buf); err != nil {
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

func Infos(infos []*resource.Info, printer printers.ResourcePrinter, allNs bool, raw bool) (*resource.Info, error) {
	var opts []fuzzyfinder.Option

	if printer != nil {
		if raw {
			opts = append(opts, rawInfoPreviewWindow(infos, printer))
		} else {
			opts = append(opts, infoPreviewWindow(infos, printer))
		}
	}

	idx, err := fuzzyfinder.Find(infos,
		func(i int) string {
			if allNs && len(infos[i].Namespace) >= 1 {
				return fmt.Sprintf("%s (%s)", infos[i].Name, infos[i].Namespace)
			}
			return infos[i].Name
		},
		opts...,
	)
	if err != nil {
		return nil, err
	}

	return infos[idx], nil
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
