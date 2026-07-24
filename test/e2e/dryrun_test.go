package e2e_test

import (
	"bytes"
	"encoding/json"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Dry-Run", func() {
	It("should generate valid PyTorchJob CRD without submitting", func() {
		var stdout, stderr bytes.Buffer
		cmd := exec.Command(arenaV2Bin, "submit", "pytorch",
			"--name", "dryrun-pytorch",
			"--image", "docker.io/library/pytorch:2.1",
			"--workers", "2",
			"--dry-run",
			"python train.py",
		)
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		Expect(err).NotTo(HaveOccurred(), "dry-run failed: %s", stderr.String())

		var parsed map[string]interface{}
		err = json.Unmarshal(stdout.Bytes(), &parsed)
		Expect(err).NotTo(HaveOccurred(), "dry-run output should be valid JSON")
		Expect(parsed["kind"]).To(Equal("PyTorchJob"))

		metadata := parsed["metadata"].(map[string]interface{})
		labels := metadata["labels"].(map[string]interface{})
		Expect(labels["arena.io/framework"]).To(Equal("pytorch"))
	})

	It("should generate valid MPIJob CRD for horovod", func() {
		var stdout, stderr bytes.Buffer
		cmd := exec.Command(arenaV2Bin, "submit", "horovod",
			"--name", "dryrun-horovod",
			"--image", "docker.io/library/horovod:latest",
			"--workers", "2",
			"--dry-run",
			"mpirun train",
		)
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		Expect(err).NotTo(HaveOccurred(), "dry-run failed: %s", stderr.String())

		var parsed map[string]interface{}
		err = json.Unmarshal(stdout.Bytes(), &parsed)
		Expect(err).NotTo(HaveOccurred())
		Expect(parsed["kind"]).To(Equal("MPIJob"))

		metadata := parsed["metadata"].(map[string]interface{})
		labels := metadata["labels"].(map[string]interface{})
		Expect(labels["arena.io/framework"]).To(Equal("horovod"))
	})

	It("should generate MPIJob CR with the cluster's storage apiVersion in dry-run", func() {
		var stdout, stderr bytes.Buffer
		cmd := exec.Command(arenaV2Bin, "submit", "mpi",
			"--name", "dryrun-mpi-version",
			"--image", "docker.io/library/openmpi:4.1",
			"--workers", "2",
			"--dry-run",
			"mpirun train",
		)
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		Expect(err).NotTo(HaveOccurred(), "dry-run failed: %s", stderr.String())

		var parsed map[string]interface{}
		err = json.Unmarshal(stdout.Bytes(), &parsed)
		Expect(err).NotTo(HaveOccurred(), "dry-run output should be valid JSON")
		Expect(parsed["kind"]).To(Equal("MPIJob"))
		Expect(parsed["apiVersion"]).To(Equal(mpiJobStorageVersion()))
	})

	It("should generate PyTorchJob CR with apiVersion kubeflow.org/v1 in dry-run", func() {
		var stdout, stderr bytes.Buffer
		cmd := exec.Command(arenaV2Bin, "submit", "pytorch",
			"--name", "dryrun-pytorch-version",
			"--image", "docker.io/library/pytorch:2.1",
			"--workers", "2",
			"--dry-run",
			"python train.py",
		)
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		Expect(err).NotTo(HaveOccurred(), "dry-run failed: %s", stderr.String())

		var parsed map[string]interface{}
		err = json.Unmarshal(stdout.Bytes(), &parsed)
		Expect(err).NotTo(HaveOccurred(), "dry-run output should be valid JSON")
		Expect(parsed["kind"]).To(Equal("PyTorchJob"))
		Expect(parsed["apiVersion"]).To(Equal("kubeflow.org/v1"))
	})

	It("should generate valid MPIJob CRD for deepspeed", func() {
		var stdout, stderr bytes.Buffer
		cmd := exec.Command(arenaV2Bin, "submit", "deepspeed",
			"--name", "dryrun-deepspeed",
			"--image", "docker.io/library/deepspeed:latest",
			"--workers", "2",
			"--dry-run",
			"deepspeed train.py",
		)
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		Expect(err).NotTo(HaveOccurred(), "dry-run failed: %s", stderr.String())

		var parsed map[string]interface{}
		err = json.Unmarshal(stdout.Bytes(), &parsed)
		Expect(err).NotTo(HaveOccurred())
		Expect(parsed["kind"]).To(Equal("MPIJob"))

		metadata := parsed["metadata"].(map[string]interface{})
		labels := metadata["labels"].(map[string]interface{})
		Expect(labels["arena.io/framework"]).To(Equal("deepspeed"))
	})

	It("should route --gpus to Master for single-node PyTorch", func() {
		var stdout, stderr bytes.Buffer
		cmd := exec.Command(arenaV2Bin, "submit", "pytorch",
			"--name", "dryrun-pytorch-single-gpu",
			"--image", "docker.io/library/pytorch:2.1",
			"--gpus", "1",
			"--dry-run",
			"python train.py",
		)
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		Expect(err).NotTo(HaveOccurred(), "dry-run failed: %s", stderr.String())

		var parsed map[string]interface{}
		err = json.Unmarshal(stdout.Bytes(), &parsed)
		Expect(err).NotTo(HaveOccurred(), "dry-run output should be valid JSON")
		Expect(parsed["kind"]).To(Equal("PyTorchJob"))

		spec := parsed["spec"].(map[string]interface{})
		replicaSpecs := spec["pytorchReplicaSpecs"].(map[string]interface{})

		// Master should exist with GPU resource
		master, ok := replicaSpecs["Master"]
		Expect(ok).To(BeTrue(), "Master replica should exist")
		masterSpec := master.(map[string]interface{})
		masterTemplate := masterSpec["template"].(map[string]interface{})
		masterPodSpec := masterTemplate["spec"].(map[string]interface{})
		masterContainers := masterPodSpec["containers"].([]interface{})
		masterContainer := masterContainers[0].(map[string]interface{})
		masterRes := masterContainer["resources"].(map[string]interface{})
		masterReqs := masterRes["requests"].(map[string]interface{})
		Expect(masterReqs["nvidia.com/gpu"]).To(Equal("1"), "Master should have 1 GPU")

		// Worker should not exist
		_, hasWorker := replicaSpecs["Worker"]
		Expect(hasWorker).To(BeFalse(), "Worker should not exist in single-node mode")
	})
})
