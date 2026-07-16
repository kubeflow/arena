package cli

import (
	"strings"

	"github.com/kubeflow/arena/pkg/constants"
)

// NormalizeFramework maps framework aliases to canonical names.
func NormalizeFramework(s string) string {
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

// OriginalFramework preserves the user's original framework choice for labeling.
func OriginalFramework(s string) string {
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

// FrameworkToKind maps a canonical framework name to its CRD kind.
func FrameworkToKind(framework string) string {
	switch framework {
	case constants.FrameworkPyTorch:
		return constants.KindPyTorchJob
	case constants.FrameworkTensorFlow:
		return constants.KindTFJob
	case constants.FrameworkMPI, constants.FrameworkHorovod, constants.FrameworkDeepSpeed:
		return constants.KindMPIJob
	default:
		return ""
	}
}

// KindToFramework maps a CRD kind to its canonical framework name.
func KindToFramework(kind string) string {
	switch kind {
	case constants.KindPyTorchJob:
		return constants.FrameworkPyTorch
	case constants.KindTFJob:
		return constants.FrameworkTensorFlow
	case constants.KindMPIJob:
		return constants.FrameworkMPI
	default:
		return ""
	}
}

// IsMPIFamily returns true if the framework uses the MPIJob CRD.
func IsMPIFamily(framework string) bool {
	return framework == constants.FrameworkMPI ||
		framework == constants.FrameworkHorovod ||
		framework == constants.FrameworkDeepSpeed
}
