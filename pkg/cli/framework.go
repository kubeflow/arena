package cli

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// frameworkLabel is the CRD label key used to store the original framework name.
// Set at submission time, read by list/top/get commands to distinguish mpi/horovod/deepspeed.
const frameworkLabel = "arena.io/framework"

// v2LabelSelector is the label selector used to filter v2-created CRDs.
// All v2-submitted jobs carry the arena.io/framework label.
const v2LabelSelector = frameworkLabel

// setFrameworkLabel sets the arena.io/framework label on a CRD object.
func setFrameworkLabel(crd *unstructured.Unstructured, framework string) {
	labels := crd.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}
	labels[frameworkLabel] = framework
	crd.SetLabels(labels)
}
