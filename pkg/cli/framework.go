package cli

import (
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// FrameworkLabel is the CRD label key used to store the original framework name.
// Set at submission time, read by list/top/get commands to distinguish mpi/horovod/deepspeed.
const FrameworkLabel = "arena.io/framework"

// V2LabelSelector is the label selector used to filter v2-created CRDs.
// All v2-submitted jobs carry the arena.io/framework label.
const V2LabelSelector = FrameworkLabel

// setFrameworkLabel sets the arena.io/framework label on a CRD object.
// Used by run.go and submit.go to preserve the original framework name.
func setFrameworkLabel(crd *unstructured.Unstructured, framework string) {
	labels := crd.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}
	labels[FrameworkLabel] = framework
	crd.SetLabels(labels)
}

// kindToFramework maps a CRD kind to its framework name.
func kindToFramework(kind string) string {
	fw := KindToFramework(kind)
	if fw == "" {
		return strings.ToLower(kind)
	}
	return fw
}
