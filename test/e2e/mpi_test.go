package e2e_test

import (
	"bytes"
	"fmt"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	outputpkg "github.com/kubeflow/arena/pkg/output"
)

var _ = Describe("MPI-based Jobs", func() {
	var namespace string

	BeforeEach(func() {
		namespace = "default"
	})

	frameworkLifecycle := func(framework, image string) {
		jobName := fmt.Sprintf("v2-%s-%d", framework, GinkgoRandomSeed())

		var out bytes.Buffer
		var err error

		By(fmt.Sprintf("Submitting a %s job", framework))
		submitCmd := exec.Command(arenaV2Bin, "job", "submit", framework,
			"--name", jobName,
			"--namespace", namespace,
			"--image", image,
			"--workers", "2",
			"mpirun train",
		)
		submitCmd.Stdout = &out
		submitCmd.Stderr = &out
		err = submitCmd.Run()
		Expect(err).NotTo(HaveOccurred(), "submit output: %s", out.String())
		out.Reset()

		By(fmt.Sprintf("Verifying framework label for %s", framework))
		getCmd := exec.Command(arenaV2Bin, "job", "get", jobName,
			"--namespace", namespace, "-o", string(outputpkg.FormatJSON))
		getCmd.Stdout = &out
		getCmd.Stderr = &out
		err = getCmd.Run()
		Expect(err).NotTo(HaveOccurred(), "get output: %s", out.String())
		Expect(out.String()).To(ContainSubstring(framework))
		out.Reset()

		By("Cleaning up")
		delCmd := exec.Command(arenaV2Bin, "job", "delete", jobName,
			"--namespace", namespace)
		delCmd.Stdout = &out
		delCmd.Stderr = &out
		_ = delCmd.Run()
	}

	It("MPI job lifecycle", func() {
		frameworkLifecycle("mpi", "mpi:latest")
	})

	It("Horovod job lifecycle", func() {
		frameworkLifecycle("horovod", "horovod:latest")
	})

	It("DeepSpeed job lifecycle", func() {
		frameworkLifecycle("deepspeed", "deepspeed:latest")
	})
})
