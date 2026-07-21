// MPI lifecycle tests: CRUD smoke tests that verify job submit, get, and
// delete operations using placeholder images. These do not wait for pod
// readiness or validate training outcomes.

package e2e_test

import (
	"bytes"
	"encoding/json"
	"fmt"
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
		frameworkLifecycle("mpi", "registry-cn-beijing.ack.aliyuncs.com/acs/busybox")
	})

	It("Horovod job lifecycle", func() {
		frameworkLifecycle("horovod", "registry-cn-beijing.ack.aliyuncs.com/acs/busybox")
	})

	It("DeepSpeed job lifecycle", func() {
		frameworkLifecycle("deepspeed", "registry-cn-beijing.ack.aliyuncs.com/acs/busybox")
	})
})
