package e2e_test

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Phase 2 features", Pending, func() {
	var (
		jobName   string
		namespace string
	)

	BeforeEach(func() {
		jobName = "test-features"
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

	It("should apply scheduling, affinity, init containers, and git sync", func() {
		By("Submitting a job with advanced features")
		var out bytes.Buffer
		runCmd := exec.Command(arenaV2Bin, "job", "run", "-f",
			filepath.Join("..", "testdata", "features.yaml"),
			"--namespace", namespace)
		runCmd.Stdout = &out
		runCmd.Stderr = &out
		Expect(runCmd.Run()).NotTo(HaveOccurred(),
			"job run failed: %s", out.String())

		By("Fetching the PyTorchJob CRD")
		out.Reset()
		getCmd := exec.Command("kubectl", "get", "pytorchjob", jobName,
			"-n", namespace, "-o", "json")
		getCmd.Stdout = &out
		getCmd.Stderr = &out
		Expect(getCmd.Run()).NotTo(HaveOccurred(),
			"kubectl get pytorchjob failed: %s", out.String())

		var crd map[string]interface{}
		Expect(json.Unmarshal(out.Bytes(), &crd)).NotTo(HaveOccurred())

		// When features are implemented, verify:
		// - Gang scheduling annotations in CRD metadata
		// - Queue annotation "high-priority" in CRD metadata
		// - priorityClassName "premium" in pod spec
		// - Affinity rules matching gpu-type=A100 in pod spec
		// - Init container "setup-logs" with busybox image
		// - Git sync container for training-operator repo
		// - NCCL_DEBUG=INFO environment variable
	})
})
