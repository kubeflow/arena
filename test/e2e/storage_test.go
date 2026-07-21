package e2e_test

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const prereqConfigMapYAML = `apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
  namespace: default
data:
  settings.yaml: |
    model: resnet50
    batch_size: 32
    epochs: 10
`

const prereqSecretYAML = `apiVersion: v1
kind: Secret
metadata:
  name: db-credentials
  namespace: default
type: Opaque
stringData:
  username: admin
  password: test-password
---
apiVersion: v1
kind: Secret
metadata:
  name: ssh-keys
  namespace: default
type: Opaque
stringData:
  id_rsa: |
    -----BEGIN RSA PRIVATE KEY-----
    fake-key-for-e2e-testing-only
    -----END RSA PRIVATE KEY-----
`

const configmapJobYAML = `version: 0.1.0
name: test-configmap
framework:
  name: pytorch
image: docker.io/pytorch/pytorch:2.1
run: cat /etc/config/settings.yaml
worker:
  replicas: 1
  resources:
    cpu: 1
    memory: 1Gi
storages:
  - name: configs
    configmap: app-config
    mount_path: /etc/config
  - name: single-file
    configmap: app-config
    key: settings.yaml
    mount_path: /app/settings.yaml
`

const secretJobYAML = `version: 0.1.0
name: test-secret
framework:
  name: pytorch
image: docker.io/pytorch/pytorch:2.1
run: test -f /root/.ssh/id_rsa && echo "secret mounted ok"
worker:
  replicas: 1
  resources:
    cpu: 1
    memory: 1Gi
storages:
  - name: creds
    secret: db-credentials
    mount_path: /secrets
  - name: ssh-key
    secret: ssh-keys
    key: id_rsa
    mount_path: /root/.ssh/id_rsa
`

var _ = Describe("Storage volume mounts", func() {
	var namespace string

	BeforeEach(func() {
		namespace = "default"
	})

	Describe("ConfigMap mounts", func() {
		BeforeEach(func() {
			// Create the prerequisite ConfigMap
			var out bytes.Buffer
			cmd := exec.Command("kubectl", "apply", "-f", "-")
			cmd.Stdin = bytes.NewReader([]byte(prereqConfigMapYAML))
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
			yamlPath, err := createTempYAML(configmapJobYAML)
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
			cmd := exec.Command("kubectl", "apply", "-f", "-")
			cmd.Stdin = bytes.NewReader([]byte(prereqSecretYAML))
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
			yamlPath, err := createTempYAML(secretJobYAML)
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
