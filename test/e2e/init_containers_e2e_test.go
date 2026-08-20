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

const initContainerJobYAML = `version: 0.1.0
name: test-init
framework:
  name: pytorch
image: docker.io/pytorch/pytorch:2.1
run: python train.py
worker:
  replicas: 1
  resources:
    cpu: 1
    memory: 1Gi
init:
  - name: setup-logs
    image: docker.io/library/busybox:1.35
    run: mkdir -p /logs
`

const gitSyncJobYAML = `version: 0.1.0
name: test-gitsync
framework:
  name: pytorch
image: docker.io/pytorch/pytorch:2.1
run: python train.py
worker:
  replicas: 1
  resources:
    cpu: 1
    memory: 1Gi
sync:
  - git: https://github.com/kubeflow/training-operator.git
    local_path: /workspace/training-operator
`

var _ = Describe("Init Containers", func() {
	var (
		jobName   string
		namespace string
	)

	BeforeEach(func() {
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

	It("should create init container with name, image, and command", func() {
		jobName = "test-init"

		By("Submitting a job with a user-defined init container")
		yamlPath, err := createTempYAML(initContainerJobYAML)
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

		By("Verifying init container in pod spec")
		spec := crd["spec"].(map[string]interface{})
		replicaSpecs := spec["pytorchReplicaSpecs"].(map[string]interface{})
		master := replicaSpecs["Master"].(map[string]interface{})
		template := master["template"].(map[string]interface{})
		podSpec := template["spec"].(map[string]interface{})
		initContainers := podSpec["initContainers"].([]interface{})
		Expect(initContainers).To(HaveLen(1))

		initContainer := initContainers[0].(map[string]interface{})
		Expect(initContainer["name"]).To(Equal("setup-logs"))
		Expect(initContainer["image"]).To(Equal("docker.io/library/busybox:1.35"))

		command := initContainer["command"].([]interface{})
		Expect(command).To(HaveLen(2))
		Expect(command[0]).To(Equal("/bin/sh"))
		Expect(command[1]).To(Equal("-c"))

		args := initContainer["args"].([]interface{})
		Expect(args).To(HaveLen(1))
		Expect(args[0]).To(Equal("mkdir -p /logs"))
	})

	It("should create git sync init container with GIT_SYNC_REPO env", func() {
		jobName = "test-gitsync"

		By("Submitting a job with git sync")
		yamlPath, err := createTempYAML(gitSyncJobYAML)
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

		By("Verifying git sync init container in pod spec")
		spec := crd["spec"].(map[string]interface{})
		replicaSpecs := spec["pytorchReplicaSpecs"].(map[string]interface{})
		master := replicaSpecs["Master"].(map[string]interface{})
		template := master["template"].(map[string]interface{})
		podSpec := template["spec"].(map[string]interface{})
		initContainers := podSpec["initContainers"].([]interface{})
		Expect(initContainers).To(HaveLen(1))

		syncContainer := initContainers[0].(map[string]interface{})
		Expect(syncContainer["name"]).To(Equal("arena-sync-0"))

		envs := syncContainer["env"].([]interface{})
		var foundRepo bool
		for _, e := range envs {
			envVar := e.(map[string]interface{})
			if envVar["name"] == "GIT_SYNC_REPO" {
				Expect(envVar["value"]).To(Equal("https://github.com/kubeflow/training-operator.git"))
				foundRepo = true
				break
			}
		}
		Expect(foundRepo).To(BeTrue(), "GIT_SYNC_REPO env var not found in git sync container")
	})
})
