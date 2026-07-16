package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/kubeflow/arena/pkg/task"
)

func TestBuildConfigMap(t *testing.T) {
	ownerRef := metav1.OwnerReference{
		APIVersion: "kubeflow.org/v1",
		Kind:       "PyTorchJob",
		Name:       "my-job",
		UID:        "abc-123",
	}

	cm := buildConfigMap("my-job", "default", "name: my-job\nframework:\n  name: pytorch", ownerRef)

	assert.Equal(t, "ConfigMap", cm.GetKind())
	assert.Equal(t, "v1", cm.GetAPIVersion())
	assert.Equal(t, "my-job", cm.GetName())
	assert.Equal(t, "default", cm.GetNamespace())

	data, found, err := unstructured.NestedStringMap(cm.Object, "data")
	require.NoError(t, err)
	assert.True(t, found)
	assert.Contains(t, data["arena-v2.yaml"], "my-job")

	refs := cm.GetOwnerReferences()
	assert.Len(t, refs, 1)
	assert.Equal(t, "my-job", refs[0].Name)
	assert.Equal(t, "PyTorchJob", refs[0].Kind)
}

func TestBuildTensorBoardDeployment(t *testing.T) {
	ownerRef := metav1.OwnerReference{
		APIVersion: "kubeflow.org/v1",
		Kind:       "PyTorchJob",
		Name:       "my-job",
		UID:        "abc-123",
	}

	deploy := buildTensorBoardDeployment("my-job-tensorboard", "my-job", "default", "tensorflow/tensorflow:2.15", "/logs", nil, ownerRef)

	assert.Equal(t, "Deployment", deploy.GetKind())
	assert.Equal(t, "apps/v1", deploy.GetAPIVersion())
	assert.Equal(t, "my-job-tensorboard", deploy.GetName())
	assert.Equal(t, "default", deploy.GetNamespace())

	replicas, found, err := unstructured.NestedInt64(deploy.Object, "spec", "replicas")
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, int64(1), replicas)

	refs := deploy.GetOwnerReferences()
	assert.Len(t, refs, 1)
	assert.Equal(t, "my-job", refs[0].Name)
}

func TestBuildTensorBoardDeployment_Labels(t *testing.T) {
	ownerRef := metav1.OwnerReference{
		APIVersion: "kubeflow.org/v1",
		Kind:       "PyTorchJob",
		Name:       "my-job",
		UID:        "abc-123",
	}

	deploy := buildTensorBoardDeployment("my-job-tensorboard", "my-job", "default", "img", "/logs", nil, ownerRef)

	matchLabels, found, err := unstructured.NestedMap(deploy.Object, "spec", "selector", "matchLabels")
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "tensorboard", matchLabels["arena.io/component"])
	assert.Equal(t, "my-job", matchLabels["arena.io/job-name"])

	templateLabels, found, err := unstructured.NestedMap(deploy.Object, "spec", "template", "metadata", "labels")
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "tensorboard", templateLabels["arena.io/component"])
	assert.Equal(t, "my-job", templateLabels["arena.io/job-name"])
}

func TestBuildTensorBoardDeployment_DefaultImage(t *testing.T) {
	ownerRef := metav1.OwnerReference{
		APIVersion: "kubeflow.org/v1",
		Kind:       "PyTorchJob",
		Name:       "my-job",
	}

	deploy := buildTensorBoardDeployment("my-job-tensorboard", "my-job", "default", "tensorflow/tensorflow:2.15", "", nil, ownerRef)

	containers, found, err := unstructured.NestedSlice(deploy.Object, "spec", "template", "spec", "containers")
	require.NoError(t, err)
	assert.True(t, found)
	assert.Len(t, containers, 1)

	container := containers[0].(map[string]interface{})
	assert.Equal(t, "tensorflow/tensorflow:2.15", container["image"])
}

func TestBuildTensorBoardDeployment_CustomLogDir(t *testing.T) {
	ownerRef := metav1.OwnerReference{
		APIVersion: "kubeflow.org/v1",
		Kind:       "PyTorchJob",
		Name:       "my-job",
	}

	deploy := buildTensorBoardDeployment("my-job-tensorboard", "my-job", "default", "tensorflow/tensorflow:2.15", "/custom/logs", nil, ownerRef)

	containers, _, _ := unstructured.NestedSlice(deploy.Object, "spec", "template", "spec", "containers")
	container := containers[0].(map[string]interface{})
	args := container["args"].([]interface{})

	// Should contain --logdir /custom/logs
	assert.Contains(t, args, "--logdir")
	assert.Contains(t, args, "/custom/logs")
}

func TestBuildTensorBoardService(t *testing.T) {
	ownerRef := metav1.OwnerReference{
		APIVersion: "kubeflow.org/v1",
		Kind:       "PyTorchJob",
		Name:       "my-job",
		UID:        "abc-123",
	}

	svc := buildTensorBoardService("my-job-tensorboard", "my-job", "default", ownerRef)

	assert.Equal(t, "Service", svc.GetKind())
	assert.Equal(t, "v1", svc.GetAPIVersion())
	assert.Equal(t, "my-job-tensorboard", svc.GetName())
	assert.Equal(t, "default", svc.GetNamespace())

	ports, found, err := unstructured.NestedSlice(svc.Object, "spec", "ports")
	require.NoError(t, err)
	assert.True(t, found)
	assert.Len(t, ports, 1)

	port := ports[0].(map[string]interface{})
	assert.Equal(t, int64(6006), port["port"])
	assert.Equal(t, int64(6006), port["targetPort"])

	refs := svc.GetOwnerReferences()
	assert.Len(t, refs, 1)
}

func TestTensorBoardPodLabels(t *testing.T) {
	labels := tensorBoardPodLabels("my-job")
	assert.Equal(t, "tensorboard", labels["arena.io/component"])
	assert.Equal(t, "my-job", labels["arena.io/job-name"])
}

func TestPtrBool(t *testing.T) {
	truePtr := ptrBool(true)
	assert.True(t, *truePtr)

	falsePtr := ptrBool(false)
	assert.False(t, *falsePtr)
}

func TestBuildConfigMap_OwnerReferenceFields(t *testing.T) {
	ownerRef := metav1.OwnerReference{
		APIVersion:         "kubeflow.org/v1",
		Kind:               "PyTorchJob",
		Name:               "owner-job",
		UID:                "uid-456",
		BlockOwnerDeletion: ptrBool(true),
		Controller:         ptrBool(true),
	}

	cm := buildConfigMap("owner-job", "ns1", "yaml-content", ownerRef)
	refs := cm.GetOwnerReferences()
	require.Len(t, refs, 1)
	assert.Equal(t, "owner-job", refs[0].Name)
	assert.Equal(t, "uid-456", string(refs[0].UID))
	assert.True(t, *refs[0].BlockOwnerDeletion)
	assert.True(t, *refs[0].Controller)
}

// ---------------------------------------------------------------------------
// TensorBoard storage injection tests
// ---------------------------------------------------------------------------

// extractTBPodSpec is a helper that returns the podSpec and container maps
// from a TensorBoard Deployment built by buildTensorBoardDeployment.
func extractTBPodSpec(t *testing.T, deploy *unstructured.Unstructured) (map[string]interface{}, map[string]interface{}) {
	t.Helper()
	containers, found, err := unstructured.NestedSlice(deploy.Object, "spec", "template", "spec", "containers")
	require.NoError(t, err)
	require.True(t, found)
	require.Len(t, containers, 1)
	container := containers[0].(map[string]interface{})
	podSpec, found, err := unstructured.NestedMap(deploy.Object, "spec", "template", "spec")
	require.NoError(t, err)
	require.True(t, found)
	return podSpec, container
}

func TestBuildTensorBoardDeployment_NoStorages(t *testing.T) {
	ownerRef := metav1.OwnerReference{Name: "test", UID: "123"}
	tk := &task.Task{
		Logging: task.Logging{
			TensorBoard: &task.TensorBoardConfig{Enabled: true},
		},
	}
	deploy := buildTensorBoardDeployment("tb", "test-job", "default", "img", "/logs", tk, ownerRef)

	podSpec, container := extractTBPodSpec(t, deploy)
	_, hasVolumes := podSpec["volumes"]
	assert.False(t, hasVolumes, "no volumes expected when task has no storages")
	_, hasMounts := container["volumeMounts"]
	assert.False(t, hasMounts, "no volumeMounts expected when task has no storages")
}

func TestBuildTensorBoardDeployment_AllStorages(t *testing.T) {
	ownerRef := metav1.OwnerReference{Name: "test", UID: "123"}
	tk := &task.Task{
		Storages: []task.Storage{
			{Name: "data", PVC: "pvc1", MountPath: "/data"},
			{Name: "cache", HostPath: "/tmp/cache", MountPath: "/cache"},
			{Name: "shm", SHM: "8Gi"},
		},
		Logging: task.Logging{
			TensorBoard: &task.TensorBoardConfig{Enabled: true},
		},
	}
	deploy := buildTensorBoardDeployment("tb", "test-job", "default", "img", "/logs", tk, ownerRef)

	podSpec, container := extractTBPodSpec(t, deploy)

	volumes := podSpec["volumes"].([]interface{})
	assert.Len(t, volumes, 3, "all storages should become volumes")

	mounts := container["volumeMounts"].([]interface{})
	assert.Len(t, mounts, 3, "all storages should have volumeMounts when no mounts field is set")

	// Verify the SHM storage gets the default /dev/shm mount path
	shmMount := mounts[2].(map[string]interface{})
	assert.Equal(t, "shm", shmMount["name"])
	assert.Equal(t, "/dev/shm", shmMount["mountPath"])
}

func TestBuildTensorBoardDeployment_MountsOverride(t *testing.T) {
	ownerRef := metav1.OwnerReference{Name: "test", UID: "123"}
	tk := &task.Task{
		Storages: []task.Storage{
			{Name: "data", PVC: "pvc1", MountPath: "/data"},
			{Name: "logs", PVC: "pvc2", MountPath: "/logs"},
			{Name: "cache", PVC: "pvc3", MountPath: "/cache"},
		},
		Logging: task.Logging{
			TensorBoard: &task.TensorBoardConfig{
				Enabled: true,
				Mounts: []task.Mount{
					{Name: "logs", MountPath: "/tb/logs"},
				},
			},
		},
	}
	deploy := buildTensorBoardDeployment("tb", "test-job", "default", "img", "/logs", tk, ownerRef)

	podSpec, container := extractTBPodSpec(t, deploy)

	volumes := podSpec["volumes"].([]interface{})
	assert.Len(t, volumes, 3, "all storages should be volumes even when mounts field is set")

	mounts := container["volumeMounts"].([]interface{})
	assert.Len(t, mounts, 1, "only mounted storage should have volumeMount")

	mount := mounts[0].(map[string]interface{})
	assert.Equal(t, "logs", mount["name"])
	assert.Equal(t, "/tb/logs", mount["mountPath"], "mount path should override storage default")
}

func TestBuildTensorBoardDeployment_MountsSubPath(t *testing.T) {
	ownerRef := metav1.OwnerReference{Name: "test", UID: "123"}
	tk := &task.Task{
		Storages: []task.Storage{
			{Name: "data", PVC: "pvc1", MountPath: "/data", SubPath: "original"},
		},
		Logging: task.Logging{
			TensorBoard: &task.TensorBoardConfig{
				Enabled: true,
				Mounts: []task.Mount{
					{Name: "data", SubPath: "override"},
				},
			},
		},
	}
	deploy := buildTensorBoardDeployment("tb", "test-job", "default", "img", "/logs", tk, ownerRef)

	_, container := extractTBPodSpec(t, deploy)
	mounts := container["volumeMounts"].([]interface{})
	mount := mounts[0].(map[string]interface{})
	assert.Equal(t, "override", mount["subPath"], "mount subPath should override storage default")
	assert.Equal(t, "/data", mount["mountPath"], "mount path falls back to storage default when not overridden")
}

func TestBuildTensorBoardDeployment_NilTensorBoardWithStorages(t *testing.T) {
	// Regression: buildTensorBoardDeployment must not panic when the task has
	// storages but Logging.TensorBoard is nil. The function's own guard
	// (`t != nil && len(t.Storages) > 0`) implies it should handle this case.
	ownerRef := metav1.OwnerReference{Name: "test", UID: "123"}
	tk := &task.Task{
		Storages: []task.Storage{
			{Name: "data", PVC: "pvc1", MountPath: "/data"},
			{Name: "cache", HostPath: "/tmp/cache", MountPath: "/cache"},
		},
		Logging: task.Logging{
			TensorBoard: nil,
		},
	}

	require.NotPanics(t, func() {
		deploy := buildTensorBoardDeployment("tb", "test-job", "default", "img", "/logs", tk, ownerRef)
		podSpec, container := extractTBPodSpec(t, deploy)

		// All storages become volumes even without TensorBoard config
		volumes := podSpec["volumes"].([]interface{})
		assert.Len(t, volumes, 2)

		// Without TensorBoard mounts override, all storages get volumeMounts
		mounts := container["volumeMounts"].([]interface{})
		assert.Len(t, mounts, 2)
	})
}

