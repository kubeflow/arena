package e2e_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("--set overrides (advanced)", func() {
	var (
		jobName   string
		namespace string
		yamlPath  string
	)

	BeforeEach(func() {
		jobName = fmt.Sprintf("v2-setadv-%d", GinkgoRandomSeed())
		namespace = "default"
	})

	AfterEach(func() {
		if yamlPath != "" {
			os.Remove(yamlPath)
		}
	})

	baseYAML := func() string {
		return fmt.Sprintf(`version: 0.1.0
name: %s
image: docker.io/library/pytorch:2.1
framework:
  name: pytorch
worker:
  replicas: 2
run: python train.py
`, jobName)
	}

	runSetDryRun := func(setExprs ...string) map[string]interface{} {
		var err error
		yamlPath, err = createTempYAML(baseYAML())
		Expect(err).NotTo(HaveOccurred())

		args := []string{"job", "run", "-f", yamlPath, "--namespace", namespace, "--dry-run"}
		for _, expr := range setExprs {
			args = append(args, "--set", expr)
		}

		var stdout, stderr bytes.Buffer
		cmd := exec.Command(arenaV2Bin, args...)
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err = cmd.Run()
		Expect(err).NotTo(HaveOccurred(), "run --set output: %s", stderr.String())

		var parsed map[string]interface{}
		Expect(json.Unmarshal(stdout.Bytes(), &parsed)).NotTo(HaveOccurred(),
			"dry-run output should be valid JSON: %s", stdout.String())
		return parsed
	}

	It("should override lifecycle.active_deadline", func() {
		crd := runSetDryRun("lifecycle.active_deadline=7d")

		spec := crd["spec"].(map[string]interface{})
		runPolicy, ok := spec["runPolicy"].(map[string]interface{})
		Expect(ok).To(BeTrue(), "spec.runPolicy should exist")
		Expect(runPolicy["activeDeadlineSeconds"]).To(Equal(int64(604800)))
	})

	It("should override scheduling.gang.enabled", func() {
		crd := runSetDryRun("scheduling.gang.enabled=true")

		spec := crd["spec"].(map[string]interface{})
		runPolicy, ok := spec["runPolicy"].(map[string]interface{})
		Expect(ok).To(BeTrue(), "spec.runPolicy should exist")
		schedulingPolicy, ok := runPolicy["schedulingPolicy"].(map[string]interface{})
		Expect(ok).To(BeTrue(), "runPolicy.schedulingPolicy should exist")
		Expect(schedulingPolicy["minAvailable"]).NotTo(BeNil())
	})

	It("should override labels", func() {
		crd := runSetDryRun("labels.team=ml")

		metadata := crd["metadata"].(map[string]interface{})
		labels := metadata["labels"].(map[string]interface{})
		Expect(labels["team"]).To(Equal("ml"))
	})

	It("should override envs", func() {
		crd := runSetDryRun("envs.NCCL_DEBUG=INFO")

		spec := crd["spec"].(map[string]interface{})
		replicaSpecs := spec["pytorchReplicaSpecs"].(map[string]interface{})
		worker := replicaSpecs["Worker"].(map[string]interface{})
		template := worker["template"].(map[string]interface{})
		podSpec := template["spec"].(map[string]interface{})
		containers := podSpec["containers"].([]interface{})
		container := containers[0].(map[string]interface{})

		envs := container["env"].([]interface{})
		found := false
		for _, e := range envs {
			env := e.(map[string]interface{})
			if env["name"] == "NCCL_DEBUG" {
				Expect(env["value"]).To(Equal("INFO"))
				found = true
			}
		}
		Expect(found).To(BeTrue(), "env NCCL_DEBUG=INFO not found")
	})
})
