package e2e_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Suspend and Resume", func() {
	var (
		jobName   string
		namespace string
	)

	BeforeEach(func() {
		jobName = fmt.Sprintf("v2-suspend-%d", GinkgoRandomSeed())
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

	It("should suspend and resume a job", func() {
		var out bytes.Buffer

		By("Submitting a PyTorch job")
		submitCmd := exec.Command(arenaV2Bin, "job", "submit", "pytorch",
			"--name", jobName,
			"--namespace", namespace,
			"--image", "docker.io/library/busybox:1.36",
			"--workers", "1",
			"sleep 300",
		)
		submitCmd.Stdout = &out
		submitCmd.Stderr = &out
		err := submitCmd.Run()
		Expect(err).NotTo(HaveOccurred(), "submit output: %s", out.String())
		out.Reset()

		By("Suspending the job")
		suspendCmd := exec.Command(arenaV2Bin, "job", "suspend", jobName,
			"--namespace", namespace)
		suspendCmd.Stdout = &out
		suspendCmd.Stderr = &out
		err = suspendCmd.Run()
		Expect(err).NotTo(HaveOccurred(), "suspend output: %s", out.String())
		Expect(out.String()).To(ContainSubstring("suspended"))
		out.Reset()

		By("Verifying suspend=true in CRD")
		kubectlCmd := exec.Command("kubectl", "get", "pytorchjob", jobName,
			"-n", namespace, "-o", "json")
		kubectlCmd.Stdout = &out
		kubectlCmd.Stderr = &out
		err = kubectlCmd.Run()
		Expect(err).NotTo(HaveOccurred(), "kubectl get output: %s", out.String())

		var crd map[string]interface{}
		Expect(json.Unmarshal(out.Bytes(), &crd)).NotTo(HaveOccurred())
		spec := crd["spec"].(map[string]interface{})
		runPolicy, ok := spec["runPolicy"].(map[string]interface{})
		Expect(ok).To(BeTrue(), "spec.runPolicy should exist")
		Expect(runPolicy["suspend"]).To(BeTrue(), "runPolicy.suspend should be true")
		out.Reset()

		By("Resuming the job")
		resumeCmd := exec.Command(arenaV2Bin, "job", "resume", jobName,
			"--namespace", namespace)
		resumeCmd.Stdout = &out
		resumeCmd.Stderr = &out
		err = resumeCmd.Run()
		Expect(err).NotTo(HaveOccurred(), "resume output: %s", out.String())
		Expect(out.String()).To(ContainSubstring("resumed"))
		out.Reset()

		By("Verifying suspend=false in CRD")
		kubectlCmd2 := exec.Command("kubectl", "get", "pytorchjob", jobName,
			"-n", namespace, "-o", "json")
		kubectlCmd2.Stdout = &out
		kubectlCmd2.Stderr = &out
		err = kubectlCmd2.Run()
		Expect(err).NotTo(HaveOccurred(), "kubectl get output: %s", out.String())

		var crd2 map[string]interface{}
		Expect(json.Unmarshal(out.Bytes(), &crd2)).NotTo(HaveOccurred())
		spec2 := crd2["spec"].(map[string]interface{})
		runPolicy2, ok := spec2["runPolicy"].(map[string]interface{})
		Expect(ok).To(BeTrue(), "spec.runPolicy should exist")
		Expect(runPolicy2["suspend"]).To(BeFalse(), "runPolicy.suspend should be false")
	})
})
