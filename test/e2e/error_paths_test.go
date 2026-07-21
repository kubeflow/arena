package e2e_test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Error paths", func() {
	var namespace string

	BeforeEach(func() {
		namespace = "default"
	})

	It("should error when --file is missing", func() {
		var out bytes.Buffer
		cmd := exec.Command(arenaV2Bin, "job", "run")
		cmd.Stdout = &out
		cmd.Stderr = &out
		err := cmd.Run()
		Expect(err).To(HaveOccurred(), "run without --file should error")
		Expect(out.String()).To(ContainSubstring("--file is required"))
	})

	It("should error on invalid YAML", func() {
		yamlPath, err := createTempYAML("this is not: valid: yaml: [")
		Expect(err).NotTo(HaveOccurred())
		defer os.Remove(yamlPath)

		var out bytes.Buffer
		cmd := exec.Command(arenaV2Bin, "job", "run", "-f", yamlPath,
			"--namespace", namespace)
		cmd.Stdout = &out
		cmd.Stderr = &out
		err = cmd.Run()
		Expect(err).To(HaveOccurred(), "invalid YAML should error")
		Expect(out.String()).To(SatisfyAny(
			ContainSubstring("failed to parse"),
			ContainSubstring("failed to load"),
		))
	})

	It("should error on missing required name", func() {
		yamlContent := `version: 0.1.0
image: docker.io/library/pytorch:2.1
framework:
  name: pytorch
worker:
  replicas: 1
run: python train.py
`
		yamlPath, err := createTempYAML(yamlContent)
		Expect(err).NotTo(HaveOccurred())
		defer os.Remove(yamlPath)

		var out bytes.Buffer
		cmd := exec.Command(arenaV2Bin, "job", "run", "-f", yamlPath,
			"--namespace", namespace)
		cmd.Stdout = &out
		cmd.Stderr = &out
		err = cmd.Run()
		Expect(err).To(HaveOccurred(), "missing name should error")
		Expect(out.String()).To(ContainSubstring("name is required"))
	})

	It("should error on unsupported framework via submit", func() {
		var out bytes.Buffer
		cmd := exec.Command(arenaV2Bin, "job", "submit", "jax",
			"--name", "no-such-framework",
			"--namespace", namespace,
			"--image", "x:1",
			"run")
		cmd.Stdout = &out
		cmd.Stderr = &out
		err := cmd.Run()
		Expect(err).To(HaveOccurred(), "unsupported framework should error")
		Expect(out.String()).To(ContainSubstring("unsupported framework"))
	})

	It("should error on duplicate job name", func() {
		jobName := fmt.Sprintf("v2-dup-%d", GinkgoRandomSeed())

		By("Submitting the job first time")
		var out bytes.Buffer
		submitCmd := exec.Command(arenaV2Bin, "job", "submit", "pytorch",
			"--name", jobName,
			"--namespace", namespace,
			"--image", "docker.io/library/busybox:1.36",
			"--workers", "1",
			"sleep 300")
		submitCmd.Stdout = &out
		submitCmd.Stderr = &out
		err := submitCmd.Run()
		Expect(err).NotTo(HaveOccurred(), "first submit output: %s", out.String())
		out.Reset()

		defer func() {
			delCmd := exec.Command(arenaV2Bin, "job", "delete", jobName,
				"--namespace", namespace)
			delCmd.Stdout = &out
			delCmd.Stderr = &out
			_ = delCmd.Run()
		}()

		By("Submitting the same job name again")
		submitCmd2 := exec.Command(arenaV2Bin, "job", "submit", "pytorch",
			"--name", jobName,
			"--namespace", namespace,
			"--image", "docker.io/library/busybox:1.36",
			"--workers", "1",
			"sleep 300")
		submitCmd2.Stdout = &out
		submitCmd2.Stderr = &out
		err = submitCmd2.Run()
		Expect(err).To(HaveOccurred(), "duplicate submit should error")
		Expect(out.String()).To(ContainSubstring("already exists"))
	})

	It("should error on non-existent job get", func() {
		missingName := fmt.Sprintf("no-such-job-%d", GinkgoRandomSeed())
		var out bytes.Buffer
		cmd := exec.Command(arenaV2Bin, "job", "get", missingName,
			"--namespace", namespace)
		cmd.Stdout = &out
		cmd.Stderr = &out
		err := cmd.Run()
		Expect(err).To(HaveOccurred(), "get non-existent job should error")
	})

	It("should detect v1 jobs", func() {
		jobName := fmt.Sprintf("v2-v1detect-%d", GinkgoRandomSeed())

		By("Creating a raw v1 PyTorchJob without arena.io/framework label")
		v1YAML := fmt.Sprintf(`apiVersion: kubeflow.org/v1
kind: PyTorchJob
metadata:
  name: %s
  namespace: default
spec:
  pytorchReplicaSpecs:
    Master:
      replicas: 1
      template:
        spec:
          containers:
            - name: pytorch
              image: docker.io/library/busybox:1.36
              command: ["sleep", "300"]
          restartPolicy: OnFailure
`, jobName)

		var out bytes.Buffer
		kubectlCmd := exec.Command("kubectl", "apply", "-f", "-")
		kubectlCmd.Stdin = bytes.NewReader([]byte(v1YAML))
		kubectlCmd.Stdout = &out
		kubectlCmd.Stderr = &out
		err := kubectlCmd.Run()
		Expect(err).NotTo(HaveOccurred(), "kubectl apply output: %s", out.String())
		out.Reset()

		defer func() {
			delCmd := exec.Command("kubectl", "delete", "pytorchjob", jobName,
				"-n", namespace, "--ignore-not-found")
			delCmd.Stdout = &out
			delCmd.Stderr = &out
			_ = delCmd.Run()
		}()

		By("Getting the v1 job — should error")
		getCmd := exec.Command(arenaV2Bin, "job", "get", jobName,
			"--namespace", namespace)
		getCmd.Stdout = &out
		getCmd.Stderr = &out
		err = getCmd.Run()
		Expect(err).To(HaveOccurred(), "v1 job get should error")
		Expect(out.String()).To(ContainSubstring("arena v1"))
	})
})
