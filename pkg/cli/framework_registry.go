package cli

import (
	"strings"

	"github.com/kubeflow/arena/pkg/constants"
)

// normalizeFramework maps framework aliases to canonical names.
func normalizeFramework(s string) string {
	switch strings.ToLower(s) {
	case constants.FrameworkPyTorch, "pytorchjob":
		return constants.FrameworkPyTorch
	case constants.FrameworkTensorFlow, "tfjob", "tf":
		return constants.FrameworkTensorFlow
	case constants.FrameworkMPI, "mpijob", constants.FrameworkHorovod, constants.FrameworkDeepSpeed:
		return constants.FrameworkMPI
	case constants.FrameworkRay:
		return constants.FrameworkRay
	default:
		return ""
	}
}

// originalFramework preserves the user's original framework choice for labeling.
func originalFramework(s string) string {
	switch strings.ToLower(s) {
	case constants.FrameworkPyTorch, "pytorchjob":
		return constants.FrameworkPyTorch
	case constants.FrameworkTensorFlow, "tfjob", "tf":
		return constants.FrameworkTensorFlow
	case constants.FrameworkHorovod:
		return constants.FrameworkHorovod
	case constants.FrameworkDeepSpeed:
		return constants.FrameworkDeepSpeed
	case constants.FrameworkMPI, "mpijob":
		return constants.FrameworkMPI
	case constants.FrameworkRay:
		return constants.FrameworkRay
	default:
		return ""
	}
}

// frameworkToKind maps a canonical framework name to its CRD kind.
func frameworkToKind(framework string) string {
	switch framework {
	case constants.FrameworkPyTorch:
		return constants.KindPyTorchJob
	case constants.FrameworkTensorFlow, "tf":
		return constants.KindTFJob
	case constants.FrameworkMPI, constants.FrameworkHorovod, constants.FrameworkDeepSpeed:
		return constants.KindMPIJob
	default:
		return ""
	}
}

// kindToFramework maps a CRD kind to its canonical framework name.
// For unrecognized kinds, it returns the lowercased kind as a fallback.
func kindToFramework(kind string) string {
	switch kind {
	case constants.KindPyTorchJob:
		return constants.FrameworkPyTorch
	case constants.KindTFJob:
		return constants.FrameworkTensorFlow
	case constants.KindMPIJob:
		return constants.FrameworkMPI
	default:
		return strings.ToLower(kind)
	}
}

// isMPIFamily returns true if the framework uses the MPIJob CRD.
func isMPIFamily(framework string) bool {
	return framework == constants.FrameworkMPI ||
		framework == constants.FrameworkHorovod ||
		framework == constants.FrameworkDeepSpeed
}
