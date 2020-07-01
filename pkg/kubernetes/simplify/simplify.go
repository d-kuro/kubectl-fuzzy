package simplify

import (
	"fmt"

	"github.com/tidwall/sjson"
)

// Transform takes the Kubernetes object JSON and simplifies it, then returns it.
// Remove some of the metadata and status.
func Transform(jsonObj string) (string, error) {
	jsonObj, err := metadata(jsonObj)
	if err != nil {
		return "", err
	}

	jsonObj, err = podTemplateMetadata(jsonObj)
	if err != nil {
		return "", err
	}

	jsonObj, err = status(jsonObj)
	if err != nil {
		return "", nil
	}

	return jsonObj, nil
}

func metadata(in string) (string, error) { //nolint:funlen
	in, err := sjson.Delete(in, `metadata.annotations.kubectl\.kubernetes\.io/last-applied-configuration`)
	if err != nil {
		return in, fmt.Errorf("failed to remove kubectl.kubernetes.io/last-applied-configuration: %w", err)
	}

	in, err = sjson.Delete(in, "metadata.generateName")
	if err != nil {
		return "", fmt.Errorf("failed to remove metadata.generateName: %w", err)
	}

	in, err = sjson.Delete(in, "metadata.selfLink")
	if err != nil {
		return "", fmt.Errorf("failed to remove metadata.selfLink: %w", err)
	}

	in, err = sjson.Delete(in, "metadata.uid")
	if err != nil {
		return "", fmt.Errorf("failed to remove metadata.uid: %w", err)
	}

	in, err = sjson.Delete(in, "metadata.resourceVersion")
	if err != nil {
		return "", fmt.Errorf("failed to remove metadata.resourceVersion: %w", err)
	}

	in, err = sjson.Delete(in, "metadata.generation")
	if err != nil {
		return "", fmt.Errorf("failed to remove metadata.generation: %w", err)
	}

	in, err = sjson.Delete(in, "metadata.creationTimestamp")
	if err != nil {
		return "", fmt.Errorf("failed to remove metadata.creationTimestamp: %w", err)
	}

	in, err = sjson.Delete(in, "metadata.deletionTimestamp")
	if err != nil {
		return "", fmt.Errorf("failed to remove metadata.deletionTimestamp: %w", err)
	}

	in, err = sjson.Delete(in, "metadata.deletionGracePeriodSeconds")
	if err != nil {
		return "", fmt.Errorf("failed to remove metadata.deletionGracePeriodSeconds: %w", err)
	}

	in, err = sjson.Delete(in, "metadata.ownerReferences")
	if err != nil {
		return "", fmt.Errorf("failed to remove metadata.ownerReferences: %w", err)
	}

	in, err = sjson.Delete(in, "metadata.finalizers")
	if err != nil {
		return "", fmt.Errorf("failed to remove metadata.finalizers: %w", err)
	}

	in, err = sjson.Delete(in, "metadata.clusterName")
	if err != nil {
		return "", fmt.Errorf("failed to remove metadata.clusterName: %w", err)
	}

	in, err = sjson.Delete(in, "metadata.managedFields")
	if err != nil {
		return "", fmt.Errorf("failed to remove metadata.managedFields: %w", err)
	}

	return in, nil
}

func podTemplateMetadata(in string) (string, error) {
	in, err := sjson.Delete(in, "spec.template.metadata.creationTimestamp")
	if err != nil {
		return "", fmt.Errorf("failed to remove spec.template.metadata.creationTimestamp: %w", err)
	}

	return in, nil
}

func status(in string) (string, error) {
	in, err := sjson.Delete(in, "status")
	if err != nil {
		return "", fmt.Errorf("failed to remove status: %w", err)
	}

	return in, nil
}
