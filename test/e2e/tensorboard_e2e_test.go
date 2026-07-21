package e2e_test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TensorBoard", func() {
	var (
		jobName   string
		namespace string
		yamlPath  string
	)

	BeforeEach(func() {
		jobName = fmt.Sprintf("v2-tb-%d", GinkgoRandomSeed())
		namespace = "default"
	})

	AfterEach(func() {
		if yamlPath != "" {
			os.Remove(yamlPath)
		}
	})

	It("should include TensorBoard resources in submit dry-run", func() {
		var stdout, stderr bytes.Buffer
		cmd := exec.Command(arenaV2Bin, "job", "submit", "pytorch",
			"--name", jobName,
			"--namespace", namespace,
			"--image", "docker.io/library/pytorch:2.1",
			"--workers", "1",
			"--tensorboard",
			"--tensorboard-logdir", "/logs",
			"--dry-run",
			"python train.py")
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		Expect(err).NotTo(HaveOccurred(), "submit --tensorboard --dry-run failed: %s", stderr.String())

		output := stdout.String()
		Expect(output).To(ContainSubstring("---"), "output should contain resource separator")
		Expect(output).To(ContainSubstring("Deployment"), "output should contain TensorBoard Deployment")
		Expect(output).To(ContainSubstring("Service"), "output should contain TensorBoard Service")
	})

	It("should support TensorBoard via YAML run --dry-run", func() {
		yamlContent := fmt.Sprintf(`version: 0.1.0
name: %s
image: docker.io/library/pytorch:2.1
framework:
  name: pytorch
worker:
  replicas: 1
run: python train.py
logging:
  tensorboard:
    enabled: true
    logdir: /logs
`, jobName)

		var err error
		yamlPath, err = createTempYAML(yamlContent)
		Expect(err).NotTo(HaveOccurred())

		var stdout, stderr bytes.Buffer
		cmd := exec.Command(arenaV2Bin, "job", "run", "-f", yamlPath,
			"--namespace", namespace,
			"--dry-run")
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err = cmd.Run()
		Expect(err).NotTo(HaveOccurred(), "run -f with tensorboard --dry-run failed: %s", stderr.String())

		output := stdout.String()
		Expect(output).To(ContainSubstring("---"), "output should contain resource separator")
		Expect(output).To(ContainSubstring("Service"), "output should contain TensorBoard Service")
	})
})
