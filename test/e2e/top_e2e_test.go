package e2e_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	outputpkg "github.com/kubeflow/arena/pkg/output"
)

var _ = Describe("Top Job", func() {
	var (
		jobName   string
		namespace string
	)

	BeforeEach(func() {
		jobName = fmt.Sprintf("v2-top-%d", GinkgoRandomSeed())
		namespace = "default"

		var out bytes.Buffer
		submitCmd := exec.Command(arenaV2Bin, "submit", "pytorch",
			"--name", jobName,
			"--namespace", namespace,
			"--image", "docker.io/library/pytorch:2.1",
			"--workers", "1",
			"python train.py",
		)
		submitCmd.Stdout = &out
		submitCmd.Stderr = &out
		err := submitCmd.Run()
		Expect(err).NotTo(HaveOccurred(), "submit output: %s", out.String())
	})

	AfterEach(func() {
		var out bytes.Buffer
		delCmd := exec.Command(arenaV2Bin, "job", "delete", jobName,
			"--namespace", namespace)
		delCmd.Stdout = &out
		delCmd.Stderr = &out
		_ = delCmd.Run()
	})

	It("should display top job table", func() {
		var out bytes.Buffer
		topCmd := exec.Command(arenaV2Bin, "top", "job",
			"--namespace", namespace)
		topCmd.Stdout = &out
		topCmd.Stderr = &out
		err := topCmd.Run()
		Expect(err).NotTo(HaveOccurred(), "top job output: %s", out.String())
		Expect(out.String()).To(ContainSubstring("NAME"))
	})

	It("should support JSON output", func() {
		var out bytes.Buffer
		topCmd := exec.Command(arenaV2Bin, "top", "job",
			"--namespace", namespace,
			"-o", string(outputpkg.FormatJSON))
		topCmd.Stdout = &out
		topCmd.Stderr = &out
		err := topCmd.Run()
		Expect(err).NotTo(HaveOccurred(), "top job JSON output: %s", out.String())

		var parsed []interface{}
		Expect(json.Unmarshal(out.Bytes(), &parsed)).NotTo(HaveOccurred(),
			"output should be valid JSON array: %s", out.String())
	})
})
