package e2e_test

import (
	"bytes"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Check Command", func() {
	It("should run arena check and produce structured output", func() {
		var stdout, stderr bytes.Buffer
		cmd := exec.Command(arenaV2Bin, "check")
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()

		output := stdout.String()

		// The check command should produce output for each CRD kind
		// regardless of whether they are installed or not.
		Expect(output).To(ContainSubstring("PyTorchJob"))
		Expect(output).To(ContainSubstring("TFJob"))
		Expect(output).To(ContainSubstring("MPIJob"))

		// Each CRD line should indicate either installed or not installed
		// We don't require a specific outcome (installed vs not) since the
		// e2e cluster may or may not have the CRDs applied.
		for _, kind := range []string{"PyTorchJob", "TFJob", "MPIJob"} {
			// Output should mention each kind with either a check or X mark
			Expect(output).To(SatisfyAny(
				ContainSubstring("✓ "+kind),
				ContainSubstring("✗ "+kind),
			), "output should indicate status for %s", kind)
		}

		// If command succeeded, all CRDs should be installed
		// If command failed, at least one CRD is missing (expected in bare clusters)
		if err != nil {
			// Failure is acceptable if CRDs are not installed
			Expect(output).To(ContainSubstring("✗"))
		}
	})

	It("should report MPIJob version information when installed", func() {
		var stdout, stderr bytes.Buffer
		cmd := exec.Command(arenaV2Bin, "check")
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		_ = cmd.Run() // May fail if CRDs not installed

		output := stdout.String()

		// If MPIJob is installed, output should contain version information
		if bytes.Contains(stdout.Bytes(), []byte("✓ MPIJob")) {
			Expect(output).To(ContainSubstring("versions:"))
			// Should mention the expected API version
			Expect(output).To(ContainSubstring("kubeflow.org"))
		}
	})
})
