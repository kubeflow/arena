package e2e_test

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Logs", func() {
	var (
		jobName   string
		namespace string
	)

	BeforeEach(func() {
		jobName = fmt.Sprintf("v2-logs-%d", GinkgoRandomSeed())
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

	It("should fetch logs from a job pod", func() {
		var out bytes.Buffer

		By("Submitting a PyTorch job that prints output")
		submitCmd := exec.Command(arenaV2Bin, "job", "submit", "pytorch",
			"--name", jobName,
			"--namespace", namespace,
			"--image", "docker.io/library/busybox:1.36",
			"--workers", "1",
			"sh -c 'echo hello-world; sleep 300'",
		)
		submitCmd.Stdout = &out
		submitCmd.Stderr = &out
		err := submitCmd.Run()
		Expect(err).NotTo(HaveOccurred(), "submit output: %s", out.String())
		out.Reset()

		By("Waiting for pod to be ready")
		waitCmd := exec.Command("kubectl", "wait", "--for=condition=PodReady",
			"pod", "-l", "training.kubeflow.org/job-name="+jobName,
			"-n", namespace, "--timeout=120s")
		waitCmd.Stdout = &out
		waitCmd.Stderr = &out
		_ = waitCmd.Run()
		out.Reset()

		By("Fetching logs")
		logsCmd := exec.Command(arenaV2Bin, "job", "logs", jobName,
			"--namespace", namespace)
		logsCmd.Stdout = &out
		logsCmd.Stderr = &out
		err = logsCmd.Run()
		Expect(err).NotTo(HaveOccurred(), "logs output: %s", out.String())
		Expect(out.String()).To(ContainSubstring("hello-world"))
	})

	It("should support --tail flag", func() {
		var out bytes.Buffer

		By("Submitting a PyTorch job that prints many lines")
		submitCmd := exec.Command(arenaV2Bin, "job", "submit", "pytorch",
			"--name", jobName,
			"--namespace", namespace,
			"--image", "docker.io/library/busybox:1.36",
			"--workers", "1",
			"sh -c 'for i in $(seq 1 20); do echo line-$i; done; sleep 300'",
		)
		submitCmd.Stdout = &out
		submitCmd.Stderr = &out
		err := submitCmd.Run()
		Expect(err).NotTo(HaveOccurred(), "submit output: %s", out.String())
		out.Reset()

		By("Waiting for pod to be ready")
		waitCmd := exec.Command("kubectl", "wait", "--for=condition=PodReady",
			"pod", "-l", "training.kubeflow.org/job-name="+jobName,
			"-n", namespace, "--timeout=120s")
		waitCmd.Stdout = &out
		waitCmd.Stderr = &out
		_ = waitCmd.Run()
		out.Reset()

		By("Fetching logs with --tail 5")
		logsCmd := exec.Command(arenaV2Bin, "job", "logs", jobName,
			"--namespace", namespace, "--tail", "5")
		logsCmd.Stdout = &out
		logsCmd.Stderr = &out
		err = logsCmd.Run()
		Expect(err).NotTo(HaveOccurred(), "logs output: %s", out.String())

		lines := strings.Split(strings.TrimSpace(out.String()), "\n")
		Expect(len(lines)).To(BeNumerically("<=", 5),
			"--tail 5 should return at most 5 lines, got %d: %s", len(lines), out.String())
	})
})
