package task

import (
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestValidateVersion(t *testing.T) {
	base := func() *Task {
		return &Task{
			Name: "test", Image: "img:1", Run: "echo",
			Framework: Framework{Name: "pytorch"},
			Worker: &Worker{Replicas: 1},
		}
	}

	tests := []struct {
		name    string
		version string
		wantErr bool
	}{
		{"empty defaults to 0.1.0", "", false},
		{"valid 0.1.0", "0.1.0", false},
		{"valid 0.1.5", "0.1.5", false},
		{"unknown minor", "0.2.0", true},
		{"unknown major", "1.0.0", true},
		{"bad format v1", "v1", true},
		{"bad format latest", "latest", true},
		{"bad format 1.0", "1.0", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := base()
			task.Version = tt.version
			err := Validate(task)
			if tt.wantErr && err == nil {
				t.Errorf("expected error for version %q", tt.version)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error for version %q: %v", tt.version, err)
			}
		})
	}
}

func TestValidateVersionDefault(t *testing.T) {
	task := &Task{
		Name: "test", Image: "img:1", Run: "echo",
		Framework: Framework{Name: "pytorch"},
		Worker: &Worker{Replicas: 1},
	}
	task.SetDefaults()
	err := Validate(task)
	require.NoError(t, err)
	if task.Version != DefaultSchemaVersion {
		t.Errorf("expected version %q, got %q", DefaultSchemaVersion, task.Version)
	}
}

func TestMinimalYAMLParse(t *testing.T) {
	data := []byte(`
name: my-job
framework:
  name: pytorch
image: pytorch:2.1
run: python train.py --epochs 10
worker:
  replicas: 1
  resources:
    nvidia.com/gpu: 1
`)
	task, err := LoadFromBytes(data)
	require.NoError(t, err)
	assert.Equal(t, "my-job", task.Name)
	assert.Equal(t, "pytorch", task.Framework.Name)
	assert.Equal(t, "python train.py --epochs 10", task.Run)
	assert.Equal(t, 1, task.Worker.Replicas)
	assert.Equal(t, "1", task.Worker.Resources["nvidia.com/gpu"])
}

func TestFullYAMLParse(t *testing.T) {
	data := []byte(`
version: 0.1.0
name: llm-finetune
namespace: ml-team
labels:
  team: platform
annotations:
  note: experiment
image: nvcr.io/nvidia/pytorch:23.10
run: torchrun train.py
shell: /bin/bash
working_dir: /root
framework:
  name: pytorch
  options:
    nproc_per_node: auto
envs:
  NCCL_DEBUG: INFO
  HF_TOKEN:
    secret: my-hf-creds
    key: token
  DB_HOST:
    configmap: db-config
    key: host
worker:
  replicas: 4
  resources:
    nvidia.com/gpu: 8
    cpu: 32
    memory: 128Gi
storages:
  - name: dataset
    mount_path: /data
    pvc: dataset-pvc
  - name: shm
    mount_path: /dev/shm
    shm: 64Gi
scheduling:
  priority: 100
  gang:
    enabled: false
  node_selector:
    disktype: ssd
  tolerations:
    - key: nvidia.com/gpu
      operator: Exists
      effect: NoSchedule
lifecycle:
  clean_pod_policy: Running
  backoff_limit: 6
image_pull_policy: Always
restart: OnFailure
`)
	task, err := LoadFromBytes(data)
	require.NoError(t, err)
	assert.Equal(t, "0.1.0", task.Version)
	assert.Equal(t, "ml-team", task.Namespace)
	assert.Equal(t, "platform", task.Labels["team"])
	assert.Equal(t, "/bin/bash", task.Shell)
	assert.Equal(t, "/root", task.WorkingDir)
	assert.Equal(t, "auto", task.Framework.Options.NprocPerNode)
	assert.Equal(t, "INFO", task.Envs["NCCL_DEBUG"].Value)
	assert.Equal(t, "my-hf-creds", task.Envs["HF_TOKEN"].Secret.Name)
	assert.Equal(t, "token", task.Envs["HF_TOKEN"].Secret.Key)
	assert.Equal(t, "db-config", task.Envs["DB_HOST"].ConfigMap.Name)
	assert.Equal(t, 4, task.Worker.Replicas)
	assert.Equal(t, "8", task.Worker.Resources["nvidia.com/gpu"])
	assert.Len(t, task.Storages, 2)
	assert.Equal(t, "dataset-pvc", task.Storages[0].PVC)
	assert.Equal(t, "64Gi", task.Storages[1].SHM)
	assert.Equal(t, 100, task.Scheduling.Priority)
	assert.Equal(t, "ssd", task.Scheduling.NodeSelector["disktype"])
	assert.Len(t, task.Scheduling.Tolerations, 1)
	assert.Equal(t, "Running", task.Lifecycle.CleanPodPolicy)
	assert.Equal(t, 6, *task.Lifecycle.BackoffLimit)
	assert.Equal(t, "Always", task.ImagePullPolicy)
	assert.Equal(t, "OnFailure", task.Restart)
}

func TestEnvValuePlainString(t *testing.T) {
	var e EnvValue
	err := yaml.Unmarshal([]byte(`"hello"`), &e)
	require.NoError(t, err)
	assert.Equal(t, "hello", e.Value)
	assert.Nil(t, e.Secret)
	assert.Nil(t, e.ConfigMap)
}

func TestEnvValueSecretRef(t *testing.T) {
	var e EnvValue
	err := yaml.Unmarshal([]byte(`{secret: my-secret, key: token}`), &e)
	require.NoError(t, err)
	assert.Equal(t, "", e.Value)
	require.NotNil(t, e.Secret)
	assert.Equal(t, "my-secret", e.Secret.Name)
	assert.Equal(t, "token", e.Secret.Key)
}

func TestEnvValueConfigMapRef(t *testing.T) {
	var e EnvValue
	err := yaml.Unmarshal([]byte(`{configmap: my-cm, key: host}`), &e)
	require.NoError(t, err)
	require.NotNil(t, e.ConfigMap)
	assert.Equal(t, "my-cm", e.ConfigMap.Name)
	assert.Equal(t, "host", e.ConfigMap.Key)
}

func TestFrameworkConfigPyTorch(t *testing.T) {
	data := []byte(`
name: t
framework:
  name: pytorch
  options:
    nproc_per_node: gpu
image: x:1
run: train
worker:
  replicas: 1
`)
	task, err := LoadFromBytes(data)
	require.NoError(t, err)
	assert.Equal(t, "gpu", task.Framework.Options.NprocPerNode)
}

func TestFrameworkConfigTensorFlow(t *testing.T) {
	data := []byte(`
name: t
framework:
  name: tensorflow
image: x:1
run: train
worker:
  replicas: 1
chief: {}
ps:
  replicas: 2
evaluator: {}
`)
	task, err := LoadFromBytes(data)
	require.NoError(t, err)
	require.NotNil(t, task.Chief)
	require.NotNil(t, task.PS)
	assert.Equal(t, 2, task.PS.Replicas)
	require.NotNil(t, task.Evaluator)
}

func TestFrameworkConfigMPI(t *testing.T) {
	data := []byte(`
name: t
framework:
  name: mpi
  options:
    slots_per_worker: 4
    mounts_on_launcher: true
    gpu_topology: true
image: x:1
run: train
worker:
  replicas: 4
`)
	task, err := LoadFromBytes(data)
	require.NoError(t, err)
	assert.Equal(t, 4, task.Framework.Options.SlotsPerWorker)
	assert.True(t, task.Framework.Options.MountsOnLauncher)
	assert.True(t, task.Framework.Options.GPUTopology)
}

func TestValidateRunRequired(t *testing.T) {
	task := &Task{
		Name:      "t",
		Image:     "x:1",
		Framework: Framework{Name: "pytorch"},
		Worker: &Worker{Replicas: 1},
	}
	err := Validate(task)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "run is required")
}

func TestValidateSuccessPolicy(t *testing.T) {
	task := &Task{
		Name:      "t",
		Image:     "x:1",
		Run:       "train",
		Framework: Framework{Name: "pytorch"},
		Worker: &Worker{Replicas: 1},
		Lifecycle: Lifecycle{SuccessPolicy: "ChiefWorker"},
	}
	err := Validate(task)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "success_policy is only valid for tensorflow")
}

func TestValidateLifecycleDurations(t *testing.T) {
	tests := []struct {
		name    string
		life    Lifecycle
		wantErr string
	}{
		{"valid active_deadline", Lifecycle{ActiveDeadline: "30s"}, ""},
		{"valid ttl", Lifecycle{TTLAfterFinished: "1h"}, ""},
		{"invalid active_deadline", Lifecycle{ActiveDeadline: "abc"}, "invalid active_deadline"},
		{"invalid ttl", Lifecycle{TTLAfterFinished: "xyz"}, "invalid ttl_after_finished"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &Task{
				Name:      "t",
				Image:     "x:1",
				Run:       "train",
				Framework: Framework{Name: "pytorch"},
				Worker:    &Worker{Replicas: 1},
				Lifecycle: tt.life,
			}
			err := Validate(task)
			if tt.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			}
		})
	}
}

func TestEffectiveShell(t *testing.T) {
	tests := []struct {
		shell    string
		expected string
	}{
		{"", "/bin/sh"},
		{"/bin/bash", "/bin/bash"},
		{"/bin/zsh", "/bin/zsh"},
		{"bash", "bash"},
		{"sh", "sh"},
		{"  ", "/bin/sh"},
	}
	for _, tt := range tests {
		task := &Task{Shell: tt.shell}
		assert.Equal(t, tt.expected, task.EffectiveShell(), "shell=%q", tt.shell)
	}
}

func TestSyncParse(t *testing.T) {
	data := []byte(`
name: t
framework:
  name: pytorch
image: x:1
run: train
worker:
  replicas: 1
storages:
  - name: code
    mount_path: /workspace
    tmp: 5Gi
  - name: dataset
    mount_path: /dataset
    tmp: 5Gi
sync:
  - git: https://github.com/org/repo.git
    branch: main
    local_path: /workspace
    mounts:
    - name: code
      mount_path: /workspace
  - rsync: /local/data
    local_path: /dataset
    mounts:
    - name: dataset
      mount_path: /dataset
`)
	task, err := LoadFromBytes(data)
	require.NoError(t, err)
	require.Len(t, task.Sync, 2)
	assert.Equal(t, "https://github.com/org/repo.git", task.Sync[0].Git)
	assert.Equal(t, "main", task.Sync[0].Branch)
	assert.Equal(t, "/local/data", task.Sync[1].Rsync)
}

func TestSyncParseWithImage(t *testing.T) {
	data := []byte(`
name: t
framework:
  name: pytorch
image: x:1
run: train
worker:
  replicas: 1
sync:
  - git: https://github.com/org/repo.git
    image: custom/git-sync:v4
    local_path: /workspace
  - rsync: server::data
    image: custom/rsync:2.0
    local_path: /rsync-data
  - hdfs: hdfs://namenode/data
    image: custom/hadoop:3.4
    local_path: /data
`)
	task, err := LoadFromBytes(data)
	require.NoError(t, err)
	require.Len(t, task.Sync, 3)
	assert.Equal(t, "custom/git-sync:v4", task.Sync[0].Image)
	assert.Equal(t, "/workspace", task.Sync[0].LocalPath)
	assert.Equal(t, "custom/rsync:2.0", task.Sync[1].Image)
	assert.Equal(t, "/rsync-data", task.Sync[1].LocalPath)
	assert.Equal(t, "custom/hadoop:3.4", task.Sync[2].Image)
	assert.Equal(t, "/data", task.Sync[2].LocalPath)
}

func TestInitContainerParse(t *testing.T) {
	data := []byte(`
name: t
framework:
  name: pytorch
image: x:1
run: train
worker:
  replicas: 1
init:
  - name: download-model
    image: busybox
    run: wget -O /data/model.bin https://example.com/model.bin
    shell: /bin/bash
`)
	task, err := LoadFromBytes(data)
	require.NoError(t, err)
	require.Len(t, task.Init, 1)
	assert.Equal(t, "download-model", task.Init[0].Name)
	assert.Equal(t, "busybox", task.Init[0].Image)
	assert.Equal(t, "/bin/bash", task.Init[0].Shell)
}

func TestFrameworkConfigMPIV2beta1(t *testing.T) {
	data := []byte(`
name: t
framework:
  name: mpi
  options:
    slots_per_worker: 4
    mpi_implementation: Intel
    launcher_creation_policy: WaitForWorkersReady
    ssh_auth_mount_path: /custom/.ssh
    run_launcher_as_worker: true
image: x:1
run: train
worker:
  replicas: 4
`)
	task, err := LoadFromBytes(data)
	require.NoError(t, err)
	assert.Equal(t, 4, task.Framework.Options.SlotsPerWorker)
	assert.Equal(t, "Intel", task.Framework.Options.MPIImplementation)
	assert.Equal(t, "WaitForWorkersReady", task.Framework.Options.LauncherCreationPolicy)
	assert.Equal(t, "/custom/.ssh", task.Framework.Options.SSHAuthMountPath)
	assert.True(t, task.Framework.Options.RunLauncherAsWorker)
}

func TestLifecycleSuspendManagedBy(t *testing.T) {
	data := []byte(`
name: t
framework:
  name: mpi
image: x:1
run: train
worker:
  replicas: 1
lifecycle:
  suspend: true
  managed_by: kueue.x-k8s.io/multikueue
`)
	task, err := LoadFromBytes(data)
	require.NoError(t, err)
	require.NotNil(t, task.Lifecycle.Suspend)
	assert.True(t, *task.Lifecycle.Suspend)
	assert.Equal(t, "kueue.x-k8s.io/multikueue", task.Lifecycle.ManagedBy)
}

func TestValidateMPIImplementation(t *testing.T) {
	task := &Task{
		Name:      "t",
		Image:     "x:1",
		Run:       "train",
		Framework: Framework{Name: "mpi", Options: FrameworkConfig{MPIImplementation: "Unknown"}},
		Worker: &Worker{Replicas: 1},
	}
	err := Validate(task)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid mpi_implementation")
}

func TestValidateMPILauncherCreationPolicy(t *testing.T) {
	task := &Task{
		Name:      "t",
		Image:     "x:1",
		Run:       "train",
		Framework: Framework{Name: "mpi", Options: FrameworkConfig{LauncherCreationPolicy: "Invalid"}},
		Worker: &Worker{Replicas: 1},
	}
	err := Validate(task)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid launcher_creation_policy")
}

func TestValidateMPIValidValues(t *testing.T) {
	task := &Task{
		Name:  "t",
		Image: "x:1",
		Run:   "train",
		Framework: Framework{Name: "mpi", Options: FrameworkConfig{
			MPIImplementation:      "Intel",
			LauncherCreationPolicy: "WaitForWorkersReady",
		}},
		Worker: &Worker{Replicas: 1},
	}
	err := Validate(task)
	require.NoError(t, err)
}

func TestValidateAffinityRulesRequiresTarget(t *testing.T) {
	task := &Task{
		Name:      "t",
		Image:     "x:1",
		Run:       "train",
		Framework: Framework{Name: "pytorch"},
		Worker: &Worker{Replicas: 1},
		Scheduling: Scheduling{
			Affinity: &Affinity{
				Rules: []AffinityRule{
					{TopologyKey: "kubernetes.io/hostname"},
				},
			},
		},
	}
	err := Validate(task)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "affinity.target is required when affinity.rules is specified")
}

func TestValidateAffinityInvalidTarget(t *testing.T) {
	task := &Task{
		Name:      "t",
		Image:     "x:1",
		Run:       "train",
		Framework: Framework{Name: "pytorch"},
		Worker: &Worker{Replicas: 1},
		Scheduling: Scheduling{
			Affinity: &Affinity{
				Target: "invalid",
				Rules:  []AffinityRule{{TopologyKey: "kubernetes.io/hostname"}},
			},
		},
	}
	err := Validate(task)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "affinity.target must be 'pod' or 'node'")
}

func TestValidateAffinityPolicyWithNodeRequiresRules(t *testing.T) {
	task := &Task{
		Name:      "t",
		Image:     "x:1",
		Run:       "train",
		Framework: Framework{Name: "pytorch"},
		Worker: &Worker{Replicas: 1},
		Scheduling: Scheduling{
			Affinity: &Affinity{
				Target: "node",
				Policy: "spread",
			},
		},
	}
	err := Validate(task)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "affinity.policy with target: node requires rules")
}

func TestValidateAffinityInvalidPolicy(t *testing.T) {
	task := &Task{
		Name:      "t",
		Image:     "x:1",
		Run:       "train",
		Framework: Framework{Name: "pytorch"},
		Worker: &Worker{Replicas: 1},
		Scheduling: Scheduling{
			Affinity: &Affinity{
				Target: "pod",
				Policy: "invalid",
			},
		},
	}
	err := Validate(task)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "affinity.policy must be 'spread', 'binpack', or 'none'")
}

func TestValidateAffinityInvalidConstraint(t *testing.T) {
	task := &Task{
		Name:      "t",
		Image:     "x:1",
		Run:       "train",
		Framework: Framework{Name: "pytorch"},
		Worker: &Worker{Replicas: 1},
		Scheduling: Scheduling{
			Affinity: &Affinity{
				Target:     "pod",
				Constraint: "invalid",
			},
		},
	}
	err := Validate(task)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "affinity.constraint must be 'preferred' or 'required'")
}

func TestValidateAffinityValidConfigurations(t *testing.T) {
	// Valid: rules with pod target
	task := &Task{
		Name:      "t",
		Image:     "x:1",
		Run:       "train",
		Framework: Framework{Name: "pytorch"},
		Worker: &Worker{Replicas: 1},
		Scheduling: Scheduling{
			Affinity: &Affinity{
				Target: "pod",
				Policy: "spread",
				Rules:  []AffinityRule{{TopologyKey: "kubernetes.io/hostname"}},
			},
		},
	}
	err := Validate(task)
	require.NoError(t, err)

	// Valid: rules with node target
	task.Scheduling.Affinity = &Affinity{
		Target: "node",
		Policy: "binpack",
		Rules: []AffinityRule{
			{MatchExpressions: []MatchExpression{{Key: "gpu", Operator: "In", Values: []string{"A100"}}}},
		},
	}
	err = Validate(task)
	require.NoError(t, err)
}

func TestValidateSyncMustSpecifySource(t *testing.T) {
	task := &Task{
		Name:      "t",
		Image:     "x:1",
		Run:       "train",
		Framework: Framework{Name: "pytorch"},
		Worker: &Worker{Replicas: 1},
		Sync:      []SyncEntry{{Branch: "main"}}, // no git, rsync, or hdfs
	}
	err := Validate(task)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must specify git, rsync, or hdfs")
}

func TestValidateSyncOnlyOneSource(t *testing.T) {
	task := &Task{
		Name:      "t",
		Image:     "x:1",
		Run:       "train",
		Framework: Framework{Name: "pytorch"},
		Worker: &Worker{Replicas: 1},
		Sync:      []SyncEntry{{Git: "https://github.com/example/repo.git", Rsync: "server::path"}},
	}
	err := Validate(task)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "can only specify one of git, rsync, or hdfs")
}

func TestValidateSyncHDFSValid(t *testing.T) {
	task := &Task{
		Name:      "t",
		Image:     "x:1",
		Run:       "train",
		Framework: Framework{Name: "pytorch"},
		Worker:    &Worker{Replicas: 1},
		Sync:      []SyncEntry{{HDFS: "hdfs://path", LocalPath: "/data"}},
	}
	err := Validate(task)
	require.NoError(t, err)
}

func TestValidateSyncValidConfigurations(t *testing.T) {
	// Valid: git only with local_path
	task := &Task{
		Name:      "t",
		Image:     "x:1",
		Run:       "train",
		Framework: Framework{Name: "pytorch"},
		Worker:    &Worker{Replicas: 1},
		Sync:      []SyncEntry{{Git: "https://github.com/example/repo.git", Branch: "main", LocalPath: "/code"}},
	}
	err := Validate(task)
	require.NoError(t, err)

	// Valid: rsync only with local_path
	task.Sync = []SyncEntry{{Rsync: "server::path", LocalPath: "/data"}}
	err = Validate(task)
	require.NoError(t, err)

	// Valid: hdfs only with local_path
	task.Sync = []SyncEntry{{HDFS: "hdfs://namenode/data", LocalPath: "/data"}}
	err = Validate(task)
	require.NoError(t, err)

	// Valid: git with mounts and local_path, mount name references existing storage
	task.Storages = []Storage{{Name: "code", MountPath: "/workspace", Tmp: "5Gi"}}
	task.Sync = []SyncEntry{{Git: "https://github.com/example/repo.git", LocalPath: "/workspace", Mounts: []Mount{{Name: "code"}}}}
	err = Validate(task)
	require.NoError(t, err)

	// Valid: git with both mounts and local_path (mounts reference storage by name)
	task.Sync = []SyncEntry{{Git: "https://github.com/example/repo.git", LocalPath: "/fallback", Mounts: []Mount{{Name: "code"}}}}
	err = Validate(task)
	require.NoError(t, err)
}

func TestValidate_SyncRequiresLocalPath(t *testing.T) {
	tt := &Task{
		Name:      "test",
		Image:     "busybox",
		Run:       "train",
		Framework: Framework{Name: "pytorch"},
		Worker:    &Worker{Replicas: 1},
		Sync: []SyncEntry{
			{Git: "https://github.com/org/repo.git"},
		},
	}
	err := Validate(tt)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "local_path")
}

func TestValidate_SyncMountReferencesStorage(t *testing.T) {
	tt := &Task{
		Name:      "test",
		Image:     "busybox",
		Run:       "train",
		Framework: Framework{Name: "pytorch"},
		Worker:    &Worker{Replicas: 1},
		Storages:  []Storage{{Name: "code", MountPath: "/code", Tmp: "5Gi"}},
		Sync: []SyncEntry{
			{Git: "https://github.com/org/repo.git", LocalPath: "/code", Mounts: []Mount{{Name: "code"}}},
		},
	}
	err := Validate(tt)
	assert.NoError(t, err)
}

func TestValidate_SyncMountNotFound(t *testing.T) {
	tt := &Task{
		Name:      "test",
		Image:     "busybox",
		Run:       "train",
		Framework: Framework{Name: "pytorch"},
		Worker:    &Worker{Replicas: 1},
		Sync: []SyncEntry{
			{Git: "https://github.com/org/repo.git", LocalPath: "/code", Mounts: []Mount{{Name: "nonexistent"}}},
		},
	}
	err := Validate(tt)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nonexistent")
	assert.Contains(t, err.Error(), "storages")
}

func TestValidate_SyncWithLocalPathOnly(t *testing.T) {
	// Valid: sync without mounts (data lands in ephemeral storage — warning at build time, not error)
	tt := &Task{
		Name:      "test",
		Image:     "busybox",
		Run:       "train",
		Framework: Framework{Name: "pytorch"},
		Worker:    &Worker{Replicas: 1},
		Sync: []SyncEntry{
			{Git: "https://github.com/org/repo.git", LocalPath: "/code"},
		},
	}
	err := Validate(tt)
	assert.NoError(t, err)
}

func TestValidate_SyncMountPathRequired(t *testing.T) {
	tt := &Task{
		Name:      "test",
		Image:     "busybox",
		Run:       "train",
		Framework: Framework{Name: "pytorch"},
		Worker:    &Worker{Replicas: 1},
		Storages:  []Storage{{Name: "code", Tmp: "5Gi"}}, // no mount_path on storage
		Sync: []SyncEntry{
			{Git: "https://github.com/org/repo.git", LocalPath: "/code", Mounts: []Mount{{Name: "code"}}},
		},
	}
	err := Validate(tt)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mount_path")
}

func TestRoleConfigParsing(t *testing.T) {
	data := []byte(`
name: test
framework:
  name: pytorch
image: pytorch:2.1
run: python train.py
worker:
  replicas: 4
  resources:
    nvidia.com/gpu: 1
master:
  resources:
    nvidia.com/gpu: 2
  envs:
    ROLE: master
`)
	task, err := LoadFromBytes(data)
	require.NoError(t, err)
	require.NotNil(t, task.Master)
	assert.Equal(t, "2", task.Master.Resources["nvidia.com/gpu"])
	assert.Equal(t, "master", task.Master.Envs["ROLE"].Value)
}

func TestAllRoleFieldsParsing(t *testing.T) {
	data := []byte(`
name: test
framework:
  name: tensorflow
image: tf:2.15
run: python train.py
worker:
  replicas: 4
chief:
  resources:
    nvidia.com/gpu: 2
ps:
  replicas: 2
  resources:
    cpu: "4"
    memory: 16Gi
evaluator:
  resources:
    cpu: "2"
`)
	task, err := LoadFromBytes(data)
	require.NoError(t, err)
	require.NotNil(t, task.Chief)
	assert.Equal(t, "2", task.Chief.Resources["nvidia.com/gpu"])
	require.NotNil(t, task.PS)
	assert.Equal(t, 2, task.PS.Replicas)
	assert.Equal(t, "4", task.PS.Resources["cpu"])
	require.NotNil(t, task.Evaluator)
	assert.Equal(t, "2", task.Evaluator.Resources["cpu"])
}

func TestValidateRoleFrameworkSpecific(t *testing.T) {
	// Master only valid for PyTorch
	task := &Task{
		Name:      "t",
		Image:     "x:1",
		Run:       "train",
		Framework: Framework{Name: "tensorflow"},
		Worker: &Worker{Replicas: 1},
		Master:    &RoleConfig{},
	}
	err := Validate(task)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "master role is only valid for pytorch")

	// Chief only valid for TF
	task = &Task{
		Name:      "t",
		Image:     "x:1",
		Run:       "train",
		Framework: Framework{Name: "pytorch"},
		Worker: &Worker{Replicas: 1},
		Chief:     &RoleConfig{},
	}
	err = Validate(task)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "chief role is only valid for tensorflow")

	// PS only valid for TF
	task = &Task{
		Name:      "t",
		Image:     "x:1",
		Run:       "train",
		Framework: Framework{Name: "pytorch"},
		Worker: &Worker{Replicas: 1},
		PS:        &RoleConfig{Replicas: 1},
	}
	err = Validate(task)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ps role is only valid for tensorflow")

	// Evaluator only valid for TF
	task = &Task{
		Name:      "t",
		Image:     "x:1",
		Run:       "train",
		Framework: Framework{Name: "pytorch"},
		Worker: &Worker{Replicas: 1},
		Evaluator: &RoleConfig{},
	}
	err = Validate(task)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "evaluator role is only valid for tensorflow")

	// Launcher only valid for MPI
	task = &Task{
		Name:      "t",
		Image:     "x:1",
		Run:       "train",
		Framework: Framework{Name: "pytorch"},
		Worker: &Worker{Replicas: 1},
		Launcher:  &RoleConfig{},
	}
	err = Validate(task)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "launcher role is only valid for mpi")
}

func TestValidateRoleFrameworkSpecificValid(t *testing.T) {
	// Master with pytorch should pass
	task := &Task{
		Name:      "t",
		Image:     "x:1",
		Run:       "train",
		Framework: Framework{Name: "pytorch"},
		Worker: &Worker{Replicas: 1},
		Master:    &RoleConfig{},
	}
	err := Validate(task)
	require.NoError(t, err)

	// Chief with tensorflow should pass
	task = &Task{
		Name:      "t",
		Image:     "x:1",
		Run:       "train",
		Framework: Framework{Name: "tensorflow"},
		Worker: &Worker{Replicas: 1},
		Chief:     &RoleConfig{},
	}
	err = Validate(task)
	require.NoError(t, err)

	// Launcher with mpi should pass
	task = &Task{
		Name:      "t",
		Image:     "x:1",
		Run:       "train",
		Framework: Framework{Name: "mpi"},
		Worker: &Worker{Replicas: 1},
		Launcher:  &RoleConfig{},
	}
	err = Validate(task)
	require.NoError(t, err)
}

func TestValidateConstrainedRolesReplicas(t *testing.T) {
	tests := []struct {
		name     string
		task     *Task
		errContains string
	}{
		{
			name: "master with replicas > 1",
			task: &Task{
				Name:      "t",
				Image:     "x:1",
				Run:       "train",
				Framework: Framework{Name: "pytorch"},
				Worker: &Worker{Replicas: 1},
				Master:    &RoleConfig{Replicas: 2},
			},
			errContains: "master role is constrained to replicas=1",
		},
		{
			name: "chief with replicas > 1",
			task: &Task{
				Name:      "t",
				Image:     "x:1",
				Run:       "train",
				Framework: Framework{Name: "tensorflow"},
				Worker: &Worker{Replicas: 1},
				Chief:     &RoleConfig{Replicas: 3},
			},
			errContains: "chief role is constrained to replicas=1",
		},
		{
			name: "launcher with replicas > 1",
			task: &Task{
				Name:      "t",
				Image:     "x:1",
				Run:       "train",
				Framework: Framework{Name: "mpi"},
				Worker: &Worker{Replicas: 1},
				Launcher:  &RoleConfig{Replicas: 2},
			},
			errContains: "launcher role is constrained to replicas=1",
		},
		{
			name: "evaluator with replicas > 1",
			task: &Task{
				Name:      "t",
				Image:     "x:1",
				Run:       "train",
				Framework: Framework{Name: "tensorflow"},
				Worker: &Worker{Replicas: 1},
				Evaluator: &RoleConfig{Replicas: 5},
			},
			errContains: "evaluator role is constrained to replicas=1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.task)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errContains)
		})
	}
}

func TestValidateConstrainedRolesReplicasZeroAllowed(t *testing.T) {
	// replicas=0 means unset (Go zero value), should be fine for constrained roles
	task := &Task{
		Name:      "t",
		Image:     "x:1",
		Run:       "train",
		Framework: Framework{Name: "pytorch"},
		Worker: &Worker{Replicas: 1},
		Master:    &RoleConfig{Replicas: 0},
	}
	err := Validate(task)
	require.NoError(t, err)

	// replicas=1 is fine for constrained roles
	task.Master.Replicas = 1
	err = Validate(task)
	require.NoError(t, err)
}

func TestValidatePSReplicas(t *testing.T) {
	// PS with replicas < 1 should fail
	task := &Task{
		Name:      "t",
		Image:     "x:1",
		Run:       "train",
		Framework: Framework{Name: "tensorflow"},
		Worker: &Worker{Replicas: 1},
		PS:        &RoleConfig{Replicas: 0},
	}
	err := Validate(task)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ps.replicas must be > 0")

	// PS with replicas >= 1 should pass
	task.PS.Replicas = 2
	err = Validate(task)
	require.NoError(t, err)
}

func TestValidatePSReplicasNegative(t *testing.T) {
	task := &Task{
		Name:      "t",
		Image:     "x:1",
		Run:       "train",
		Framework: Framework{Name: "tensorflow"},
		Worker: &Worker{Replicas: 1},
		PS:        &RoleConfig{Replicas: -1},
	}
	err := Validate(task)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ps.replicas must be > 0")
}

func TestMPIRoleParsing(t *testing.T) {
	data := []byte(`
name: test
framework:
  name: mpi
image: mpi:latest
run: mpirun ./train
worker:
  replicas: 4
launcher:
  resources:
    cpu: "2"
    memory: 4Gi
`)
	task, err := LoadFromBytes(data)
	require.NoError(t, err)
	require.NotNil(t, task.Launcher)
	assert.Equal(t, "2", task.Launcher.Resources["cpu"])
	assert.Equal(t, "4Gi", task.Launcher.Resources["memory"])
}

func TestValidateWorkerNilPyTorchWithMaster(t *testing.T) {
	task := &Task{
		Name:      "t",
		Image:     "x:1",
		Run:       "train",
		Framework: Framework{Name: "pytorch"},
		Worker:    nil,
		Master:    &RoleConfig{Resources: Resources{"cpu": "1"}},
	}
	err := Validate(task)
	require.NoError(t, err)
}

func TestValidateWorkerNilPyTorchNoMaster(t *testing.T) {
	task := &Task{
		Name:      "t",
		Image:     "x:1",
		Run:       "train",
		Framework: Framework{Name: "pytorch"},
		Worker:    nil,
		Master:    nil,
	}
	err := Validate(task)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pytorch requires worker or master")
}

func TestValidateWorkerNilTensorFlow(t *testing.T) {
	task := &Task{
		Name:      "t",
		Image:     "x:1",
		Run:       "train",
		Framework: Framework{Name: "tensorflow"},
		Worker:    nil,
	}
	err := Validate(task)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "worker is required for tensorflow")
}

func TestValidateWorkerNilMPI(t *testing.T) {
	task := &Task{
		Name:      "t",
		Image:     "x:1",
		Run:       "train",
		Framework: Framework{Name: "mpi"},
		Worker:    nil,
	}
	err := Validate(task)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "worker is required for mpi")
}

func TestStorage_Validate(t *testing.T) {
	tests := []struct {
		name    string
		storage Storage
		wantErr string
	}{
		{
			name:    "valid configmap with key",
			storage: Storage{Name: "cfg", ConfigMap: "foo", Key: "file.txt", MountPath: "/etc"},
			wantErr: "",
		},
		{
			name:    "valid configmap without key (full mount)",
			storage: Storage{Name: "cfg", ConfigMap: "foo", MountPath: "/etc"},
			wantErr: "",
		},
		{
			name:    "valid secret with key",
			storage: Storage{Name: "sec", Secret: "bar", Key: "token", MountPath: "/token"},
			wantErr: "",
		},
		{
			name:    "valid secret without key (full mount)",
			storage: Storage{Name: "sec", Secret: "bar", MountPath: "/sec"},
			wantErr: "",
		},
		{
			name:    "valid pvc",
			storage: Storage{Name: "data", PVC: "my-pvc", MountPath: "/data"},
			wantErr: "",
		},
		{
			name:    "empty name",
			storage: Storage{PVC: "my-pvc", MountPath: "/data"},
			wantErr: "storage name must not be empty",
		},
		{
			name:    "empty mount_path",
			storage: Storage{Name: "data", PVC: "my-pvc"},
			wantErr: "mountPath must not be empty",
		},
		{
			name:    "no storage type",
			storage: Storage{Name: "empty", MountPath: "/data"},
			wantErr: "must specify exactly one of pvc, shm, tmp, hostpath, configmap, or secret",
		},
		{
			name:    "multiple storage types",
			storage: Storage{Name: "multi", PVC: "pvc", ConfigMap: "cm", MountPath: "/data"},
			wantErr: "cannot specify multiple storage types",
		},
		{
			name:    "key without configmap/secret",
			storage: Storage{Name: "bad", PVC: "pvc", Key: "file.txt", MountPath: "/data"},
			wantErr: "key can only be used with configmap or secret",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.storage.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			} else {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("expected error containing %q, got %v", tt.wantErr, err)
				}
			}
		})
	}
}

// TestValidateRoleCombinations covers valid and invalid role+framework combinations
// that exercise the role validation logic in Validate().
func TestValidateRoleCombinations(t *testing.T) {
	tests := []struct {
		name        string
		task        *Task
		wantErr     bool
		errContains string
	}{
		// --- PyTorch valid combinations ---
		{
			name: "pytorch: worker only (no master)",
			task: &Task{
				Name: "t", Image: "x:1", Run: "train",
				Framework: Framework{Name: "pytorch"},
				Worker:    &Worker{Replicas: 2},
			},
			wantErr: false,
		},
		{
			name: "pytorch: master only (no worker)",
			task: &Task{
				Name: "t", Image: "x:1", Run: "train",
				Framework: Framework{Name: "pytorch"},
				Master:    &RoleConfig{Replicas: 1, Resources: Resources{"nvidia.com/gpu": "1"}},
			},
			wantErr: false,
		},
		{
			name: "pytorch: worker + master",
			task: &Task{
				Name: "t", Image: "x:1", Run: "train",
				Framework: Framework{Name: "pytorch"},
				Worker:    &Worker{Replicas: 4},
				Master:    &RoleConfig{Resources: Resources{"nvidia.com/gpu": "8"}},
			},
			wantErr: false,
		},

		// --- TensorFlow valid combinations ---
		{
			name: "tensorflow: worker only (no optional roles)",
			task: &Task{
				Name: "t", Image: "x:1", Run: "train",
				Framework: Framework{Name: "tensorflow"},
				Worker:    &Worker{Replicas: 1},
			},
			wantErr: false,
		},
		{
			name: "tensorflow: worker + chief",
			task: &Task{
				Name: "t", Image: "x:1", Run: "train",
				Framework: Framework{Name: "tensorflow"},
				Worker:    &Worker{Replicas: 2},
				Chief:     &RoleConfig{Resources: Resources{"nvidia.com/gpu": "1"}},
			},
			wantErr: false,
		},
		{
			name: "tensorflow: worker + ps",
			task: &Task{
				Name: "t", Image: "x:1", Run: "train",
				Framework: Framework{Name: "tensorflow"},
				Worker:    &Worker{Replicas: 2},
				PS:        &RoleConfig{Replicas: 2},
			},
			wantErr: false,
		},
		{
			name: "tensorflow: worker + all roles",
			task: &Task{
				Name: "t", Image: "x:1", Run: "train",
				Framework: Framework{Name: "tensorflow"},
				Worker:    &Worker{Replicas: 4},
				Chief:     &RoleConfig{},
				PS:        &RoleConfig{Replicas: 3},
				Evaluator: &RoleConfig{},
			},
			wantErr: false,
		},

		// --- MPI valid combinations ---
		{
			name: "mpi: worker only (no launcher)",
			task: &Task{
				Name: "t", Image: "x:1", Run: "train",
				Framework: Framework{Name: "mpi"},
				Worker:    &Worker{Replicas: 4},
			},
			wantErr: false,
		},
		{
			name: "mpi: worker + launcher",
			task: &Task{
				Name: "t", Image: "x:1", Run: "train",
				Framework: Framework{Name: "mpi"},
				Worker:    &Worker{Replicas: 4},
				Launcher:  &RoleConfig{Resources: Resources{"cpu": "2"}},
			},
			wantErr: false,
		},

		// --- Horovod/DeepSpeed with launcher ---
		{
			name: "horovod: worker + launcher",
			task: &Task{
				Name: "t", Image: "x:1", Run: "train",
				Framework: Framework{Name: "horovod"},
				Worker:    &Worker{Replicas: 2},
				Launcher:  &RoleConfig{},
			},
			wantErr: false,
		},
		{
			name: "deepspeed: worker + launcher",
			task: &Task{
				Name: "t", Image: "x:1", Run: "train",
				Framework: Framework{Name: "deepspeed"},
				Worker:    &Worker{Replicas: 2},
				Launcher:  &RoleConfig{},
			},
			wantErr: false,
		},

		// --- Invalid: role with wrong framework ---
		{
			name: "invalid: chief with pytorch",
			task: &Task{
				Name: "t", Image: "x:1", Run: "train",
				Framework: Framework{Name: "pytorch"},
				Worker:    &Worker{Replicas: 1},
				Chief:     &RoleConfig{},
			},
			wantErr:     true,
			errContains: "chief role is only valid for tensorflow",
		},
		{
			name: "invalid: master with tensorflow",
			task: &Task{
				Name: "t", Image: "x:1", Run: "train",
				Framework: Framework{Name: "tensorflow"},
				Worker:    &Worker{Replicas: 1},
				Master:    &RoleConfig{},
			},
			wantErr:     true,
			errContains: "master role is only valid for pytorch",
		},
		{
			name: "invalid: ps with mpi",
			task: &Task{
				Name: "t", Image: "x:1", Run: "train",
				Framework: Framework{Name: "mpi"},
				Worker:    &Worker{Replicas: 1},
				PS:        &RoleConfig{Replicas: 1},
			},
			wantErr:     true,
			errContains: "ps role is only valid for tensorflow",
		},
		{
			name: "invalid: evaluator with mpi",
			task: &Task{
				Name: "t", Image: "x:1", Run: "train",
				Framework: Framework{Name: "mpi"},
				Worker:    &Worker{Replicas: 1},
				Evaluator: &RoleConfig{},
			},
			wantErr:     true,
			errContains: "evaluator role is only valid for tensorflow",
		},
		{
			name: "invalid: launcher with pytorch",
			task: &Task{
				Name: "t", Image: "x:1", Run: "train",
				Framework: Framework{Name: "pytorch"},
				Worker:    &Worker{Replicas: 1},
				Launcher:  &RoleConfig{},
			},
			wantErr:     true,
			errContains: "launcher role is only valid for mpi",
		},
		{
			name: "invalid: launcher with tensorflow",
			task: &Task{
				Name: "t", Image: "x:1", Run: "train",
				Framework: Framework{Name: "tensorflow"},
				Worker:    &Worker{Replicas: 1},
				Launcher:  &RoleConfig{},
			},
			wantErr:     true,
			errContains: "launcher role is only valid for mpi",
		},

		// --- Invalid: constrained role with replicas > 1 ---
		{
			name: "invalid: master replicas > 1",
			task: &Task{
				Name: "t", Image: "x:1", Run: "train",
				Framework: Framework{Name: "pytorch"},
				Worker:    &Worker{Replicas: 1},
				Master:    &RoleConfig{Replicas: 2},
			},
			wantErr:     true,
			errContains: "master role is constrained to replicas=1",
		},
		{
			name: "invalid: evaluator replicas > 1",
			task: &Task{
				Name: "t", Image: "x:1", Run: "train",
				Framework: Framework{Name: "tensorflow"},
				Worker:    &Worker{Replicas: 1},
				Evaluator: &RoleConfig{Replicas: 2},
			},
			wantErr:     true,
			errContains: "evaluator role is constrained to replicas=1",
		},

		// --- Invalid: ps with replicas = 0 ---
		{
			name: "invalid: ps replicas = 0",
			task: &Task{
				Name: "t", Image: "x:1", Run: "train",
				Framework: Framework{Name: "tensorflow"},
				Worker:    &Worker{Replicas: 1},
				PS:        &RoleConfig{Replicas: 0},
			},
			wantErr:     true,
			errContains: "ps.replicas must be > 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.task)
			if tt.wantErr {
				require.Error(t, err, "expected error containing %q", tt.errContains)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestTensorBoardConfig_MountsField(t *testing.T) {
	cfg := &TensorBoardConfig{
		Enabled: true,
		LogDir:  "/logs",
		Mounts: []Mount{
			{Name: "data", MountPath: "/custom/data"},
		},
	}
	assert.True(t, cfg.Enabled)
	assert.Len(t, cfg.Mounts, 1)
	assert.Equal(t, "data", cfg.Mounts[0].Name)
	assert.Equal(t, "/custom/data", cfg.Mounts[0].MountPath)
}

func TestTensorBoardConfig_MountsYAMLParse(t *testing.T) {
	data := []byte(`
name: test
framework:
  name: pytorch
image: pytorch:2.1
run: python train.py
worker:
  replicas: 1
storages:
  - name: dataset
    mount_path: /data
    pvc: dataset-pvc
  - name: output
    mount_path: /output
    pvc: output-pvc
logging:
  tensorboard:
    enabled: true
    logdir: /logs
    image: tensorflow/tensorflow:2.15
    mounts:
      - name: dataset
        mount_path: /data
      - name: output
        mount_path: /output
`)
	task, err := LoadFromBytes(data)
	require.NoError(t, err)
	require.NotNil(t, task.Logging.TensorBoard)
	assert.True(t, task.Logging.TensorBoard.Enabled)
	assert.Equal(t, "/logs", task.Logging.TensorBoard.LogDir)
	assert.Equal(t, "tensorflow/tensorflow:2.15", task.Logging.TensorBoard.Image)
	require.Len(t, task.Logging.TensorBoard.Mounts, 2)
	assert.Equal(t, "dataset", task.Logging.TensorBoard.Mounts[0].Name)
	assert.Equal(t, "/data", task.Logging.TensorBoard.Mounts[0].MountPath)
	assert.Equal(t, "output", task.Logging.TensorBoard.Mounts[1].Name)
	assert.Equal(t, "/output", task.Logging.TensorBoard.Mounts[1].MountPath)
}

func TestValidate_TensorBoardMountsInvalidName(t *testing.T) {
	tk := &Task{
		Name:      "test",
		Image:     "test:latest",
		Run:       "echo",
		Framework: Framework{Name: "pytorch"},
		Master:    &RoleConfig{},
		Logging: Logging{
			TensorBoard: &TensorBoardConfig{
				Enabled: true,
				Mounts: []Mount{
					{Name: "nonexistent"},
				},
			},
		},
	}
	err := Validate(tk)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "logging.tensorboard.mounts[0].name \"nonexistent\" not found in storages")
}

func TestValidate_TensorBoardMountsMissingMountPath(t *testing.T) {
	tk := &Task{
		Name:      "test",
		Image:     "test:latest",
		Run:       "echo",
		Framework: Framework{Name: "pytorch"},
		Master:    &RoleConfig{},
		Storages: []Storage{
			{Name: "data", PVC: "pvc"}, // no mount_path
		},
		Logging: Logging{
			TensorBoard: &TensorBoardConfig{
				Enabled: true,
				Mounts: []Mount{
					{Name: "data"}, // no mount_path override
				},
			},
		},
	}
	err := Validate(tk)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "logging.tensorboard.mounts[0].mount_path is required")
}

func TestValidate_TensorBoardMountsValid(t *testing.T) {
	tk := &Task{
		Name:      "test",
		Image:     "test:latest",
		Run:       "echo",
		Framework: Framework{Name: "pytorch"},
		Master:    &RoleConfig{},
		Storages: []Storage{
			{Name: "data", PVC: "pvc", MountPath: "/data"},
		},
		Logging: Logging{
			TensorBoard: &TensorBoardConfig{
				Enabled: true,
				Mounts: []Mount{
					{Name: "data", MountPath: "/custom"},
				},
			},
		},
	}
	err := Validate(tk)
	assert.NoError(t, err)
}

func TestFrameworkConfigNoAPIVersion(t *testing.T) {
	// APIVersion should not exist on FrameworkConfig.
	// This test verifies the field was removed.
	fc := FrameworkConfig{}
	fcVal := reflect.ValueOf(fc)
	for i := 0; i < fcVal.NumField(); i++ {
		assert.NotEqual(t, "APIVersion", fcVal.Type().Field(i).Name,
			"FrameworkConfig should not have APIVersion field")
	}
}

func TestPerRoleRunOverrideYAMLParse(t *testing.T) {
	data := []byte(`
name: tf-distributed
framework:
  name: tensorflow
image: tensorflow:2.15
run: python train.py --epochs 10
worker:
  replicas: 3
  run: python worker_train.py
ps:
  replicas: 2
  run: python serve_ps.py
`)
	task, err := LoadFromBytes(data)
	require.NoError(t, err)
	assert.Equal(t, "python train.py --epochs 10", task.Run)
	assert.Equal(t, "python worker_train.py", task.Worker.Run)
	assert.Equal(t, "python serve_ps.py", task.PS.Run)
}

func TestValidate_DuplicateStorageNames(t *testing.T) {
	task := &Task{
		Name:      "t",
		Image:     "x:1",
		Run:       "train",
		Framework: Framework{Name: "pytorch"},
		Worker:    &Worker{Replicas: 1},
		Storages: []Storage{
			{Name: "data", PVC: "pvc-1", MountPath: "/data1"},
			{Name: "data", PVC: "pvc-2", MountPath: "/data2"},
		},
	}
	err := Validate(task)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate storage name")
	assert.Contains(t, err.Error(), "data")
}

func TestEnvValue_ReuseResetsFields(t *testing.T) {
	var e EnvValue

	err := yaml.Unmarshal([]byte(`{secret: my-secret, key: token}`), &e)
	require.NoError(t, err)
	require.NotNil(t, e.Secret)
	assert.Equal(t, "my-secret", e.Secret.Name)

	err = yaml.Unmarshal([]byte(`"plain-value"`), &e)
	require.NoError(t, err)
	assert.Equal(t, "plain-value", e.Value)
	assert.Nil(t, e.Secret, "Secret field should be nil after reusing EnvValue for a plain string")
	assert.Nil(t, e.ConfigMap, "ConfigMap field should be nil after reusing EnvValue for a plain string")
}

func TestEnvValue_BothSecretAndConfigMap(t *testing.T) {
	var e EnvValue
	err := yaml.Unmarshal([]byte(`{secret: my-secret, configmap: my-cm, key: token}`), &e)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exactly one of 'secret' or 'configmap'")
}

func TestEnvValue_MarshalYAML_BothSecretAndConfigMap(t *testing.T) {
	e := EnvValue{
		Secret:    &EnvFrom{Name: "my-secret", Key: "token"},
		ConfigMap: &EnvFrom{Name: "my-cm", Key: "host"},
	}
	_, err := e.MarshalYAML()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot set both secret and configmap")
}

func TestEnvValue_EmptyKeyForSecretAndConfigMap(t *testing.T) {
	tests := []struct {
		name string
		yaml string
		err  string
	}{
		{"secret without key", `{secret: my-secret}`, "key must not be empty for secret reference"},
		{"configmap without key", `{configmap: my-cm}`, "key must not be empty for configmap reference"},
		{"secret with empty key", `{secret: my-secret, key: ""}`, "key must not be empty for secret reference"},
		{"configmap with empty key", `{configmap: my-cm, key: ""}`, "key must not be empty for configmap reference"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var e EnvValue
			err := yaml.Unmarshal([]byte(tt.yaml), &e)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.err)
		})
	}
}
