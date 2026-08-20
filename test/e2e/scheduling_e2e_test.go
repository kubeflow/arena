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

const priorityJobYAML = `version: 0.1.0
name: test-priority
framework:
  name: pytorch
image: docker.io/pytorch/pytorch:2.1
run: python train.py
worker:
  replicas: 1
  resources:
    cpu: 1
    memory: 1Gi
scheduling:
  priority_class_name: premium
`

const affinityJobYAML = `version: 0.1.0
name: test-affinity
framework:
  name: pytorch
image: docker.io/pytorch/pytorch:2.1
run: python train.py
worker:
  replicas: 1
  resources:
    cpu: 1
    memory: 1Gi
scheduling:
  affinity:
    policy: spread
    constraint: required
    target: node
    rules:
      - match_labels:
          gpu-type: A100
`

var _ = Describe("Scheduling", func() {
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

	It("should set priorityClassName in pod spec and schedulingPolicy", func() {
		jobName = "test-priority"

		By("Submitting a job with priorityClassName")
		yamlPath, err := createTempYAML(priorityJobYAML)
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

		By("Verifying priorityClassName in pod spec")
		spec := crd["spec"].(map[string]interface{})
		replicaSpecs := spec["pytorchReplicaSpecs"].(map[string]interface{})
		master := replicaSpecs["Master"].(map[string]interface{})
		template := master["template"].(map[string]interface{})
		podSpec := template["spec"].(map[string]interface{})
		Expect(podSpec["priorityClassName"]).To(Equal("premium"))

		By("Verifying priorityClass in schedulingPolicy")
		runPolicy := spec["runPolicy"].(map[string]interface{})
		schedulingPolicy := runPolicy["schedulingPolicy"].(map[string]interface{})
		Expect(schedulingPolicy["priorityClass"]).To(Equal("premium"))
	})

	It("should set nodeAffinity from affinity rules", func() {
		jobName = "test-affinity"

		By("Submitting a job with node affinity rules")
		yamlPath, err := createTempYAML(affinityJobYAML)
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

		By("Verifying nodeAffinity in pod spec")
		spec := crd["spec"].(map[string]interface{})
		replicaSpecs := spec["pytorchReplicaSpecs"].(map[string]interface{})
		master := replicaSpecs["Master"].(map[string]interface{})
		template := master["template"].(map[string]interface{})
		podSpec := template["spec"].(map[string]interface{})
		affinity := podSpec["affinity"].(map[string]interface{})
		nodeAffinity := affinity["nodeAffinity"].(map[string]interface{})
		required := nodeAffinity["requiredDuringSchedulingIgnoredDuringExecution"].(map[string]interface{})
		nodeSelectorTerms := required["nodeSelectorTerms"].([]interface{})
		Expect(nodeSelectorTerms).To(HaveLen(1))

		term := nodeSelectorTerms[0].(map[string]interface{})
		matchExpressions := term["matchExpressions"].([]interface{})
		Expect(matchExpressions).To(HaveLen(1))

		expr := matchExpressions[0].(map[string]interface{})
		Expect(expr["key"]).To(Equal("gpu-type"))
		Expect(expr["operator"]).To(Equal("In"))
		Expect(expr["values"]).To(Equal([]interface{}{"A100"}))
	})
})
