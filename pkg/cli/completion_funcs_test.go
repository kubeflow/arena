package cli

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestCompleteFrameworkType(t *testing.T) {
	completions, directive := completeFrameworkType(nil, nil, "")
	assert.Len(t, completions, 10)
	assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
	expected := []string{
		"pytorch\tPyTorch", "pytorchjob\tPyTorch",
		"tensorflow\tTensorFlow", "tfjob\tTensorFlow", "tf\tTensorFlow",
		"mpi\tMPI", "mpijob\tMPI",
		"horovod\tHorovod", "deepspeed\tDeepSpeed", "ray\tRay",
	}
	assert.Equal(t, expected, completions)
}

func TestCompleteFrameworkTypeWithArgs(t *testing.T) {
	completions, directive := completeFrameworkType(nil, []string{"pytorch"}, "")
	assert.Nil(t, completions)
	assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
}

func TestCompleteOutputFormat(t *testing.T) {
	completions, directive := completeOutputFormat(nil, nil, "")
	assert.Equal(t, []string{
		"table\tTable format", "wide\tWide table",
		"json\tJSON output", "yaml\tYAML output",
	}, completions)
	assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
}

func TestCompleteFile(t *testing.T) {
	completions, directive := completeFile(nil, nil, "")
	assert.Nil(t, completions)
	assert.Equal(t, cobra.ShellCompDirectiveDefault, directive)
}

func TestCompleteStaticChoices(t *testing.T) {
	fn := completeStaticChoices("a\tAlpha", "b\tBeta", "c")
	completions, directive := fn(nil, nil, "")
	assert.Equal(t, []string{"a\tAlpha", "b\tBeta", "c"}, completions)
	assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
}

func TestCompleteJobNameMultipleArgs(t *testing.T) {
	completions, directive := completeJobName(nil, []string{"existing-job"}, "")
	assert.Nil(t, completions)
	assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
}

func TestCompleteFrameworkTypeHasDescriptions(t *testing.T) {
	completions, _ := completeFrameworkType(nil, nil, "")
	for _, c := range completions {
		assert.Contains(t, c, "\t",
			"completion %q should contain a tab separator for description", c)
	}
}
