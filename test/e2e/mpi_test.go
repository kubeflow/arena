//go:build v2e2e

// MPI lifecycle tests: CRUD smoke tests that verify job submit, get, and
// delete operations using placeholder images. These do not wait for pod
// readiness or validate training outcomes.

package e2e_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	outputpkg "github.com/kubeflow/arena/pkg/output"
)

var _ = Describe("MPI-based Jobs", func() {
	var (
		namespace string
		jobName   string
	)

	BeforeEach(func() {
		namespace = "default"
		jobName = ""
	})

	AfterEach(func() {
		if jobName == "" {
			return
		}
		var out bytes.Buffer
		delCmd := exec.Command(arenaV2Bin, "job", "delete", jobName,
			"--namespace", namespace)
		delCmd.Stdout = &out
		delCmd.Stderr = &out
		_ = delCmd.Run()
		jobName = ""
	})

	frameworkLifecycle := func(framework, image string) {
		jobName = fmt.Sprintf("v2-%s-%d", framework, GinkgoRandomSeed())

		var out bytes.Buffer
		var err error

		By(fmt.Sprintf("Validating dry-run CRD structure for %s", framework))
		var dryStdout bytes.Buffer
		dryCmd := exec.Command(arenaV2Bin, "submit", framework,
			"--name", jobName+"-dry",
			"--namespace", namespace,
			"--image", image,
			"--workers", "2",
			"--dry-run",
			"sh -c 'echo hello-world; sleep 120'",
		)
		dryCmd.Stdout = &dryStdout
		dryCmd.Stderr = &out
		err = dryCmd.Run()
		Expect(err).NotTo(HaveOccurred(), "dry-run failed: %s", out.String())

		var crd map[string]interface{}
		err = json.Unmarshal(dryStdout.Bytes(), &crd)
		Expect(err).NotTo(HaveOccurred(), "dry-run output should be valid JSON")
		Expect(crd["kind"]).To(Equal("MPIJob"), "CRD kind should be MPIJob")
		Expect(crd["apiVersion"]).To(Equal(mpiJobStorageVersion()),
			"CRD apiVersion should match cluster storage version")

		// Validate metadata structure
		metadata, ok := crd["metadata"].(map[string]interface{})
		Expect(ok).To(BeTrue(), "CRD should have metadata")
		Expect(metadata["name"]).To(Equal(jobName + "-dry"))
		Expect(metadata["namespace"]).To(Equal(namespace))

		// Validate framework label
		labels, ok := metadata["labels"].(map[string]interface{})
		Expect(ok).To(BeTrue(), "CRD metadata should have labels")
		Expect(labels["arena.io/framework"]).To(Equal(framework))

		// Validate spec structure has replica specs
		spec, ok := crd["spec"].(map[string]interface{})
		Expect(ok).To(BeTrue(), "CRD should have spec")
		mpiReplicaSpecs, ok := spec["mpiReplicaSpecs"].(map[string]interface{})
		Expect(ok).To(BeTrue(), "CRD spec should have mpiReplicaSpecs")
		worker, ok := mpiReplicaSpecs["Worker"].(map[string]interface{})
		Expect(ok).To(BeTrue(), "mpiReplicaSpecs should have Worker")

		// Validate worker has resource requests
		template, ok := worker["template"].(map[string]interface{})
		Expect(ok).To(BeTrue(), "Worker should have template")
		podSpec, ok := template["spec"].(map[string]interface{})
		Expect(ok).To(BeTrue(), "Worker template should have spec")
		containers, ok := podSpec["containers"].([]interface{})
		Expect(ok).To(BeTrue(), "Worker pod spec should have containers")
		Expect(containers).NotTo(BeEmpty(), "Worker should have at least one container")

		By(fmt.Sprintf("Submitting a %s job", framework))
		submitCmd := exec.Command(arenaV2Bin, "submit", framework,
			"--name", jobName,
			"--namespace", namespace,
			"--image", image,
			"--workers", "2",
			"sh -c 'echo hello-world; sleep 120'",
		)
		submitCmd.Stdout = &out
		submitCmd.Stderr = &out
		err = submitCmd.Run()
		Expect(err).NotTo(HaveOccurred(), "submit output: %s", out.String())
		out.Reset()

		By(fmt.Sprintf("Verifying CRD structure for %s via get", framework))
		getCmd := exec.Command(arenaV2Bin, "job", "get", jobName,
			"--namespace", namespace, "-o", string(outputpkg.FormatJSON))
		getCmd.Stdout = &out
		getCmd.Stderr = &out
		err = getCmd.Run()
		Expect(err).NotTo(HaveOccurred(), "get output: %s", out.String())

		// Validate the retrieved CRD has expected structure
		Expect(out.String()).To(ContainSubstring(framework))
		// Expect(out.String()).To(ContainSubstring("MPIJob"))
	}

	It("MPI job lifecycle", func() {
		frameworkLifecycle("mpi", busyboxImage())
	})

	It("Horovod job lifecycle", func() {
		frameworkLifecycle("horovod", busyboxImage())
	})

	It("DeepSpeed job lifecycle", func() {
		frameworkLifecycle("deepspeed", busyboxImage())
	})
})

const mpiLauncherPVCYAML = `apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: mpi-e2e-pvc
  namespace: default
spec:
  accessModes: ["ReadWriteOnce"]
  resources:
    requests:
      storage: 1Gi
`

// mounts_on_launcher=false: all volumes stay declared on the launcher pod
// (init containers may reference them) but the launcher main container
// mounts nothing.
var _ = Describe("MPI launcher volume policy", func() {
	var namespace string

	BeforeEach(func() {
		namespace = "default"
		var out bytes.Buffer
		cmd := exec.Command("kubectl", "apply", "-f", "-")
		cmd.Stdin = bytes.NewReader([]byte(mpiLauncherPVCYAML))
		cmd.Stdout = &out
		cmd.Stderr = &out
		Expect(cmd.Run()).NotTo(HaveOccurred(),
			"kubectl apply pvc failed: %s", out.String())
	})

	AfterEach(func() {
		var out bytes.Buffer
		delCmd := exec.Command(arenaV2Bin, "job", "delete", "test-mpi-mounts",
			"--namespace", namespace)
		delCmd.Stdout = &out
		delCmd.Stderr = &out
		_ = delCmd.Run()

		out.Reset()
		pvcCmd := exec.Command("kubectl", "delete", "pvc", "mpi-e2e-pvc",
			"-n", namespace, "--ignore-not-found")
		pvcCmd.Stdout = &out
		pvcCmd.Stderr = &out
		_ = pvcCmd.Run()
	})

	names := func(field string, holder map[string]interface{}) []string {
		items, ok := holder[field].([]interface{})
		if !ok {
			return nil
		}
		result := make([]string, 0, len(items))
		for _, item := range items {
			if m, ok := item.(map[string]interface{}); ok {
				if name, ok := m["name"].(string); ok {
					result = append(result, name)
				}
			}
		}
		return result
	}

	It("should keep volumes on launcher but clear main container mounts when mounts_on_launcher=false", func() {
		jobYAML := fmt.Sprintf(`version: 0.1.0
name: test-mpi-mounts
framework:
  name: mpi
  options:
    mounts_on_launcher: false
image: %s
run: sleep 3600
worker:
  replicas: 1
  resources:
    cpu: 1
    memory: 1Gi
storages:
  - name: dataset
    pvc: mpi-e2e-pvc
    mount_path: /data
  - name: cache
    tmp: 1Gi
    mount_path: /cache
init:
  - name: setup
    image: %s
    run: echo hi
    mounts:
      - name: dataset
`, busyboxImage(), busyboxImage())

		By("Submitting an MPI job with PVC + tmp storages")
		yamlPath, err := createTempYAML(jobYAML)
		Expect(err).NotTo(HaveOccurred())
		defer os.Remove(yamlPath)

		var out bytes.Buffer
		runCmd := exec.Command(arenaV2Bin, "job", "run", "-f",
			yamlPath, "--namespace", namespace)
		runCmd.Stdout = &out
		runCmd.Stderr = &out
		Expect(runCmd.Run()).NotTo(HaveOccurred(),
			"job run failed: %s", out.String())

		By("Fetching the MPIJob CRD")
		out.Reset()
		getCmd := exec.Command("kubectl", "get", "mpijob", "test-mpi-mounts",
			"-n", namespace, "-o", "json")
		getCmd.Stdout = &out
		getCmd.Stderr = &out
		Expect(getCmd.Run()).NotTo(HaveOccurred(),
			"kubectl get mpijob failed: %s", out.String())

		var crd map[string]interface{}
		Expect(json.Unmarshal(out.Bytes(), &crd)).NotTo(HaveOccurred())

		spec := crd["spec"].(map[string]interface{})
		replicaSpecs := spec["mpiReplicaSpecs"].(map[string]interface{})

		By("Verifying launcher keeps all volumes but the main container mounts nothing")
		launcher := replicaSpecs["Launcher"].(map[string]interface{})
		launcherPod := launcher["template"].(map[string]interface{})["spec"].(map[string]interface{})
		Expect(names("volumes", launcherPod)).To(ConsistOf("dataset", "cache"),
			"launcher should keep all volumes declared (init containers may reference them)")

		launcherContainers := launcherPod["containers"].([]interface{})
		launcherMain := launcherContainers[0].(map[string]interface{})
		Expect(names("volumeMounts", launcherMain)).To(BeEmpty(),
			"launcher main container should have no volumeMounts when mounts_on_launcher=false")

		launcherInits := launcherPod["initContainers"].([]interface{})
		launcherInit := launcherInits[0].(map[string]interface{})
		Expect(names("volumeMounts", launcherInit)).To(ConsistOf("dataset"),
			"init container mounts must never be rewritten")

		By("Verifying worker keeps all volumes")
		worker := replicaSpecs["Worker"].(map[string]interface{})
		workerPod := worker["template"].(map[string]interface{})["spec"].(map[string]interface{})
		Expect(names("volumes", workerPod)).To(ConsistOf("dataset", "cache"),
			"worker should keep both volumes")
	})
})
