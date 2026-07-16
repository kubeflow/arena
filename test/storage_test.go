package integration

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kubeflow/arena/pkg/task"
)

func TestStorageConfigMap(t *testing.T) {
	yamlPath := filepath.Join(testdataDir(t), "configmap.yaml")
	taskObj, err := task.LoadFromFile(yamlPath)
	require.NoError(t, err, "LoadFromFile should succeed for configmap.yaml")

	assert.Equal(t, "test-configmap", taskObj.Name)
	require.Len(t, taskObj.Storages, 2, "configmap.yaml should define 2 storages")

	p := providerFor(taskObj.Framework.Name)
	require.NotNil(t, p)

	crd, err := p.BuildCRD(taskObj)
	require.NoError(t, err)
	assert.Equal(t, "PyTorchJob", crd.GetKind())

	spec := crd.Object["spec"].(map[string]interface{})
	replicaSpecs := spec["pytorchReplicaSpecs"].(map[string]interface{})
	master := replicaSpecs["Master"].(map[string]interface{})
	template := master["template"].(map[string]interface{})
	podSpec := template["spec"].(map[string]interface{})

	volumes := podSpec["volumes"].([]interface{})
	require.Len(t, volumes, 2, "should have 2 volumes")

	volMap := make(map[string]map[string]interface{})
	for _, v := range volumes {
		vol := v.(map[string]interface{})
		volMap[vol["name"].(string)] = vol
	}

	cm1, ok := volMap["configs"]
	require.True(t, ok, "volume 'configs' should exist")
	cmSpec := cm1["configMap"].(map[string]interface{})
	assert.Equal(t, "app-config", cmSpec["name"])

	cm2, ok := volMap["single-file"]
	require.True(t, ok, "volume 'single-file' should exist")
	cmSpec2 := cm2["configMap"].(map[string]interface{})
	assert.Equal(t, "app-config", cmSpec2["name"])

	containers := podSpec["containers"].([]interface{})
	container := containers[0].(map[string]interface{})
	mounts := container["volumeMounts"].([]interface{})
	require.Len(t, mounts, 2, "should have 2 volumeMounts")

	mountMap := make(map[string]map[string]interface{})
	for _, m := range mounts {
		mnt := m.(map[string]interface{})
		mountMap[mnt["name"].(string)] = mnt
	}

	m1, ok := mountMap["configs"]
	require.True(t, ok, "mount 'configs' should exist")
	assert.Equal(t, "/etc/config", m1["mountPath"])
	_, hasSubPath1 := m1["subPath"]
	assert.False(t, hasSubPath1, "mount 'configs' should not have subPath (no key set)")

	m2, ok := mountMap["single-file"]
	require.True(t, ok, "mount 'single-file' should exist")
	assert.Equal(t, "/app/settings.yaml", m2["mountPath"])
	assert.Equal(t, "settings.yaml", m2["subPath"], "mount 'single-file' should have subPath = key")
}

func TestStorageSecret(t *testing.T) {
	yamlPath := filepath.Join(testdataDir(t), "secret.yaml")
	taskObj, err := task.LoadFromFile(yamlPath)
	require.NoError(t, err, "LoadFromFile should succeed for secret.yaml")

	assert.Equal(t, "test-secret", taskObj.Name)
	require.Len(t, taskObj.Storages, 2, "secret.yaml should define 2 storages")

	p := providerFor(taskObj.Framework.Name)
	require.NotNil(t, p)

	crd, err := p.BuildCRD(taskObj)
	require.NoError(t, err)
	assert.Equal(t, "PyTorchJob", crd.GetKind())

	spec := crd.Object["spec"].(map[string]interface{})
	replicaSpecs := spec["pytorchReplicaSpecs"].(map[string]interface{})
	master := replicaSpecs["Master"].(map[string]interface{})
	template := master["template"].(map[string]interface{})
	podSpec := template["spec"].(map[string]interface{})

	volumes := podSpec["volumes"].([]interface{})
	require.Len(t, volumes, 2, "should have 2 volumes")

	volMap := make(map[string]map[string]interface{})
	for _, v := range volumes {
		vol := v.(map[string]interface{})
		volMap[vol["name"].(string)] = vol
	}

	s1, ok := volMap["creds"]
	require.True(t, ok, "volume 'creds' should exist")
	sSpec := s1["secret"].(map[string]interface{})
	assert.Equal(t, "db-credentials", sSpec["secretName"])

	s2, ok := volMap["ssh-key"]
	require.True(t, ok, "volume 'ssh-key' should exist")
	sSpec2 := s2["secret"].(map[string]interface{})
	assert.Equal(t, "ssh-keys", sSpec2["secretName"])

	containers := podSpec["containers"].([]interface{})
	container := containers[0].(map[string]interface{})
	mounts := container["volumeMounts"].([]interface{})
	require.Len(t, mounts, 2, "should have 2 volumeMounts")

	mountMap := make(map[string]map[string]interface{})
	for _, m := range mounts {
		mnt := m.(map[string]interface{})
		mountMap[mnt["name"].(string)] = mnt
	}

	m1, ok := mountMap["creds"]
	require.True(t, ok, "mount 'creds' should exist")
	assert.Equal(t, "/secrets", m1["mountPath"])
	_, hasSubPath1 := m1["subPath"]
	assert.False(t, hasSubPath1, "mount 'creds' should not have subPath (no key set)")

	m2, ok := mountMap["ssh-key"]
	require.True(t, ok, "mount 'ssh-key' should exist")
	assert.Equal(t, "/root/.ssh/id_rsa", m2["mountPath"])
	assert.Equal(t, "id_rsa", m2["subPath"], "mount 'ssh-key' should have subPath = key")
}
