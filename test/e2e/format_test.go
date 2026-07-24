package e2e_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v3"

	outputpkg "github.com/kubeflow/arena/pkg/output"
)

var _ = Describe("Output Formats", func() {
	var (
		jobName   string
		namespace string
	)

	BeforeEach(func() {
		jobName = fmt.Sprintf("v2-fmt-%d", GinkgoRandomSeed())
		namespace = "default"

		var out bytes.Buffer
		submitCmd := exec.Command(arenaV2Bin, "submit", "pytorch",
			"--name", jobName,
			"--namespace", namespace,
			"--image", "docker.io/library/pytorch:2.1",
			"--workers", "1",
			"echo hello",
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

	It("should support table format", func() {
		var stdout, stderr bytes.Buffer
		cmd := exec.Command(arenaV2Bin, "job", "list",
			"--namespace", namespace, "-o", string(outputpkg.FormatTable))
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		Expect(err).NotTo(HaveOccurred(), "command failed: %s", stderr.String())
		Expect(stdout.String()).To(ContainSubstring("NAME"))
	})

	It("should support JSON format", func() {
		var stdout, stderr bytes.Buffer
		cmd := exec.Command(arenaV2Bin, "job", "list",
			"--namespace", namespace, "-o", string(outputpkg.FormatJSON))
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		Expect(err).NotTo(HaveOccurred(), "command failed: %s", stderr.String())

		var parsed []interface{}
		err = json.Unmarshal(stdout.Bytes(), &parsed)
		Expect(err).NotTo(HaveOccurred(), "output should be valid JSON")
	})

	It("should support YAML format", func() {
		var stdout, stderr bytes.Buffer
		cmd := exec.Command(arenaV2Bin, "job", "list",
			"--namespace", namespace, "-o", string(outputpkg.FormatYAML))
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		Expect(err).NotTo(HaveOccurred(), "command failed: %s", stderr.String())

		var parsed []interface{}
		err = yaml.Unmarshal(stdout.Bytes(), &parsed)
		Expect(err).NotTo(HaveOccurred(), "output should be valid YAML")
	})

	It("should support wide format", func() {
		var stdout, stderr bytes.Buffer
		cmd := exec.Command(arenaV2Bin, "job", "list",
			"--namespace", namespace, "-o", string(outputpkg.FormatWide))
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		Expect(err).NotTo(HaveOccurred(), "command failed: %s", stderr.String())
		Expect(stdout.String()).To(ContainSubstring("NAME"))
		Expect(stdout.String()).To(ContainSubstring("GPU"))
	})
})
