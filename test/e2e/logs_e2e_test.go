//go:build v2e2e

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
		submitCmd := exec.Command(arenaV2Bin, "submit", "pytorch",
			"--name", jobName,
			"--namespace", namespace,
			"--image", busyboxImage(),
			"--workers", "1",
			"sh -c 'echo hello-world; sleep 120'",
		)
		submitCmd.Stdout = &out
		submitCmd.Stderr = &out
		err := submitCmd.Run()
		Expect(err).NotTo(HaveOccurred(), "submit output: %s", out.String())
		out.Reset()

		By("Waiting for pod to be created")
		createCmd := exec.Command("kubectl", "wait", "--for=create",
			"pod", "-l", "training.kubeflow.org/job-name="+jobName,
			"-n", namespace, "--timeout=120s")
		createCmd.Stdout = &out
		createCmd.Stderr = &out
		err = createCmd.Run()
		Expect(err).NotTo(HaveOccurred(), "pod was not created: %s", out.String())
		out.Reset()

		By("Waiting for pod to be ready")
		readyCmd := exec.Command("kubectl", "wait", "--for=condition=Ready",
			"pod", "-l", "training.kubeflow.org/job-name="+jobName,
			"-n", namespace, "--timeout=120s")
		readyCmd.Stdout = &out
		readyCmd.Stderr = &out
		err = readyCmd.Run()
		Expect(err).NotTo(HaveOccurred(), "pod did not become ready: %s", out.String())
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
		submitCmd := exec.Command(arenaV2Bin, "submit", "pytorch",
			"--name", jobName,
			"--namespace", namespace,
			"--image", busyboxImage(),
			"--workers", "1",
			"sh -c 'for i in $(seq 1 20); do echo line-$i; done; sleep 120'",
		)
		submitCmd.Stdout = &out
		submitCmd.Stderr = &out
		err := submitCmd.Run()
		Expect(err).NotTo(HaveOccurred(), "submit output: %s", out.String())
		out.Reset()

		By("Waiting for pod to be created")
		createCmd := exec.Command("kubectl", "wait", "--for=create",
			"pod", "-l", "training.kubeflow.org/job-name="+jobName,
			"-n", namespace, "--timeout=120s")
		createCmd.Stdout = &out
		createCmd.Stderr = &out
		err = createCmd.Run()
		Expect(err).NotTo(HaveOccurred(), "pod was not created: %s", out.String())
		out.Reset()

		By("Waiting for pod to be ready")
		readyCmd := exec.Command("kubectl", "wait", "--for=condition=Ready",
			"pod", "-l", "training.kubeflow.org/job-name="+jobName,
			"-n", namespace, "--timeout=120s")
		readyCmd.Stdout = &out
		readyCmd.Stderr = &out
		err = readyCmd.Run()
		Expect(err).NotTo(HaveOccurred(), "pod did not become ready: %s", out.String())
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
