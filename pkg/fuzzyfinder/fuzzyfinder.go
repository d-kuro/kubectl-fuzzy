package fuzzyfinder

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/ktr0731/go-fuzzyfinder"

	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/duration"
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

	printWithKind := multipleGVKsRequested(infos)

	idx, err := fuzzyfinder.Find(infos,
		func(i int) string {
			var b strings.Builder

			if printWithKind {
				fmt.Fprintf(&b, "%s/", strings.ToLower(infos[i].Mapping.GroupVersionKind.GroupKind().String()))
			}

			fmt.Fprintf(&b, infos[i].Name)

			if allNs && len(infos[i].Namespace) >= 1 {
				fmt.Fprintf(&b, " (%s)", infos[i].Namespace)
			}

			return b.String()
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

func multipleGVKsRequested(infos []*resource.Info) bool {
	if len(infos) < 2 {
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

// CronJobs return a cronjob after fuzzyfinder.
func CronJobs(cronJobs []batchv1beta1.CronJob) (batchv1beta1.CronJob, error) {
	var opts []fuzzyfinder.Option
	opts = append(opts, cronJobPreviewWindow(cronJobs))

	idx, err := fuzzyfinder.Find(cronJobs,
		func(i int) string {
			return cronJobs[i].Name
		},
		opts...,
	)
	if err != nil {
		return batchv1beta1.CronJob{}, err
	}

	return cronJobs[idx], nil
}

func cronJobPreviewWindow(cronJobs []batchv1beta1.CronJob) fuzzyfinder.Option {
	return fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
		if i >= 0 {
			cj := cronJobs[i]
			lastScheduleTime := "<none>"
			if cj.Status.LastScheduleTime != nil {
				lastScheduleTime = translateTimestampSince(*cj.Status.LastScheduleTime)
			}
			return fmt.Sprintf(
				"%s\n\nSCHEDULE: %s\nSUPEND: %v\nACTIVE: %d\nLAST SCHEDULE: %s\nAGE: %s",
				cj.Name,
				cj.Spec.Schedule,
				*cj.Spec.Suspend,
				len(cj.Status.Active),
				lastScheduleTime,
				translateTimestampSince(cj.CreationTimestamp),
			)
		}
		return ""
	})
}

// translateTimestampSince returns the elapsed time since timestamp in
// human-readable approximation.
func translateTimestampSince(timestamp metav1.Time) string {
	if timestamp.IsZero() {
		return "<unknown>"
	}

	return duration.HumanDuration(time.Since(timestamp.Time))
}
