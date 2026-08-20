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

const envJobYAML = `version: 0.1.0
name: test-env
framework:
  name: pytorch
image: docker.io/pytorch/pytorch:2.1
run: echo $NCCL_DEBUG
envs:
  NCCL_DEBUG: "INFO"
worker:
  replicas: 1
  resources:
    cpu: 1
    memory: 1Gi
`

var _ = Describe("Environment Variables", func() {
	var (
		jobName   string
		namespace string
	)

	BeforeEach(func() {
		jobName = "test-env"
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

	It("should set NCCL_DEBUG env var in container", func() {
		By("Submitting a job with env vars")
		yamlPath, err := createTempYAML(envJobYAML)
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

		By("Verifying NCCL_DEBUG env var in container")
		spec := crd["spec"].(map[string]interface{})
		replicaSpecs := spec["pytorchReplicaSpecs"].(map[string]interface{})
		master := replicaSpecs["Master"].(map[string]interface{})
		template := master["template"].(map[string]interface{})
		podSpec := template["spec"].(map[string]interface{})
		containers := podSpec["containers"].([]interface{})
		container := containers[0].(map[string]interface{})
		envs := container["env"].([]interface{})

		var found bool
		for _, e := range envs {
			envVar := e.(map[string]interface{})
			if envVar["name"] == "NCCL_DEBUG" {
				Expect(envVar["value"]).To(Equal("INFO"))
				found = true
				break
			}
		}
		Expect(found).To(BeTrue(), "NCCL_DEBUG env var not found in container")
	})
})
