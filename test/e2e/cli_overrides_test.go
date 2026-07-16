package e2e_test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("get --details", func() {
	var (
		jobName   string
		namespace string
	)

	BeforeEach(func() {
		jobName = fmt.Sprintf("v2-details-%d", GinkgoRandomSeed())
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

	It("should show detailed job information", func() {
		var out bytes.Buffer
		submitCmd := exec.Command(arenaV2Bin, "job", "submit", "pytorch",
			"--name", jobName,
			"--namespace", namespace,
			"--image", "pytorch:2.1",
			"--workers", "2",
			"python train.py")
		submitCmd.Stdout = &out
		submitCmd.Stderr = &out
		err := submitCmd.Run()
		Expect(err).NotTo(HaveOccurred(), "submit output: %s", out.String())

		out.Reset()
		getCmd := exec.Command(arenaV2Bin, "job", "get", jobName,
			"--namespace", namespace, "--details")
		getCmd.Stdout = &out
		getCmd.Stderr = &out
		err = getCmd.Run()
		Expect(err).NotTo(HaveOccurred(), "get --details output: %s", out.String())

		output := out.String()
		Expect(output).To(ContainSubstring(jobName))
	})
})

var _ = Describe("CLI overrides", func() {
	var (
		jobName   string
		namespace string
		yamlPath  string
	)

	BeforeEach(func() {
		jobName = fmt.Sprintf("v2-override-%d", GinkgoRandomSeed())
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

	It("should override YAML values with CLI flags in run -f", func() {
		yamlContent := fmt.Sprintf(`version: 0.1.0
name: %s
image: pytorch:2.1
framework:
  name: pytorch
worker:
  replicas: 1
run: python train.py
`, jobName)

		var err error
		yamlPath, err = createTempYAML(yamlContent)
		Expect(err).NotTo(HaveOccurred())

		var out bytes.Buffer
		runCmd := exec.Command(arenaV2Bin, "job", "run", "-f", yamlPath,
			"--namespace", namespace,
			"--set", "worker.replicas=4",
			"--set", "image=custom/pytorch:latest",
			"--dry-run")
		runCmd.Stdout = &out
		runCmd.Stderr = &out
		err = runCmd.Run()
		Expect(err).NotTo(HaveOccurred(), "run -f with overrides output: %s", out.String())

		output := out.String()
		Expect(output).To(ContainSubstring("custom/pytorch:latest"))
	})

	It("should override framework options via CLI", func() {
		yamlContent := fmt.Sprintf(`version: 0.1.0
name: %s
image: pytorch:2.1
framework:
  name: pytorch
worker:
  replicas: 1
run: python train.py
`, jobName)

		var err error
		yamlPath, err = createTempYAML(yamlContent)
		Expect(err).NotTo(HaveOccurred())

		var out bytes.Buffer
		runCmd := exec.Command(arenaV2Bin, "job", "run", "-f", yamlPath,
			"--namespace", namespace,
			"--set", "worker.resources.'nvidia.com/gpu'=2",
			"--set", "worker.resources.cpu=4",
			"--set", "worker.resources.memory=16Gi",
			"--dry-run")
		runCmd.Stdout = &out
		runCmd.Stderr = &out
		err = runCmd.Run()
		Expect(err).NotTo(HaveOccurred(), "run -f with resource overrides output: %s", out.String())

		output := out.String()
		Expect(output).To(ContainSubstring("nvidia.com/gpu"))
		Expect(output).To(ContainSubstring("16Gi"))
	})
})
