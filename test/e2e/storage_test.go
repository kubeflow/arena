package e2e_test

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Storage volume mounts", func() {
	var namespace string

	BeforeEach(func() {
		namespace = "default"
	})

	Describe("ConfigMap mounts", func() {
		BeforeEach(func() {
			// Create the prerequisite ConfigMap
			var out bytes.Buffer
			cmd := exec.Command("kubectl", "apply", "-f",
				filepath.Join("..", "testdata", "prereq-configmap.yaml"),
				"-n", namespace)
			cmd.Stdout = &out
			cmd.Stderr = &out
			Expect(cmd.Run()).NotTo(HaveOccurred(),
				"kubectl apply prereq-configmap failed: %s", out.String())
		})

		AfterEach(func() {
			// Delete the job
			var out bytes.Buffer
			delCmd := exec.Command(arenaV2Bin, "job", "delete", "test-configmap",
				"--namespace", namespace)
			delCmd.Stdout = &out
			delCmd.Stderr = &out
			_ = delCmd.Run()

			// Delete the ConfigMap
			out.Reset()
			cmCmd := exec.Command("kubectl", "delete", "configmap", "app-config",
				"-n", namespace, "--ignore-not-found")
			cmCmd.Stdout = &out
			cmCmd.Stderr = &out
			_ = cmCmd.Run()
		})

		It("should mount ConfigMap as volumes with correct paths", func() {
			By("Submitting a job with ConfigMap storages")
			var out bytes.Buffer
			runCmd := exec.Command(arenaV2Bin, "job", "run", "-f",
				filepath.Join("..", "testdata", "configmap.yaml"),
				"--namespace", namespace)
			runCmd.Stdout = &out
			runCmd.Stderr = &out
			Expect(runCmd.Run()).NotTo(HaveOccurred(),
				"job run failed: %s", out.String())

			By("Fetching the PyTorchJob CRD")
			out.Reset()
			getCmd := exec.Command("kubectl", "get", "pytorchjob", "test-configmap",
				"-n", namespace, "-o", "json")
			getCmd.Stdout = &out
			getCmd.Stderr = &out
			Expect(getCmd.Run()).NotTo(HaveOccurred(),
				"kubectl get pytorchjob failed: %s", out.String())

			var crd map[string]interface{}
			Expect(json.Unmarshal(out.Bytes(), &crd)).NotTo(HaveOccurred())

			spec := crd["spec"].(map[string]interface{})
			replicaSpecs := spec["pytorchReplicaSpecs"].(map[string]interface{})
			master := replicaSpecs["Master"].(map[string]interface{})
			template := master["template"].(map[string]interface{})
			podSpec := template["spec"].(map[string]interface{})

			volumes := podSpec["volumes"].([]interface{})
			Expect(volumes).To(HaveLen(2))

			volMap := make(map[string]map[string]interface{})
			for _, v := range volumes {
				vol := v.(map[string]interface{})
				volMap[vol["name"].(string)] = vol
			}

			Expect(volMap).To(HaveKey("configs"))
			cmSpec := volMap["configs"]["configMap"].(map[string]interface{})
			Expect(cmSpec["name"]).To(Equal("app-config"))

			Expect(volMap).To(HaveKey("single-file"))
			cmSpec2 := volMap["single-file"]["configMap"].(map[string]interface{})
			Expect(cmSpec2["name"]).To(Equal("app-config"))

			containers := podSpec["containers"].([]interface{})
			container := containers[0].(map[string]interface{})
			mounts := container["volumeMounts"].([]interface{})
			Expect(mounts).To(HaveLen(2))

			mountMap := make(map[string]map[string]interface{})
			for _, m := range mounts {
				mnt := m.(map[string]interface{})
				mountMap[mnt["name"].(string)] = mnt
			}

			Expect(mountMap["configs"]["mountPath"]).To(Equal("/etc/config"))
			Expect(mountMap["single-file"]["mountPath"]).To(Equal("/app/settings.yaml"))
			Expect(mountMap["single-file"]["subPath"]).To(Equal("settings.yaml"))
		})
	})

	Describe("Secret mounts", func() {
		BeforeEach(func() {
			// Create the prerequisite Secrets
			var out bytes.Buffer
			cmd := exec.Command("kubectl", "apply", "-f",
				filepath.Join("..", "testdata", "prereq-secret.yaml"),
				"-n", namespace)
			cmd.Stdout = &out
			cmd.Stderr = &out
			Expect(cmd.Run()).NotTo(HaveOccurred(),
				"kubectl apply prereq-secret failed: %s", out.String())
		})

		AfterEach(func() {
			// Delete the job
			var out bytes.Buffer
			delCmd := exec.Command(arenaV2Bin, "job", "delete", "test-secret",
				"--namespace", namespace)
			delCmd.Stdout = &out
			delCmd.Stderr = &out
			_ = delCmd.Run()

			// Delete the Secrets
			out.Reset()
			delSec := exec.Command("kubectl", "delete", "secret",
				"db-credentials", "ssh-keys",
				"-n", namespace, "--ignore-not-found")
			delSec.Stdout = &out
			delSec.Stderr = &out
			_ = delSec.Run()
		})

		It("should mount Secrets as volumes with correct paths", func() {
			By("Submitting a job with Secret storages")
			var out bytes.Buffer
			runCmd := exec.Command(arenaV2Bin, "job", "run", "-f",
				filepath.Join("..", "testdata", "secret.yaml"),
				"--namespace", namespace)
			runCmd.Stdout = &out
			runCmd.Stderr = &out
			Expect(runCmd.Run()).NotTo(HaveOccurred(),
				"job run failed: %s", out.String())

			By("Fetching the PyTorchJob CRD")
			out.Reset()
			getCmd := exec.Command("kubectl", "get", "pytorchjob", "test-secret",
				"-n", namespace, "-o", "json")
			getCmd.Stdout = &out
			getCmd.Stderr = &out
			Expect(getCmd.Run()).NotTo(HaveOccurred(),
				"kubectl get pytorchjob failed: %s", out.String())

			var crd map[string]interface{}
			Expect(json.Unmarshal(out.Bytes(), &crd)).NotTo(HaveOccurred())

			spec := crd["spec"].(map[string]interface{})
			replicaSpecs := spec["pytorchReplicaSpecs"].(map[string]interface{})
			master := replicaSpecs["Master"].(map[string]interface{})
			template := master["template"].(map[string]interface{})
			podSpec := template["spec"].(map[string]interface{})

			volumes := podSpec["volumes"].([]interface{})
			Expect(volumes).To(HaveLen(2))

			volMap := make(map[string]map[string]interface{})
			for _, v := range volumes {
				vol := v.(map[string]interface{})
				volMap[vol["name"].(string)] = vol
			}

			Expect(volMap).To(HaveKey("creds"))
			sSpec := volMap["creds"]["secret"].(map[string]interface{})
			Expect(sSpec["secretName"]).To(Equal("db-credentials"))

			Expect(volMap).To(HaveKey("ssh-key"))
			sSpec2 := volMap["ssh-key"]["secret"].(map[string]interface{})
			Expect(sSpec2["secretName"]).To(Equal("ssh-keys"))

			containers := podSpec["containers"].([]interface{})
			container := containers[0].(map[string]interface{})
			mounts := container["volumeMounts"].([]interface{})
			Expect(mounts).To(HaveLen(2))

			mountMap := make(map[string]map[string]interface{})
			for _, m := range mounts {
				mnt := m.(map[string]interface{})
				mountMap[mnt["name"].(string)] = mnt
			}

			Expect(mountMap["creds"]["mountPath"]).To(Equal("/secrets"))
			Expect(mountMap["ssh-key"]["mountPath"]).To(Equal("/root/.ssh/id_rsa"))
			Expect(mountMap["ssh-key"]["subPath"]).To(Equal("id_rsa"))
		})
	})
})
