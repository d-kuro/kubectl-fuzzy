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
)

func Pods(pods []corev1.Pod, allNamespaces bool, printer printers.ResourcePrinter) (corev1.Pod, error) {
	gvk := schema.GroupVersionKind{Kind: "Pod", Version: "v1"}

	var opts []fuzzyfinder.Option

	if printer != nil {
		opts = append(opts, fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i >= 0 {
				buf := &bytes.Buffer{}
				pods[i].GetObjectKind().SetGroupVersionKind(gvk)
				if err := printer.PrintObj(&pods[i], buf); err != nil {
					return fmt.Sprintf("preview display error:\n%s\n%s", err, pods[i].GetObjectKind().GroupVersionKind())
				}
				return strings.TrimPrefix(buf.String(), "---\n")
			}
			return ""
		}))
	}

	idx, err := fuzzyfinder.Find(pods,
		func(i int) string {
			if allNamespaces {
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

func Infos(infos []*resource.Info, allNamespaces bool, printer printers.ResourcePrinter) (*resource.Info, error) {
	var opts []fuzzyfinder.Option

	if printer != nil {
		opts = append(opts, fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i >= 0 {
				buf := &bytes.Buffer{}
				if err := printer.PrintObj(infos[i].Object, buf); err != nil {
					return fmt.Sprintf("preview display error: %s", err)
				}
				return strings.TrimPrefix(buf.String(), "---\n")
			}
			return ""
		}))
	}

	idx, err := fuzzyfinder.Find(infos,
		func(i int) string {
			if allNamespaces {
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
