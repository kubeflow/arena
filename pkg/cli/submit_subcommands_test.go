package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// resetSubmitCommandState restores the cobra-side flag state of the submit
// command tree (the parent and every framework subcommand) after earlier
// ExecuteWithArgs calls in the same process. pflag never clears a flag's
// Changed bit — nor the implicit --help flag's value — between Parse calls,
// so a subcommand that once saw --name/--image would skip cobra's
// required-flag validation forever, and a command whose --help ran once
// would print help on every later execution. resetSubmitFlags covers the
// bound package variables; this covers the state pflag keeps on the shared
// command objects.
//
// The flags are reached through Lookup on each command's flag set: the
// leaked state lives on per-command pflag.Flag objects — every command in
// the tree registers its own --name and --image, plus cobra's implicit
// --help — so each of the three flags is looked up by name on every
// command, and Lookup's nil return lets the loop skip commands that have
// not (yet) defined one.
func resetSubmitCommandState(t *testing.T) {
	t.Helper()
	cmds := append([]*cobra.Command{submitCmd}, submitCmd.Commands()...)
	for _, cmd := range cmds {
		for _, name := range []string{"name", "image", "help"} {
			f := cmd.Flags().Lookup(name)
			if f == nil {
				continue
			}
			if name == "help" {
				// Cobra checks the help flag's value, not its Changed bit, so
				// resetting the value through the flag's Value is enough.
				_ = f.Value.Set("false")
			}
			f.Changed = false
		}
	}
}

// executeSubmitForJSON runs the CLI with the given args, captures stdout, and
// unmarshals the CRD JSON printed by a --dry-run submit.
func executeSubmitForJSON(t *testing.T, args ...string) map[string]interface{} {
	t.Helper()
	resetSubmitCommandState(t)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	err := ExecuteWithArgs(args)
	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)

	require.NoError(t, err, "submit should succeed, stdout: %s", buf.String())

	var crd map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &crd),
		"dry-run output should be valid JSON, got: %s", buf.String())
	return crd
}

// executeSubmitCaptureStdout runs the CLI with the given args and returns
// everything written to stdout (used for --help output).
func executeSubmitCaptureStdout(t *testing.T, args ...string) string {
	t.Helper()
	resetSubmitCommandState(t)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	err := ExecuteWithArgs(args)
	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)

	require.NoError(t, err)
	return buf.String()
}

func TestSubmitSubcommands_DispatchDryRun(t *testing.T) {
	tests := []struct {
		subcommand string
		kind       string
	}{
		{"pytorchjob", "PyTorchJob"},
		{"tfjob", "TFJob"},
		{"mpijob", "MPIJob"},
		{"horovodjob", "MPIJob"},
		{"deepspeedjob", "MPIJob"},
	}
	for _, tt := range tests {
		t.Run(tt.subcommand, func(t *testing.T) {
			resetSubmitFlags(t)
			crd := executeSubmitForJSON(t, "submit", tt.subcommand,
				"--name", "subcmd-test", "--image", "test:latest",
				"--workers", "2", "--dry-run", "echo", "hi")
			assert.Equal(t, tt.kind, crd["kind"])
		})
	}
}

// TestSubmitSubcommands_SubcommandPathCarriesRunConfig goes one level deeper
// than the kind-only dispatch assertions: a subcommand-path dry-run must
// carry the run command (trailing args) and the parsed --workers flag through
// the shared runSubmit core into the CRD.
func TestSubmitSubcommands_SubcommandPathCarriesRunConfig(t *testing.T) {
	resetSubmitFlags(t)
	crd := executeSubmitForJSON(t, "submit", "pytorchjob",
		"--name", "cfg-test", "--image", "test:latest",
		"--workers", "2", "--dry-run", "python", "train.py")

	replicaSpecs := crd["spec"].(map[string]interface{})["pytorchReplicaSpecs"].(map[string]interface{})
	master, ok := replicaSpecs["Master"].(map[string]interface{})
	require.True(t, ok, "dry-run CRD should carry a Master replica spec: %v", replicaSpecs)
	worker, ok := replicaSpecs["Worker"].(map[string]interface{})
	require.True(t, ok, "--workers 2 should produce a Worker replica spec: %v", replicaSpecs)

	// v1 PyTorch semantics: --workers N means N total processes (1 master +
	// N-1 workers), so --workers 2 yields one worker replica next to the
	// master. The Worker spec existing at all proves the flag reached the
	// shared core (--workers <= 1 would build a master-only CRD).
	assert.Equal(t, float64(1), master["replicas"])
	assert.Equal(t, float64(1), worker["replicas"])

	// The trailing args "python" "train.py" are joined into the run command
	// on the container (run+shell model: command [shell, -c], args [run]).
	masterPodSpec := master["template"].(map[string]interface{})["spec"].(map[string]interface{})
	masterContainer := masterPodSpec["containers"].([]interface{})[0].(map[string]interface{})
	assert.Equal(t, []interface{}{"/bin/sh", "-c"}, masterContainer["command"])
	assert.Equal(t, []interface{}{"python train.py"}, masterContainer["args"])
}

func TestSubmitSubcommands_RejectForeignFlags(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "pytorch rejects TF --ps",
			args: []string{"submit", "pytorch", "--name", "x", "--image", "y", "--dry-run", "--ps", "2", "python", "train.py"},
		},
		{
			name: "pytorch rejects MPI --slots-per-worker",
			args: []string{"submit", "pytorch", "--name", "x", "--image", "y", "--dry-run", "--slots-per-worker", "2", "python", "train.py"},
		},
		{
			name: "tfjob rejects PyTorch --nproc-per-node",
			args: []string{"submit", "tfjob", "--name", "x", "--image", "y", "--dry-run", "--nproc-per-node", "auto", "python", "train.py"},
		},
		{
			name: "mpijob rejects TF --chief",
			args: []string{"submit", "mpijob", "--name", "x", "--image", "y", "--dry-run", "--chief", "python", "train.py"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetSubmitFlags(t)
			err := ExecuteWithArgs(tt.args)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "unknown flag")
		})
	}
}

func TestSubmitSubcommands_HelpShowsOnlyRelevantFlags(t *testing.T) {
	tests := []struct {
		subcommand string
		contains   []string
		omits      []string
	}{
		{
			subcommand: "pytorchjob",
			contains:   []string{"--nproc-per-node", "--name", "--workers"},
			omits:      []string{"--ps", "--chief", "--slots-per-worker", "--gputopology"},
		},
		{
			subcommand: "tfjob",
			contains:   []string{"--ps", "--chief", "--success-policy"},
			omits:      []string{"--nproc-per-node", "--slots-per-worker"},
		},
		{
			subcommand: "mpijob",
			contains:   []string{"--slots-per-worker", "--gputopology", "--mounts-on-launcher"},
			omits:      []string{"--ps", "--chief", "--nproc-per-node"},
		},
		{
			subcommand: "horovodjob",
			contains:   []string{"--slots-per-worker"},
			omits:      []string{"--ps", "--nproc-per-node"},
		},
		{
			subcommand: "deepspeedjob",
			contains:   []string{"--slots-per-worker"},
			omits:      []string{"--ps", "--nproc-per-node"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.subcommand, func(t *testing.T) {
			resetSubmitFlags(t)
			help := executeSubmitCaptureStdout(t, "submit", tt.subcommand, "--help")
			for _, want := range tt.contains {
				assert.Contains(t, help, want)
			}
			for _, notWant := range tt.omits {
				assert.NotContains(t, help, notWant)
			}
		})
	}
}

func TestSubmitSubcommands_AliasMatrixDispatch(t *testing.T) {
	tests := []struct {
		alias string
		kind  string
	}{
		{"pytorch", "PyTorchJob"},
		{"pytorchjob", "PyTorchJob"},
		{"tensorflow", "TFJob"},
		{"tfjob", "TFJob"},
		{"tf", "TFJob"},
		{"mpi", "MPIJob"},
		{"mpijob", "MPIJob"},
		{"mj", "MPIJob"},
		{"horovod", "MPIJob"},
		{"horovodjob", "MPIJob"},
		{"hj", "MPIJob"},
		{"deepspeed", "MPIJob"},
		{"deepspeedjob", "MPIJob"},
		{"dp", "MPIJob"},
	}
	for _, tt := range tests {
		t.Run(tt.alias, func(t *testing.T) {
			resetSubmitFlags(t)
			crd := executeSubmitForJSON(t, "submit", tt.alias,
				"--name", "alias-test", "--image", "test:latest", "--dry-run", "echo", "hi")
			assert.Equal(t, tt.kind, crd["kind"])
		})
	}
}

func TestSubmitSubcommands_FrameworkLabelIsCanonical(t *testing.T) {
	tests := []struct {
		typed string
		want  string
	}{
		{"pytorch", "pytorch"},
		{"pytorchjob", "pytorch"},
		{"tensorflow", "tensorflow"},
		{"tf", "tensorflow"},
		{"tfjob", "tensorflow"},
		{"mpi", "mpi"},
		{"mpijob", "mpi"},
		{"horovod", "horovod"},
		{"deepspeed", "deepspeed"},
	}
	for _, tt := range tests {
		t.Run(tt.typed, func(t *testing.T) {
			resetSubmitFlags(t)
			crd := executeSubmitForJSON(t, "submit", tt.typed,
				"--name", "label-test", "--image", "test:latest", "--dry-run", "echo", "hi")
			labels, ok := crd["metadata"].(map[string]interface{})["labels"].(map[string]interface{})
			require.True(t, ok, "dry-run CRD should carry labels: %v", crd["metadata"])
			assert.Equal(t, tt.want, labels["arena.io/framework"])
		})
	}
}

func TestSubmitSubcommands_ParentFallbackBehaviors(t *testing.T) {
	t.Run("v1-only type with flags still gets the dedicated error", func(t *testing.T) {
		resetSubmitFlags(t)
		err := ExecuteWithArgs([]string{"submit", "ray", "--name", "x", "--image", "y", "python", "train.py"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not supported by arena-v2 yet")
	})

	t.Run("unknown type still gets the unsupported error", func(t *testing.T) {
		resetSubmitFlags(t)
		err := ExecuteWithArgs([]string{"submit", "bogus", "--name", "x", "--image", "y", "python", "train.py"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported framework type")
	})

	t.Run("case variant falls back to the parent and still submits", func(t *testing.T) {
		resetSubmitFlags(t)
		crd := executeSubmitForJSON(t, "submit", "PyTorch",
			"--name", "case-test", "--image", "test:latest", "--dry-run", "echo", "hi")
		assert.Equal(t, "PyTorchJob", crd["kind"])
	})

	t.Run("fallback path missing required flags matches the subcommand error", func(t *testing.T) {
		resetSubmitCommandState(t)
		resetSubmitFlags(t)
		err := ExecuteWithArgs([]string{"submit", "PyTorch", "python", "train.py"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), `required flag(s) "image", "name" not set`)
	})

	t.Run("bare submit prints help instead of erroring", func(t *testing.T) {
		resetSubmitFlags(t)
		help := executeSubmitCaptureStdout(t, "submit")
		for _, sub := range []string{"pytorchjob", "tfjob", "mpijob", "horovodjob", "deepspeedjob"} {
			assert.Contains(t, help, sub, "bare submit help should list the %s subcommand", sub)
		}
	})

	t.Run("flags but no framework prints help instead of erroring", func(t *testing.T) {
		resetSubmitFlags(t)
		help := executeSubmitCaptureStdout(t, "submit", "--name", "x", "--image", "y")
		assert.Contains(t, help, "Available Commands:")
	})

	t.Run("missing required flags still fails before RunE", func(t *testing.T) {
		// Cobra-level required-flag validation is a subcommand contract
		// (--name/--image are marked required on each subcommand, not on the
		// parent fallback parser). Earlier subtests parse --name/--image and
		// pflag never clears the Changed bit, so validation would be skipped
		// if that state leaked in. Reset it so this subtest passes regardless
		// of ordering. (Parent-path missing --name is covered by
		// TestSubmitCmd_NameAndImageRequired via task validation.)
		resetSubmitCommandState(t)
		resetSubmitFlags(t)
		err := ExecuteWithArgs([]string{"submit", "pytorchjob", "python", "train.py"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "required flag")
	})

	t.Run("flags before the type still route to the subcommand", func(t *testing.T) {
		resetSubmitFlags(t)
		crd := executeSubmitForJSON(t, "submit",
			"--name", "order-test", "--image", "test:latest",
			"pytorchjob", "--dry-run", "echo", "hi")
		assert.Equal(t, "PyTorchJob", crd["kind"])
	})
}

func TestSubmitParentHelp_ListsSubcommandsOnly(t *testing.T) {
	resetSubmitFlags(t)
	help := executeSubmitCaptureStdout(t, "submit", "--help")

	for _, sub := range []string{"pytorchjob", "tfjob", "mpijob", "horovodjob", "deepspeedjob"} {
		assert.Contains(t, help, sub, "submit --help should list the %s subcommand", sub)
	}
	// The parent's flag wall must not appear; global (root persistent) flags
	// like --namespace still do, which is fine.
	for _, notWant := range []string{"--nproc-per-node", "--slots-per-worker", "--ps ", "--clean-task-policy", "--share-memory"} {
		assert.NotContains(t, help, notWant)
	}
}
