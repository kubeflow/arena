package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kubeflow/arena/pkg/constants"
)

func TestNormalizeFramework_Registry(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// PyTorch variants
		{"pytorch lowercase", "pytorch", "pytorch"},
		{"pytorch mixed case", "PyTorch", "pytorch"},
		{"pytorchjob", "pytorchjob", "pytorch"},
		{"PyTorchJob", "PyTorchJob", "pytorch"},
		// TensorFlow variants
		{"tensorflow lowercase", "tensorflow", "tensorflow"},
		{"TensorFlow mixed case", "TensorFlow", "tensorflow"},
		{"tfjob", "tfjob", "tensorflow"},
		{"TFJob", "TFJob", "tensorflow"},
		{"tf alias", "tf", "tensorflow"},
		{"TF alias", "TF", "tensorflow"},
		// MPI variants
		{"mpi lowercase", "mpi", "mpi"},
		{"MPI uppercase", "MPI", "mpi"},
		{"mpijob", "mpijob", "mpi"},
		{"MPIJob", "MPIJob", "mpi"},
		{"horovod", "horovod", "horovod"},
		{"Horovod", "Horovod", "horovod"},
		{"deepspeed", "deepspeed", "deepspeed"},
		{"DeepSpeed", "DeepSpeed", "deepspeed"},
		// Ray
		{"ray", "ray", ""},
		{"Ray", "Ray", ""},
		// Unknown/empty
		{"unknown framework", "jax", ""},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, normalizeFramework(tt.input))
		})
	}
}

func TestOriginalFramework_Registry(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// PyTorch
		{"pytorch", "pytorch", "pytorch"},
		{"PyTorch", "PyTorch", "pytorch"},
		{"pytorchjob", "pytorchjob", "pytorch"},
		// TensorFlow
		{"tensorflow", "tensorflow", "tensorflow"},
		{"tfjob", "tfjob", "tensorflow"},
		{"tf", "tf", "tensorflow"},
		// MPI-family preserves distinction
		{"horovod", "horovod", "horovod"},
		{"Horovod", "Horovod", "horovod"},
		{"deepspeed", "deepspeed", "deepspeed"},
		{"DeepSpeed", "DeepSpeed", "deepspeed"},
		{"mpi", "mpi", "mpi"},
		{"MPI", "MPI", "mpi"},
		{"mpijob", "mpijob", "mpi"},
		// Ray
		{"ray", "ray", "ray"},
		// Unknown/empty
		{"unknown", "jax", ""},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, originalFramework(tt.input))
		})
	}
}

func TestFrameworkToKind_Registry(t *testing.T) {
	tests := []struct {
		name      string
		framework string
		expected  string
	}{
		{"pytorch", "pytorch", "PyTorchJob"},
		{"tensorflow", "tensorflow", "TFJob"},
		{"mpi", "mpi", "MPIJob"},
		{"horovod", "horovod", "MPIJob"},
		{"deepspeed", "deepspeed", "MPIJob"},
		{"ray", "ray", ""},
		{"unknown", "jax", ""},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, frameworkToKind(tt.framework))
		})
	}
}

func TestKindToFramework_Registry(t *testing.T) {
	tests := []struct {
		name     string
		kind     string
		expected string
	}{
		{"PyTorchJob", "PyTorchJob", "pytorch"},
		{"TFJob", "TFJob", "tensorflow"},
		{"MPIJob", "MPIJob", "mpi"},
		{"unknown kind", "UnknownJob", "unknownjob"},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, kindToFramework(tt.kind))
		})
	}
}

func TestIsMPIFamily_Registry(t *testing.T) {
	tests := []struct {
		name      string
		framework string
		expected  bool
	}{
		{"mpi", "mpi", true},
		{"horovod", "horovod", true},
		{"deepspeed", "deepspeed", true},
		{"pytorch", "pytorch", false},
		{"tensorflow", "tensorflow", false},
		{"ray", "ray", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, isMPIFamily(tt.framework))
		})
	}
}

func TestLookupFramework_SupportedAliases(t *testing.T) {
	tests := []struct{ input, want string }{
		{"pytorch", constants.FrameworkPyTorch},
		{"pytorchjob", constants.FrameworkPyTorch},
		{"PyTorchJob", constants.FrameworkPyTorch},
		{"tf", constants.FrameworkTensorFlow},
		{"tfjob", constants.FrameworkTensorFlow},
		{"tensorflow", constants.FrameworkTensorFlow},
		{"mpi", constants.FrameworkMPI},
		{"mpijob", constants.FrameworkMPI},
		{"mj", constants.FrameworkMPI},
		{"horovod", constants.FrameworkHorovod},
		{"horovodjob", constants.FrameworkHorovod},
		{"hj", constants.FrameworkHorovod},
		{"deepspeed", constants.FrameworkDeepSpeed},
		{"deepspeedjob", constants.FrameworkDeepSpeed},
		{"dp", constants.FrameworkDeepSpeed},
	}
	for _, tt := range tests {
		canonical, unsupported := lookupFramework(tt.input)
		assert.Equal(t, tt.want, canonical, "input %q", tt.input)
		assert.False(t, unsupported, "input %q should be supported", tt.input)
	}
}

func TestLookupFramework_V1TypesUnsupported(t *testing.T) {
	v1OnlyAliases := []string{
		"ray", "rayjob", "rj",
		"spark", "sparkjob",
		"volcano", "volcanojob", "vj",
		"et", "etjob",
	}
	for _, input := range v1OnlyAliases {
		canonical, unsupported := lookupFramework(input)
		assert.Empty(t, canonical, "input %q", input)
		assert.True(t, unsupported, "input %q should be flagged as an unsupported v1 type", input)
	}
}

func TestLookupFramework_Unknown(t *testing.T) {
	canonical, unsupported := lookupFramework("bogus")
	assert.Empty(t, canonical)
	assert.False(t, unsupported)
}

func TestSubmitFrameworkTypeErrors(t *testing.T) {
	err := submitCmd.RunE(submitCmd, []string{"rayjob"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not supported by arena-v2 yet")

	err = submitCmd.RunE(submitCmd, []string{"bogus"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported framework type")
}
