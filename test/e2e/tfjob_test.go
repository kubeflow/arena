// TensorFlow lifecycle tests: CRUD smoke tests that verify job submit, list,
// get, and delete operations using placeholder images. These do not wait
// for pod readiness or validate training outcomes.

package e2e_test

import (
	"bytes"
	"fmt"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TensorFlow Job", func() {
	var (
		jobName   string
		namespace string
	)

	BeforeEach(func() {
		jobName = fmt.Sprintf("v2-tf-%d", GinkgoRandomSeed())
		namespace = "default"
	})

	AfterEach(func() {
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

		By("Submitting a TensorFlow job")
		submitCmd := exec.Command(arenaV2Bin, "submit", "tensorflow",
			"--name", jobName,
			"--namespace", namespace,
			"--image", "docker.io/library/tensorflow:2.15",
			"--workers", "2",
			"python train.py",
		)
		submitCmd.Stdout = &out
		submitCmd.Stderr = &out
		err = submitCmd.Run()
		Expect(err).NotTo(HaveOccurred(), "submit output: %s", out.String())
		out.Reset()

		By("Listing jobs")
		listCmd := exec.Command(arenaV2Bin, "job", "list",
			"--namespace", namespace)
		listCmd.Stdout = &out
		listCmd.Stderr = &out
		err = listCmd.Run()
		Expect(err).NotTo(HaveOccurred())
		Expect(out.String()).To(ContainSubstring(jobName))
		out.Reset()

		By("Getting job details")
		getCmd := exec.Command(arenaV2Bin, "job", "get", jobName,
			"--namespace", namespace)
		getCmd.Stdout = &out
		getCmd.Stderr = &out
		err = getCmd.Run()
		Expect(err).NotTo(HaveOccurred())
		out.Reset()

		By("Deleting the job")
		delCmd := exec.Command(arenaV2Bin, "job", "delete", jobName,
			"--namespace", namespace)
		delCmd.Stdout = &out
		delCmd.Stderr = &out
		err = delCmd.Run()
		Expect(err).NotTo(HaveOccurred())
	})
})
