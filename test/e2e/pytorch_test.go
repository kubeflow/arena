// PyTorch lifecycle tests: CRUD smoke tests that verify job submit, list,
// get, and delete operations using placeholder images. These do not wait
// for pod readiness or validate training outcomes.

package e2e_test

import (
	"bytes"
	"fmt"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	outputpkg "github.com/kubeflow/arena/pkg/output"
)

var _ = Describe("PyTorch Job", func() {
	var (
		jobName   string
		namespace string
	)

	BeforeEach(func() {
		jobName = fmt.Sprintf("v2-pytorch-%d", GinkgoRandomSeed())
		namespace = "default"
	})

	AfterEach(func() {
		// Cleanup: delete the job if it exists
		var out bytes.Buffer
		delCmd := exec.Command(arenaV2Bin, "job", "delete", jobName,
			"--namespace", namespace)
		delCmd.Stdout = &out
		delCmd.Stderr = &out
		_ = delCmd.Run()
	})

	It("should submit, list, get, and delete successfully", func() {
		var out bytes.Buffer
		var err error

		By("Submitting a PyTorch job")
		submitCmd := exec.Command(arenaV2Bin, "job", "submit", "pytorch",
			"--name", jobName,
			"--namespace", namespace,
			"--image", "pytorch:2.1",
			"--workers", "2",
			"python train.py",
		)
		submitCmd.Stdout = &out
		submitCmd.Stderr = &out
		err = submitCmd.Run()
		Expect(err).NotTo(HaveOccurred(), "submit output: %s", out.String())
		out.Reset()

		By("Listing jobs and verifying the job appears")
		listCmd := exec.Command(arenaV2Bin, "job", "list",
			"--namespace", namespace,
			"-o", string(outputpkg.FormatJSON))
		listCmd.Stdout = &out
		listCmd.Stderr = &out
		err = listCmd.Run()
		Expect(err).NotTo(HaveOccurred(), "list output: %s", out.String())
		Expect(out.String()).To(ContainSubstring(jobName))
		out.Reset()

		By("Getting job details")
		getCmd := exec.Command(arenaV2Bin, "job", "get", jobName,
			"--namespace", namespace)
		getCmd.Stdout = &out
		getCmd.Stderr = &out
		err = getCmd.Run()
		Expect(err).NotTo(HaveOccurred(), "get output: %s", out.String())
		out.Reset()

		By("Deleting the job")
		delCmd := exec.Command(arenaV2Bin, "job", "delete", jobName,
			"--namespace", namespace)
		delCmd.Stdout = &out
		delCmd.Stderr = &out
		err = delCmd.Run()
		Expect(err).NotTo(HaveOccurred(), "delete output: %s", out.String())
	})
})
