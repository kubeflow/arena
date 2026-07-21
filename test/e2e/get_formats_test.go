package e2e_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v3"
)

var _ = Describe("Get output formats", func() {
	var (
		jobName   string
		namespace string
	)

	BeforeEach(func() {
		jobName = fmt.Sprintf("v2-getfmt-%d", GinkgoRandomSeed())
		namespace = "default"

		var out bytes.Buffer
		submitCmd := exec.Command(arenaV2Bin, "job", "submit", "pytorch",
			"--name", jobName,
			"--namespace", namespace,
			"--image", "docker.io/library/pytorch:2.1",
			"--workers", "1",
			"echo hello")
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

	It("should output JSON with job get -o json", func() {
		var stdout, stderr bytes.Buffer
		cmd := exec.Command(arenaV2Bin, "job", "get", jobName,
			"--namespace", namespace, "-o", "json")
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		Expect(err).NotTo(HaveOccurred(), "job get -o json failed: %s", stderr.String())

		var parsed map[string]interface{}
		err = json.Unmarshal(stdout.Bytes(), &parsed)
		Expect(err).NotTo(HaveOccurred(), "output should be valid JSON: %s", stdout.String())

		status, ok := parsed["status"].(map[string]interface{})
		Expect(ok).To(BeTrue(), "parsed output should have a status field")
		Expect(status["name"]).To(Equal(jobName))
	})

	It("should output YAML with job get -o yaml", func() {
		var stdout, stderr bytes.Buffer
		cmd := exec.Command(arenaV2Bin, "job", "get", jobName,
			"--namespace", namespace, "-o", "yaml")
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		Expect(err).NotTo(HaveOccurred(), "job get -o yaml failed: %s", stderr.String())

		var parsed map[string]interface{}
		err = yaml.Unmarshal(stdout.Bytes(), &parsed)
		Expect(err).NotTo(HaveOccurred(), "output should be valid YAML: %s", stdout.String())

		status, ok := parsed["status"].(map[string]interface{})
		Expect(ok).To(BeTrue(), "parsed output should have a status field")
		Expect(status["name"]).To(Equal(jobName))
	})
})
