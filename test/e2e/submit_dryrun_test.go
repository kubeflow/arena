package e2e_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func runSubmitDryRun(args ...string) map[string]interface{} {
	cmdArgs := append([]string{"job", "submit"}, args...)
	cmdArgs = append(cmdArgs, "--dry-run")

	var stdout, stderr bytes.Buffer
	cmd := exec.Command(arenaV2Bin, cmdArgs...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	Expect(err).NotTo(HaveOccurred(), "dry-run failed: %s", stderr.String())

	var parsed map[string]interface{}
	Expect(json.Unmarshal(stdout.Bytes(), &parsed)).NotTo(HaveOccurred(),
		"dry-run output should be valid JSON: %s", stdout.String())
	return parsed
}

func getReplicaSpecs(crd map[string]interface{}, key string) map[string]interface{} {
	spec := crd["spec"].(map[string]interface{})
	replicaSpecs := spec[key].(map[string]interface{})
	return replicaSpecs
}

func getFirstContainer(replicaSpec map[string]interface{}) map[string]interface{} {
	template := replicaSpec["template"].(map[string]interface{})
	podSpec := template["spec"].(map[string]interface{})
	containers := podSpec["containers"].([]interface{})
	return containers[0].(map[string]interface{})
}

var _ = Describe("Submit flags (dry-run)", func() {
	var (
		jobName   string
		namespace string
	)

	BeforeEach(func() {
		jobName = fmt.Sprintf("v2-dryrun-%d", GinkgoRandomSeed())
		namespace = "default"
	})

	It("PyTorch --workers 1 produces master-only CRD", func() {
		crd := runSubmitDryRun("pytorch",
			"--name", jobName,
			"--namespace", namespace,
			"--image", "docker.io/library/pytorch:2.1",
			"--workers", "1",
			"python train.py")

		replicaSpecs := getReplicaSpecs(crd, "pytorchReplicaSpecs")
		Expect(replicaSpecs).To(HaveKey("Master"))
		Expect(replicaSpecs).NotTo(HaveKey("Worker"))
	})

	It("PyTorch --workers 4 produces 1 master + 3 workers", func() {
		crd := runSubmitDryRun("pytorch",
			"--name", jobName,
			"--namespace", namespace,
			"--image", "docker.io/library/pytorch:2.1",
			"--workers", "4",
			"python train.py")

		replicaSpecs := getReplicaSpecs(crd, "pytorchReplicaSpecs")
		Expect(replicaSpecs).To(HaveKey("Master"))
		Expect(replicaSpecs).To(HaveKey("Worker"))

		master := replicaSpecs["Master"].(map[string]interface{})
		Expect(master["replicas"]).To(Equal(int64(1)))

		worker := replicaSpecs["Worker"].(map[string]interface{})
		Expect(worker["replicas"]).To(Equal(int64(3)))
	})

	It("TensorFlow with PS, Chief, and Evaluator roles", func() {
		crd := runSubmitDryRun("tensorflow",
			"--name", jobName,
			"--namespace", namespace,
			"--image", "docker.io/library/tensorflow:2.15",
			"--workers", "2",
			"--ps-count", "1",
			"--chief",
			"--evaluator",
			"python train.py")

		replicaSpecs := getReplicaSpecs(crd, "tfReplicaSpecs")
		Expect(replicaSpecs).To(HaveKey("Worker"))
		Expect(replicaSpecs).To(HaveKey("Chief"))
		Expect(replicaSpecs).To(HaveKey("PS"))
		Expect(replicaSpecs).To(HaveKey("Evaluator"))
	})

	It("MPI with --slots-per-worker", func() {
		crd := runSubmitDryRun("mpi",
			"--name", jobName,
			"--namespace", namespace,
			"--image", "docker.io/library/openmpi:4.1",
			"--workers", "2",
			"--slots-per-worker", "4",
			"mpirun train")

		replicaSpecs := getReplicaSpecs(crd, "mpiReplicaSpecs")
		Expect(replicaSpecs).To(HaveKey("Worker"))
		Expect(replicaSpecs).To(HaveKey("Launcher"))
	})

	It("submit with resource flags --gpus --cpus --mem", func() {
		crd := runSubmitDryRun("pytorch",
			"--name", jobName,
			"--namespace", namespace,
			"--image", "docker.io/library/pytorch:2.1",
			"--workers", "1",
			"--gpus", "2",
			"--cpus", "4",
			"--mem", "8Gi",
			"python train.py")

		replicaSpecs := getReplicaSpecs(crd, "pytorchReplicaSpecs")
		master := replicaSpecs["Master"].(map[string]interface{})
		container := getFirstContainer(master)

		resources := container["resources"].(map[string]interface{})
		limits := resources["limits"].(map[string]interface{})
		Expect(limits["nvidia.com/gpu"]).To(Equal("2"))
		Expect(limits["cpu"]).To(Equal("4"))
		Expect(limits["memory"]).To(Equal("8Gi"))
	})

	It("submit with --env --label --annotation", func() {
		crd := runSubmitDryRun("pytorch",
			"--name", jobName,
			"--namespace", namespace,
			"--image", "docker.io/library/pytorch:2.1",
			"--workers", "1",
			"--env", "FOO=bar",
			"--label", "team=ml",
			"--annotation", "version=v1",
			"python train.py")

		metadata := crd["metadata"].(map[string]interface{})

		labels := metadata["labels"].(map[string]interface{})
		Expect(labels["team"]).To(Equal("ml"))

		annotations := metadata["annotations"].(map[string]interface{})
		Expect(annotations["version"]).To(Equal("v1"))

		replicaSpecs := getReplicaSpecs(crd, "pytorchReplicaSpecs")
		master := replicaSpecs["Master"].(map[string]interface{})
		container := getFirstContainer(master)

		envs := container["env"].([]interface{})
		found := false
		for _, e := range envs {
			env := e.(map[string]interface{})
			if env["name"] == "FOO" {
				Expect(env["value"]).To(Equal("bar"))
				found = true
			}
		}
		Expect(found).To(BeTrue(), "env FOO=bar not found in container spec")
	})

	It("submit with --selector --toleration", func() {
		crd := runSubmitDryRun("pytorch",
			"--name", jobName,
			"--namespace", namespace,
			"--image", "docker.io/library/pytorch:2.1",
			"--workers", "1",
			"--selector", "gpu-type=A100",
			"--toleration", "gpu=true:NoSchedule",
			"python train.py")

		replicaSpecs := getReplicaSpecs(crd, "pytorchReplicaSpecs")
		master := replicaSpecs["Master"].(map[string]interface{})
		template := master["template"].(map[string]interface{})
		podSpec := template["spec"].(map[string]interface{})

		nodeSelector := podSpec["nodeSelector"].(map[string]interface{})
		Expect(nodeSelector["gpu-type"]).To(Equal("A100"))

		tolerations := podSpec["tolerations"].([]interface{})
		Expect(tolerations).To(HaveLen(1))
		tol := tolerations[0].(map[string]interface{})
		Expect(tol["key"]).To(Equal("gpu"))
		Expect(tol["value"]).To(Equal("true"))
		Expect(tol["effect"]).To(Equal("NoSchedule"))
	})
})
