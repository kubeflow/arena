//go:build v2e2e

package e2e_test

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// TODO: enable these tests once a batch scheduler (Volcano/Kueue) is
// installed in the e2e Kind cluster.
var _ = Describe("Gang Scheduling", Pending, func() {
	var (
		jobName   string
		namespace string
	)

	BeforeEach(func() {
		jobName = "test-gang"
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

	It("should set minAvailable in schedulingPolicy when gang is enabled", func() {
		gangJobYAML := `version: 0.1.0
name: test-gang
framework:
  name: pytorch
image: docker.io/pytorch/pytorch:2.1
run: python train.py
worker:
  replicas: 2
  resources:
    cpu: 1
    memory: 1Gi
scheduling:
  gang:
    enabled: true
`

		By("Submitting a job with gang scheduling enabled")
		yamlPath, err := createTempYAML(gangJobYAML)
		Expect(err).NotTo(HaveOccurred())
		defer os.Remove(yamlPath)

		var out bytes.Buffer
		runCmd := exec.Command(arenaV2Bin, "job", "run", "-f",
			yamlPath, "--namespace", namespace)
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

		By("Verifying minAvailable in schedulingPolicy")
		spec := crd["spec"].(map[string]interface{})
		runPolicy := spec["runPolicy"].(map[string]interface{})
		schedulingPolicy := runPolicy["schedulingPolicy"].(map[string]interface{})
		Expect(schedulingPolicy["minAvailable"]).To(Equal(float64(2)))
	})

	It("should set queue in schedulingPolicy", func() {
		queueJobYAML := `version: 0.1.0
name: test-gang
framework:
  name: pytorch
image: docker.io/pytorch/pytorch:2.1
run: python train.py
worker:
  replicas: 2
  resources:
    cpu: 1
    memory: 1Gi
scheduling:
  gang:
    enabled: true
  queue: high-priority
`

		By("Submitting a job with queue")
		yamlPath, err := createTempYAML(queueJobYAML)
		Expect(err).NotTo(HaveOccurred())
		defer os.Remove(yamlPath)

		var out bytes.Buffer
		runCmd := exec.Command(arenaV2Bin, "job", "run", "-f",
			yamlPath, "--namespace", namespace)
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

		By("Verifying queue in schedulingPolicy")
		spec := crd["spec"].(map[string]interface{})
		runPolicy := spec["runPolicy"].(map[string]interface{})
		schedulingPolicy := runPolicy["schedulingPolicy"].(map[string]interface{})
		Expect(schedulingPolicy["queue"]).To(Equal("high-priority"))
	})
})
