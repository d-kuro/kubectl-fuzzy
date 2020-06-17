package fuzzyfinder

import (
	"fmt"

	"github.com/ktr0731/go-fuzzyfinder"
	corev1 "k8s.io/api/core/v1"
)

func Pods(pods []corev1.Pod, allNamespaces bool) (corev1.Pod, error) {
	idx, err := fuzzyfinder.Find(pods,
		func(i int) string {
			if allNamespaces {
				return fmt.Sprintf("%s (%s)", pods[i].Name, pods[i].Namespace)
			}
			return pods[i].Name
		})
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
