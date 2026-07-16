package e2e_test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("run -f (YAML file submission)", func() {
	var (
		jobName   string
		namespace string
		yamlPath  string
	)

	BeforeEach(func() {
		jobName = fmt.Sprintf("v2-yaml-submit-%d", GinkgoRandomSeed())
		namespace = "default"
	})

	AfterEach(func() {
		var out bytes.Buffer
		delCmd := exec.Command(arenaV2Bin, "job", "delete", jobName,
			"--namespace", namespace)
		delCmd.Stdout = &out
		delCmd.Stderr = &out
		_ = delCmd.Run()

		if yamlPath != "" {
			os.Remove(yamlPath)
		}
	})

	It("should submit a job from YAML file", func() {
		yamlContent := fmt.Sprintf(`version: 0.1.0
name: %s
image: pytorch:2.1
framework:
  name: pytorch
worker:
  replicas: 1
run: echo "hello from yaml"
`, jobName)

		var err error
		yamlPath, err = createTempYAML(yamlContent)
		Expect(err).NotTo(HaveOccurred())

		var out bytes.Buffer
		runCmd := exec.Command(arenaV2Bin, "job", "run", "-f", yamlPath,
			"--namespace", namespace)
		runCmd.Stdout = &out
		runCmd.Stderr = &out
		err = runCmd.Run()
		Expect(err).NotTo(HaveOccurred(), "run -f output: %s", out.String())

		out.Reset()
		getCmd := exec.Command(arenaV2Bin, "job", "get", jobName,
			"--namespace", namespace)
		getCmd.Stdout = &out
		getCmd.Stderr = &out
		err = getCmd.Run()
		Expect(err).NotTo(HaveOccurred(), "get output: %s", out.String())
	})

	It("should support dry-run with YAML file", func() {
		yamlContent := fmt.Sprintf(`version: 0.1.0
name: %s
image: pytorch:2.1
framework:
  name: pytorch
worker:
  replicas: 2
  resources:
    nvidia.com/gpu: "1"
run: python train.py
`, jobName)

		var err error
		yamlPath, err = createTempYAML(yamlContent)
		Expect(err).NotTo(HaveOccurred())

		var out bytes.Buffer
		runCmd := exec.Command(arenaV2Bin, "job", "run", "-f", yamlPath,
			"--dry-run")
		runCmd.Stdout = &out
		runCmd.Stderr = &out
		err = runCmd.Run()
		Expect(err).NotTo(HaveOccurred(), "dry-run output: %s", out.String())

		output := out.String()
		Expect(output).To(ContainSubstring("PyTorchJob"))
		Expect(output).To(ContainSubstring(jobName))
		Expect(output).To(ContainSubstring("nvidia.com/gpu"))
	})
})
