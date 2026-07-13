package e2e_test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("delete -f (delete from YAML file)", func() {
	var (
		jobName   string
		namespace string
		yamlPath  string
	)

	BeforeEach(func() {
		jobName = fmt.Sprintf("v2-yaml-del-%d", GinkgoRandomSeed())
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

	It("should delete a job using YAML file", func() {
		var out bytes.Buffer
		submitCmd := exec.Command(arenaV2Bin, "job", "submit", "pytorch",
			"--name", jobName,
			"--namespace", namespace,
			"--image", "pytorch:2.1",
			"--workers", "1",
			"echo test")
		submitCmd.Stdout = &out
		submitCmd.Stderr = &out
		err := submitCmd.Run()
		Expect(err).NotTo(HaveOccurred(), "submit output: %s", out.String())

		yamlContent := fmt.Sprintf(`version: 0.1.0
name: %s
image: pytorch:2.1
framework:
  name: pytorch
worker:
  replicas: 1
run: echo test
`, jobName)

		var err2 error
		yamlPath, err2 = createTempYAML(yamlContent)
		Expect(err2).NotTo(HaveOccurred())

		out.Reset()
		delCmd := exec.Command(arenaV2Bin, "job", "delete", "-f", yamlPath,
			"--namespace", namespace)
		delCmd.Stdout = &out
		delCmd.Stderr = &out
		err = delCmd.Run()
		Expect(err).NotTo(HaveOccurred(), "delete -f output: %s", out.String())

		out.Reset()
		getCmd := exec.Command(arenaV2Bin, "job", "get", jobName,
			"--namespace", namespace)
		getCmd.Stdout = &out
		getCmd.Stderr = &out
		err = getCmd.Run()
		Expect(err).To(HaveOccurred(), "job should be deleted")
	})
})
