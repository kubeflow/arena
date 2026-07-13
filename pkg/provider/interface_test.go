package provider

import (
	"bytes"
	"io"
	"os"
	"testing"
	"time"

	"github.com/kubeflow/arena/pkg/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildSchedulingPolicy_WithQueue(t *testing.T) {
	tt := &task.Task{
		Scheduling: task.Scheduling{
			Gang:              task.GangConfig{Enabled: false},
			Queue:             "high-priority",
			PriorityClassName: "",
		},
	}
	result := buildSchedulingPolicy(tt)
	if result == nil {
		t.Fatal("expected non-nil result when queue is set")
	}
	if result["queue"] != "high-priority" {
		t.Errorf("expected queue=high-priority, got %v", result["queue"])
	}
}

func TestBuildSchedulingPolicy_WithPriorityClass(t *testing.T) {
	tt := &task.Task{
		Scheduling: task.Scheduling{
			Gang:              task.GangConfig{Enabled: false},
			Queue:             "",
			PriorityClassName: "premium",
		},
	}
	result := buildSchedulingPolicy(tt)
	if result == nil {
		t.Fatal("expected non-nil result when priorityClassName is set")
	}
	if result["priorityClass"] != "premium" {
		t.Errorf("expected priorityClass=premium, got %v", result["priorityClass"])
	}
}

func TestBuildSchedulingPolicy_WithBoth(t *testing.T) {
	tt := &task.Task{
		Scheduling: task.Scheduling{
			Gang:              task.GangConfig{Enabled: true},
			Queue:             "high-priority",
			PriorityClassName: "premium",
		},
	}
	result := buildSchedulingPolicy(tt)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result["queue"] != "high-priority" {
		t.Errorf("expected queue=high-priority, got %v", result["queue"])
	}
	if result["priorityClass"] != "premium" {
		t.Errorf("expected priorityClass=premium, got %v", result["priorityClass"])
	}
}

func TestBuildSchedulingPolicyGang(t *testing.T) {
	// Gang enabled, no roles: minAvailable falls back to 1 (safety minimum)
	tt := &task.Task{
		Scheduling: task.Scheduling{
			Gang: task.GangConfig{Enabled: true},
		},
	}

	sp := buildSchedulingPolicy(tt)
	require.NotNil(t, sp)

	minAvail, ok := sp["minAvailable"]
	assert.True(t, ok, "gang=true should set minAvailable")
	assert.Equal(t, int64(1), minAvail, "minAvailable should be 1 when no roles are defined (safety minimum)")

	// Gang enabled with roles: minAvailable equals total replicas
	ttWithRoles := &task.Task{
		Scheduling: task.Scheduling{
			Gang: task.GangConfig{Enabled: true},
		},
		Worker: &task.Worker{Replicas: 3},
		Master: &task.RoleConfig{Replicas: 1},
	}

	sp = buildSchedulingPolicy(ttWithRoles)
	require.NotNil(t, sp)

	minAvail, ok = sp["minAvailable"]
	assert.True(t, ok, "gang=true should set minAvailable")
	assert.Equal(t, int64(4), minAvail, "minAvailable should equal total replicas (3 workers + 1 master)")
}

func TestBuildSchedulingPolicyGangFalse(t *testing.T) {
	tt := &task.Task{
		Scheduling: task.Scheduling{
			Gang: task.GangConfig{Enabled: false},
		},
	}

	sp := buildSchedulingPolicy(tt)
	assert.Nil(t, sp, "gang=false with no queue/priorityClass should return nil")
}

func TestBuildInitContainers_Empty(t *testing.T) {
	tt := &task.Task{
		Init: []task.InitContainer{},
		Sync: []task.SyncEntry{},
	}
	result := buildInitContainers(tt)
	if result != nil {
		t.Errorf("expected nil for empty init and sync, got %v", result)
	}
}

func TestBuildInitContainers_SingleUserDefined(t *testing.T) {
	tt := &task.Task{
		Init: []task.InitContainer{
			{Name: "setup", Image: "busybox", Run: "mkdir -p /data", Shell: "/bin/bash"},
		},
		Sync: []task.SyncEntry{},
	}
	result := buildInitContainers(tt)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 init container, got %d", len(result))
	}
	if result[0]["name"] != "setup" {
		t.Errorf("expected name=setup, got %v", result[0]["name"])
	}
	if result[0]["image"] != "busybox" {
		t.Errorf("expected image=busybox, got %v", result[0]["image"])
	}
}

func TestBuildInitContainers_ShellFallback(t *testing.T) {
	// Test 1: init.Shell is set
	tt := &task.Task{
		Init: []task.InitContainer{
			{Name: "t1", Image: "x:1", Run: "cmd", Shell: "/bin/bash"},
		},
		Sync: []task.SyncEntry{},
	}
	result := buildInitContainers(tt)
	cmd := result[0]["command"].([]interface{})
	if cmd[0] != "/bin/bash" {
		t.Errorf("expected shell=/bin/bash from init, got %v", cmd[0])
	}

	// Test 2: init.Shell empty, fallback to task.Shell
	tt = &task.Task{
		Shell: "/bin/zsh",
		Init: []task.InitContainer{
			{Name: "t2", Image: "x:1", Run: "cmd"},
		},
		Sync: []task.SyncEntry{},
	}
	result = buildInitContainers(tt)
	cmd = result[0]["command"].([]interface{})
	if cmd[0] != "/bin/zsh" {
		t.Errorf("expected shell=/bin/zsh from task, got %v", cmd[0])
	}

	// Test 3: both empty, fallback to /bin/sh
	tt = &task.Task{
		Init: []task.InitContainer{
			{Name: "t3", Image: "x:1", Run: "cmd"},
		},
		Sync: []task.SyncEntry{},
	}
	result = buildInitContainers(tt)
	cmd = result[0]["command"].([]interface{})
	if cmd[0] != "/bin/sh" {
		t.Errorf("expected shell=/bin/sh default, got %v", cmd[0])
	}
}

func TestBuildInitContainers_Multiple(t *testing.T) {
	tt := &task.Task{
		Init: []task.InitContainer{
			{Name: "init1", Image: "img1:1", Run: "cmd1"},
			{Name: "init2", Image: "img2:1", Run: "cmd2"},
		},
		Sync: []task.SyncEntry{},
	}
	result := buildInitContainers(tt)
	if len(result) != 2 {
		t.Fatalf("expected 2 init containers, got %d", len(result))
	}
	if result[0]["name"] != "init1" {
		t.Errorf("expected first init name=init1, got %v", result[0]["name"])
	}
	if result[1]["name"] != "init2" {
		t.Errorf("expected second init name=init2, got %v", result[1]["name"])
	}
}

func TestExtractGitProjectName(t *testing.T) {
	tests := []struct {
		gitURL   string
		expected string
	}{
		{"https://github.com/kubeflow/training-operator.git", "training-operator"},
		{"https://github.com/example/repo", "repo"},
		{"git@github.com:user/project.git", "project"},
		{"https://gitlab.com/group/subgroup/myapp.git", "myapp"},
	}
	for _, tt := range tests {
		result := extractGitProjectName(tt.gitURL)
		if result != tt.expected {
			t.Errorf("extractGitProjectName(%q) = %q, want %q", tt.gitURL, result, tt.expected)
		}
	}
}

func TestBuildSyncInitContainers_Git(t *testing.T) {
	tt := &task.Task{
		Sync: []task.SyncEntry{
			{Git: "https://github.com/kubeflow/training-operator.git", Branch: "main", LocalPath: "/code"},
		},
	}
	result := buildSyncInitContainers(tt)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 init container, got %d", len(result))
	}
	if result[0]["name"] != "arena-sync-0" {
		t.Errorf("expected name=arena-sync-0, got %v", result[0]["name"])
	}
	envs := result[0]["env"].([]map[string]interface{})
	found := map[string]string{}
	for _, e := range envs {
		found[e["name"].(string)] = e["value"].(string)
	}
	if found["GIT_SYNC_REPO"] != "https://github.com/kubeflow/training-operator.git" {
		t.Errorf("expected GIT_SYNC_REPO, got %v", found["GIT_SYNC_REPO"])
	}
	if found["GIT_SYNC_DEST"] != "training-operator" {
		t.Errorf("expected GIT_SYNC_DEST=training-operator, got %v", found["GIT_SYNC_DEST"])
	}
	if found["GIT_SYNC_ROOT"] != "/code" {
		t.Errorf("expected GIT_SYNC_ROOT=/code, got %v", found["GIT_SYNC_ROOT"])
	}
	if found["GIT_SYNC_BRANCH"] != "main" {
		t.Errorf("expected GIT_SYNC_BRANCH=main, got %v", found["GIT_SYNC_BRANCH"])
	}
}

func TestBuildSyncInitContainers_Rsync(t *testing.T) {
	tt := &task.Task{
		Sync: []task.SyncEntry{{Rsync: "10.0.0.1::data/train.zip", LocalPath: "/data"}},
	}
	result := buildSyncInitContainers(tt)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 init container, got %d", len(result))
	}
	if result[0]["name"] != "arena-sync-0" {
		t.Errorf("expected name=arena-sync-0, got %v", result[0]["name"])
	}
	cmd := result[0]["command"].([]interface{})
	if len(cmd) != 4 {
		t.Fatalf("expected 4 command args, got %d", len(cmd))
	}
	if cmd[0] != "rsync" {
		t.Errorf("expected rsync command, got %v", cmd[0])
	}
}

func TestBuildSyncInitContainers_HDFS(t *testing.T) {
	tt := &task.Task{
		Sync: []task.SyncEntry{{HDFS: "hdfs://namenode:9000/data/model", LocalPath: "/models"}},
	}
	result := buildSyncInitContainers(tt)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 init container, got %d", len(result))
	}
	if result[0]["name"] != "arena-sync-0" {
		t.Errorf("expected name=arena-sync-0, got %v", result[0]["name"])
	}
	if result[0]["image"] != "apache/hadoop:3.5.0" {
		t.Errorf("expected default hdfs image, got %v", result[0]["image"])
	}
	cmd := result[0]["command"].([]interface{})
	if len(cmd) != 5 {
		t.Fatalf("expected 5 command args, got %d", len(cmd))
	}
	if cmd[0] != "hdfs" || cmd[1] != "dfs" || cmd[2] != "-get" {
		t.Errorf("expected hdfs dfs -get command, got %v", cmd)
	}
	if cmd[3] != "hdfs://namenode:9000/data/model" {
		t.Errorf("expected hdfs source, got %v", cmd[3])
	}
	if cmd[4] != "/models" {
		t.Errorf("expected local path /models, got %v", cmd[4])
	}
}

func TestBuildSyncInitContainers_CustomImages(t *testing.T) {
	tt := &task.Task{
		Sync: []task.SyncEntry{
			{Git: "https://github.com/example/repo.git", Image: "custom/git-sync:v4", LocalPath: "/code"},
			{Rsync: "server::data", Image: "custom/rsync:2.0", LocalPath: "/data"},
			{HDFS: "hdfs://path", Image: "custom/hadoop:3.4", LocalPath: "/hdfs"},
		},
	}
	result := buildSyncInitContainers(tt)
	if len(result) != 3 {
		t.Fatalf("expected 3 init containers, got %d", len(result))
	}
	if result[0]["image"] != "custom/git-sync:v4" {
		t.Errorf("expected custom git image, got %v", result[0]["image"])
	}
	if result[1]["image"] != "custom/rsync:2.0" {
		t.Errorf("expected custom rsync image, got %v", result[1]["image"])
	}
	if result[2]["image"] != "custom/hadoop:3.4" {
		t.Errorf("expected custom hdfs image, got %v", result[2]["image"])
	}
}

func TestBuildSyncInitContainers_MultipleEntries(t *testing.T) {
	tt := &task.Task{
		Sync: []task.SyncEntry{
			{Git: "https://github.com/org/repo1.git", LocalPath: "/code"},
			{Rsync: "10.0.0.1::data", LocalPath: "/data"},
			{HDFS: "hdfs://namenode/model", LocalPath: "/model"},
		},
	}
	result := buildSyncInitContainers(tt)
	if len(result) != 3 {
		t.Fatalf("expected 3 init containers, got %d", len(result))
	}
	if result[0]["name"] != "arena-sync-0" {
		t.Errorf("expected first container name=arena-sync-0, got %v", result[0]["name"])
	}
	if result[1]["name"] != "arena-sync-1" {
		t.Errorf("expected second container name=arena-sync-1, got %v", result[1]["name"])
	}
	if result[2]["name"] != "arena-sync-2" {
		t.Errorf("expected third container name=arena-sync-2, got %v", result[2]["name"])
	}
}

func TestBuildPodAffinityTerms_MatchExpressions(t *testing.T) {
	rules := []task.AffinityRule{
		{
			TopologyKey: "kubernetes.io/hostname",
			Weight:      50,
			MatchExpressions: []task.MatchExpression{
				{Key: "app", Operator: "In", Values: []string{"web", "api"}},
				{Key: "tier", Operator: "Exists"},
			},
		},
	}
	terms, err := buildPodAffinityTerms(rules, "preferred")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(terms) != 1 {
		t.Fatalf("expected 1 term, got %d", len(terms))
	}
	wt := terms[0].(map[string]interface{})
	podTerm := wt["podAffinityTerm"].(map[string]interface{})
	ls := podTerm["labelSelector"].(map[string]interface{})
	exprs := ls["matchExpressions"].([]interface{})
	if len(exprs) != 2 {
		t.Fatalf("expected 2 matchExpressions, got %d", len(exprs))
	}
	expr0 := exprs[0].(map[string]interface{})
	if expr0["key"] != "app" || expr0["operator"] != "In" {
		t.Errorf("unexpected first expression: %v", expr0)
	}
	vals := expr0["values"].([]string)
	if len(vals) != 2 || vals[0] != "web" {
		t.Errorf("unexpected values: %v", vals)
	}
}

func TestBuildPodAffinityTerms_Namespaces(t *testing.T) {
	rules := []task.AffinityRule{
		{
			TopologyKey: "kubernetes.io/hostname",
			Weight:      50,
			Namespaces:  []string{"ns1", "ns2"},
		},
	}
	terms, err := buildPodAffinityTerms(rules, "preferred")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wt := terms[0].(map[string]interface{})
	podTerm := wt["podAffinityTerm"].(map[string]interface{})
	ns := podTerm["namespaces"].([]string)
	if len(ns) != 2 || ns[0] != "ns1" || ns[1] != "ns2" {
		t.Errorf("expected namespaces [ns1, ns2], got %v", ns)
	}
}

func TestBuildPodAffinityTerms_NamespaceSelector(t *testing.T) {
	rules := []task.AffinityRule{
		{
			TopologyKey: "kubernetes.io/hostname",
			Weight:      50,
			NamespaceSelector: &task.LabelSelector{
				MatchLabels: map[string]string{"env": "prod"},
				MatchExpressions: []task.MatchExpression{
					{Key: "team", Operator: "In", Values: []string{"ml"}},
				},
			},
		},
	}
	terms, err := buildPodAffinityTerms(rules, "required")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	term := terms[0].(map[string]interface{})
	nsSel := term["namespaceSelector"].(map[string]interface{})
	ml := nsSel["matchLabels"].(map[string]string)
	if ml["env"] != "prod" {
		t.Errorf("expected namespaceSelector matchLabels env=prod, got %v", ml)
	}
	exprs := nsSel["matchExpressions"].([]interface{})
	if len(exprs) != 1 {
		t.Fatalf("expected 1 matchExpression in namespaceSelector, got %d", len(exprs))
	}
	expr := exprs[0].(map[string]interface{})
	if expr["key"] != "team" || expr["operator"] != "In" {
		t.Errorf("unexpected namespaceSelector expression: %v", expr)
	}
}

func TestBuildNodeSelectorTerms_MatchFields(t *testing.T) {
	rules := []task.AffinityRule{
		{
			MatchFields: []task.MatchExpression{
				{Key: "metadata.name", Operator: "In", Values: []string{"node-1"}},
			},
		},
	}
	terms := buildNodeSelectorTerms(rules)
	if len(terms) != 1 {
		t.Fatalf("expected 1 term, got %d", len(terms))
	}
	term := terms[0].(map[string]interface{})
	fields := term["matchFields"].([]interface{})
	if len(fields) != 1 {
		t.Fatalf("expected 1 matchField, got %d", len(fields))
	}
	field := fields[0].(map[string]interface{})
	if field["key"] != "metadata.name" || field["operator"] != "In" {
		t.Errorf("unexpected matchField: %v", field)
	}
	vals := field["values"].([]string)
	if len(vals) != 1 || vals[0] != "node-1" {
		t.Errorf("unexpected matchField values: %v", vals)
	}
}

func TestParseDuration_ValidInputs(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
	}{
		{"1h", time.Hour},
		{"30m", 30 * time.Minute},
		{"10s", 10 * time.Second},
		{"2d", 2 * 24 * time.Hour},
		{"1.5d", time.Duration(1.5 * 24 * float64(time.Hour))},
		{"", 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := parseDuration(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseDuration_InvalidInputs(t *testing.T) {
	tests := []string{
		"two hours",
		"abc",
		"1x",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, err := parseDuration(input)
			assert.Error(t, err, "parseDuration should return error for invalid input: %s", input)
			assert.Contains(t, err.Error(), "invalid duration")
		})
	}
}

func TestGetJobPodSelector(t *testing.T) {
	tests := []struct {
		provider Provider
		jobName  string
		expected string
	}{
		{&PyTorchProvider{}, "my-job", "training.kubeflow.org/job-name=my-job"},
		{&TensorFlowProvider{}, "my-job", "training.kubeflow.org/job-name=my-job"},
		{&MPIProvider{}, "my-job", "training.kubeflow.org/job-name=my-job"},
	}
	for _, tt := range tests {
		got := tt.provider.GetJobPodSelector(tt.jobName)
		if got != tt.expected {
			t.Errorf("%T.GetJobPodSelector(%q) = %q, want %q", tt.provider, tt.jobName, got, tt.expected)
		}
	}
}

func TestGetLogPodSelector(t *testing.T) {
	tests := []struct {
		provider Provider
		jobName  string
		expected string
	}{
		{&PyTorchProvider{}, "my-job", "training.kubeflow.org/job-name=my-job,training.kubeflow.org/replica-type=master"},
		{&TensorFlowProvider{}, "my-job", "training.kubeflow.org/job-name=my-job,training.kubeflow.org/replica-type=chief"},
		{&MPIProvider{}, "my-job", "training.kubeflow.org/job-name=my-job,training.kubeflow.org/replica-type=launcher"},
	}
	for _, tt := range tests {
		got := tt.provider.GetLogPodSelector(tt.jobName)
		if got != tt.expected {
			t.Errorf("%T.GetLogPodSelector(%q) = %q, want %q", tt.provider, tt.jobName, got, tt.expected)
		}
	}
}

func TestBuildVolumes_ConfigMap(t *testing.T) {
	task := &task.Task{
		Storages: []task.Storage{
			{Name: "config", ConfigMap: "app-config", MountPath: "/etc/config"},
		},
	}
	vols, mounts := BuildVolumes(task)
	if len(vols) != 1 {
		t.Fatalf("expected 1 volume, got %d", len(vols))
	}
	vol := vols[0].(map[string]interface{})
	if vol["name"] != "config" {
		t.Errorf("expected volume name 'config', got %v", vol["name"])
	}
	cm, ok := vol["configMap"].(map[string]interface{})
	if !ok {
		t.Fatal("expected configMap in volume")
	}
	if cm["name"] != "app-config" {
		t.Errorf("expected configMap.name 'app-config', got %v", cm["name"])
	}
	mount := mounts[0].(map[string]interface{})
	if mount["mountPath"] != "/etc/config" {
		t.Errorf("expected mountPath '/etc/config', got %v", mount["mountPath"])
	}
	if _, hasSubPath := mount["subPath"]; hasSubPath {
		t.Error("subPath should not be present for whole configmap mount")
	}
}

func TestBuildVolumes_ConfigMapWithKey(t *testing.T) {
	task := &task.Task{
		Storages: []task.Storage{
			{Name: "conf", ConfigMap: "foo", Key: "conf.yaml", MountPath: "/app/conf.yaml"},
		},
	}
	vols, mounts := BuildVolumes(task)
	vol := vols[0].(map[string]interface{})
	cm := vol["configMap"].(map[string]interface{})
	if cm["name"] != "foo" {
		t.Errorf("expected configMap.name 'foo', got %v", cm["name"])
	}
	mount := mounts[0].(map[string]interface{})
	if mount["subPath"] != "conf.yaml" {
		t.Errorf("expected subPath 'conf.yaml', got %v", mount["subPath"])
	}
	if mount["mountPath"] != "/app/conf.yaml" {
		t.Errorf("expected mountPath '/app/conf.yaml', got %v", mount["mountPath"])
	}
}

func TestBuildVolumes_Secret(t *testing.T) {
	task := &task.Task{
		Storages: []task.Storage{
			{Name: "creds", Secret: "db-credentials", MountPath: "/secrets"},
		},
	}
	vols, mounts := BuildVolumes(task)
	vol := vols[0].(map[string]interface{})
	secret, ok := vol["secret"].(map[string]interface{})
	if !ok {
		t.Fatal("expected secret in volume")
	}
	if secret["secretName"] != "db-credentials" {
		t.Errorf("expected secret.secretName 'db-credentials', got %v", secret["secretName"])
	}
	mount := mounts[0].(map[string]interface{})
	if _, hasSubPath := mount["subPath"]; hasSubPath {
		t.Error("subPath should not be present for whole secret mount")
	}
}

func TestBuildInitContainersWithMounts(t *testing.T) {
	tt := &task.Task{
		Name:  "test",
		Image: "busybox",
		Run:   "echo hello",
		Framework: task.Framework{Name: "pytorch"},
		Worker: &task.Worker{Replicas: 1},
		Storages: []task.Storage{
			{Name: "data", MountPath: "/data", PVC: "data-pvc"},
		},
		Init: []task.InitContainer{
			{
				Name:  "download-model",
				Image: "busybox",
				Run:   "wget -O /data/model.bin https://example.com/model.bin",
				Mounts: []task.Mount{
					{Name: "data", MountPath: "/data"},
				},
			},
		},
	}

	containers := buildInitContainers(tt)
	require.Len(t, containers, 1)

	mounts, ok := containers[0]["volumeMounts"].([]interface{})
	require.True(t, ok, "init container should have volumeMounts")
	require.Len(t, mounts, 1)
	m0 := mounts[0].(map[string]interface{})
	assert.Equal(t, "data", m0["name"])
	assert.Equal(t, "/data", m0["mountPath"])
}

func TestBuildInitContainersWithMountsSubPath(t *testing.T) {
	tt := &task.Task{
		Storages: []task.Storage{
			{Name: "config", MountPath: "/etc/config"},
		},
		Init: []task.InitContainer{
			{
				Name:  "setup",
				Image: "busybox",
				Run:   "echo setup",
				Mounts: []task.Mount{
					{Name: "config", MountPath: "/etc/config", SubPath: "app.conf"},
				},
			},
		},
	}

	containers := buildInitContainers(tt)
	require.Len(t, containers, 1)

	mounts, ok := containers[0]["volumeMounts"].([]interface{})
	require.True(t, ok, "init container should have volumeMounts")
	require.Len(t, mounts, 1)
	m0 := mounts[0].(map[string]interface{})
	assert.Equal(t, "config", m0["name"])
	assert.Equal(t, "/etc/config", m0["mountPath"])
	assert.Equal(t, "app.conf", m0["subPath"])
}

func TestBuildInitContainersWithNoMounts(t *testing.T) {
	tt := &task.Task{
		Init: []task.InitContainer{
			{
				Name:  "setup",
				Image: "busybox",
				Run:   "echo setup",
			},
		},
	}

	containers := buildInitContainers(tt)
	require.Len(t, containers, 1)

	_, hasVolumeMounts := containers[0]["volumeMounts"]
	assert.False(t, hasVolumeMounts, "init container without mounts should not have volumeMounts")
}

func TestBuildInitContainers_SubPathFallback(t *testing.T) {
	tt := &task.Task{
		Storages: []task.Storage{
			{Name: "ssh", MountPath: "/root/.ssh", Secret: "keys", SubPath: "id_rsa"},
		},
		Init: []task.InitContainer{
			{Name: "setup-ssh", Image: "busybox", Run: "echo setup", Mounts: []task.Mount{{Name: "ssh"}}},
		},
	}
	result := buildInitContainers(tt)
	require.Len(t, result, 1)
	mounts := result[0]["volumeMounts"].([]interface{})
	require.Len(t, mounts, 1)
	m0 := mounts[0].(map[string]interface{})
	assert.Equal(t, "ssh", m0["name"])
	assert.Equal(t, "/root/.ssh", m0["mountPath"])
	assert.Equal(t, "id_rsa", m0["subPath"])
}

func TestBuildInitContainers_MountPathOverride(t *testing.T) {
	tt := &task.Task{
		Storages: []task.Storage{
			{Name: "data", MountPath: "/default-data", PVC: "data-pvc"},
		},
		Init: []task.InitContainer{
			{Name: "init-data", Image: "busybox", Run: "echo init", Mounts: []task.Mount{{Name: "data", MountPath: "/custom-data"}}},
		},
	}
	result := buildInitContainers(tt)
	require.Len(t, result, 1)
	mounts := result[0]["volumeMounts"].([]interface{})
	require.Len(t, mounts, 1)
	m0 := mounts[0].(map[string]interface{})
	assert.Equal(t, "/custom-data", m0["mountPath"])
}

func TestBuildVolumes_SecretWithKey(t *testing.T) {
	task := &task.Task{
		Storages: []task.Storage{
			{Name: "ssh", Secret: "ssh-keys", Key: "id_rsa", MountPath: "/root/.ssh/id_rsa"},
		},
	}
	vols, mounts := BuildVolumes(task)
	vol := vols[0].(map[string]interface{})
	secret := vol["secret"].(map[string]interface{})
	if secret["secretName"] != "ssh-keys" {
		t.Errorf("expected secret.secretName 'ssh-keys', got %v", secret["secretName"])
	}
	mount := mounts[0].(map[string]interface{})
	if mount["subPath"] != "id_rsa" {
		t.Errorf("expected subPath 'id_rsa', got %v", mount["subPath"])
	}
}

func TestBuildSyncInitContainersWithMountsOverride(t *testing.T) {
	task := &task.Task{
		Name:  "test",
		Image: "busybox",
		Run:   "echo hello",
		Framework: task.Framework{Name: "pytorch"},
		Worker: &task.Worker{Replicas: 1},
		Storages: []task.Storage{
			{Name: "code", MountPath: "/default-code", Tmp: "1Gi"},
		},
		Sync: []task.SyncEntry{
			{
				Git:       "https://github.com/org/repo.git",
				LocalPath: "/workspace",
				Mounts: []task.Mount{
					{Name: "code", MountPath: "/workspace"},
				},
			},
		},
	}

	containers := buildSyncInitContainers(task)
	require.Len(t, containers, 1)

	// Init container should mount the overridden "code" volume at /workspace, not /default-code
	mounts, ok := containers[0]["volumeMounts"].([]interface{})
	require.True(t, ok)
	require.Len(t, mounts, 1)
	m0 := mounts[0].(map[string]interface{})
	assert.Equal(t, "code", m0["name"])
	assert.Equal(t, "/workspace", m0["mountPath"])
}

func TestBuildSyncInitContainers_RsyncWithMountsOverride(t *testing.T) {
	task := &task.Task{
		Name:  "test",
		Image: "busybox",
		Run:   "echo hello",
		Framework: task.Framework{Name: "pytorch"},
		Worker: &task.Worker{Replicas: 1},
		Storages: []task.Storage{
			{Name: "data", MountPath: "/default-data", PVC: "data-pvc"},
		},
		Sync: []task.SyncEntry{
			{
				Rsync:     "10.0.0.1::data/train.zip",
				LocalPath: "/old-path",
				Mounts: []task.Mount{
					{Name: "data", MountPath: "/data"},
				},
			},
		},
	}

	containers := buildSyncInitContainers(task)
	require.Len(t, containers, 1)

	// Check volume mount uses the override path
	mounts, ok := containers[0]["volumeMounts"].([]interface{})
	require.True(t, ok)
	require.Len(t, mounts, 1)
	m0 := mounts[0].(map[string]interface{})
	assert.Equal(t, "data", m0["name"])
	assert.Equal(t, "/data", m0["mountPath"])

	// Check command destination uses LocalPath (/old-path), not mountPath (/data)
	cmd := containers[0]["command"].([]interface{})
	require.Len(t, cmd, 4)
	assert.Equal(t, "rsync", cmd[0])
	assert.Equal(t, "-avP", cmd[1])
	assert.Equal(t, "10.0.0.1::data/train.zip", cmd[2])
	assert.Equal(t, "/old-path", cmd[3], "rsync destination should use LocalPath, not mountPath")
}

func TestBuildSyncInitContainers_HDFSWithMountsOverride(t *testing.T) {
	task := &task.Task{
		Name:  "test",
		Image: "busybox",
		Run:   "echo hello",
		Framework: task.Framework{Name: "pytorch"},
		Worker: &task.Worker{Replicas: 1},
		Storages: []task.Storage{
			{Name: "models", MountPath: "/default-models", PVC: "models-pvc"},
		},
		Sync: []task.SyncEntry{
			{
				HDFS:      "hdfs://namenode:9000/data/model",
				LocalPath: "/old-path",
				Mounts: []task.Mount{
					{Name: "models", MountPath: "/models"},
				},
			},
		},
	}

	containers := buildSyncInitContainers(task)
	require.Len(t, containers, 1)

	// Check volume mount uses the override path
	mounts, ok := containers[0]["volumeMounts"].([]interface{})
	require.True(t, ok)
	require.Len(t, mounts, 1)
	m0 := mounts[0].(map[string]interface{})
	assert.Equal(t, "models", m0["name"])
	assert.Equal(t, "/models", m0["mountPath"])

	// Check command destination uses LocalPath (/old-path), not mountPath (/models)
	cmd := containers[0]["command"].([]interface{})
	require.Len(t, cmd, 5)
	assert.Equal(t, "hdfs", cmd[0])
	assert.Equal(t, "dfs", cmd[1])
	assert.Equal(t, "-get", cmd[2])
	assert.Equal(t, "hdfs://namenode:9000/data/model", cmd[3])
	assert.Equal(t, "/old-path", cmd[4], "hdfs destination should use LocalPath, not mountPath")
}

func TestBuildVolumes_SHMDefaultMountPath(t *testing.T) {
	task := &task.Task{
		Storages: []task.Storage{
			{Name: "shm-vol", SHM: "8Gi"},
		},
	}
	vols, mounts := BuildVolumes(task)
	require.Len(t, vols, 1)
	require.Len(t, mounts, 1)
	mount := mounts[0].(map[string]interface{})
	assert.Equal(t, "/dev/shm", mount["mountPath"])
}

func TestResolveMounts_Empty(t *testing.T) {
	result := ResolveMounts(nil, nil)
	assert.Nil(t, result)
}

func TestResolveMounts_FallbackToStorage(t *testing.T) {
	mounts := []task.Mount{{Name: "code"}}
	storages := []task.Storage{{Name: "code", MountPath: "/default-code", PVC: "pvc"}}
	result := ResolveMounts(mounts, storages)
	require.Len(t, result, 1)
	assert.Equal(t, "code", result[0]["name"])
	assert.Equal(t, "/default-code", result[0]["mountPath"])
}

func TestResolveMounts_OverridePath(t *testing.T) {
	mounts := []task.Mount{{Name: "code", MountPath: "/override"}}
	storages := []task.Storage{{Name: "code", MountPath: "/default-code", PVC: "pvc"}}
	result := ResolveMounts(mounts, storages)
	require.Len(t, result, 1)
	assert.Equal(t, "/override", result[0]["mountPath"])
}

func TestResolveMounts_SubPathFallback(t *testing.T) {
	mounts := []task.Mount{{Name: "ssh"}}
	storages := []task.Storage{{Name: "ssh", MountPath: "/root/.ssh", Secret: "keys", SubPath: "id_rsa"}}
	result := ResolveMounts(mounts, storages)
	require.Len(t, result, 1)
	assert.Equal(t, "id_rsa", result[0]["subPath"])
}

func TestResolveMounts_SubPathOverride(t *testing.T) {
	mounts := []task.Mount{{Name: "ssh", SubPath: "id_ed25519"}}
	storages := []task.Storage{{Name: "ssh", MountPath: "/root/.ssh", Secret: "keys", SubPath: "id_rsa"}}
	result := ResolveMounts(mounts, storages)
	require.Len(t, result, 1)
	assert.Equal(t, "id_ed25519", result[0]["subPath"])
}

func TestResolveMounts_ConfigMapKeyFallback(t *testing.T) {
	mounts := []task.Mount{{Name: "config"}}
	storages := []task.Storage{
		{Name: "config", MountPath: "/etc/config", ConfigMap: "app-config", Key: "config.yaml"},
	}
	result := ResolveMounts(mounts, storages)
	require.Len(t, result, 1)
	assert.Equal(t, "config", result[0]["name"])
	assert.Equal(t, "/etc/config", result[0]["mountPath"])
	assert.Equal(t, "config.yaml", result[0]["subPath"], "subPath should fall back to Storage.Key when SubPath is empty")
}

func TestResolveMounts_SecretKeyFallback(t *testing.T) {
	mounts := []task.Mount{{Name: "ssh"}}
	storages := []task.Storage{
		{Name: "ssh", MountPath: "/root/.ssh", Secret: "ssh-keys", Key: "id_rsa"},
	}
	result := ResolveMounts(mounts, storages)
	require.Len(t, result, 1)
	assert.Equal(t, "id_rsa", result[0]["subPath"], "subPath should fall back to Storage.Key for Secret storages")
}

func TestResolveMounts_StorageNotFound(t *testing.T) {
	mounts := []task.Mount{{Name: "nonexistent", MountPath: "/data"}}
	storages := []task.Storage{{Name: "code", MountPath: "/code", PVC: "pvc"}}
	result := ResolveMounts(mounts, storages)
	require.Len(t, result, 1)
	// Volume name still comes from mount, just no fallback path available
	assert.Equal(t, "nonexistent", result[0]["name"])
}

func TestResolveMounts_MultipleEntries(t *testing.T) {
	mounts := []task.Mount{
		{Name: "code", MountPath: "/workspace"},
		{Name: "data"},
	}
	storages := []task.Storage{
		{Name: "code", MountPath: "/code", PVC: "code-pvc"},
		{Name: "data", MountPath: "/data", PVC: "data-pvc"},
	}
	result := ResolveMounts(mounts, storages)
	require.Len(t, result, 2)
	assert.Equal(t, "code", result[0]["name"])
	assert.Equal(t, "/workspace", result[0]["mountPath"])
	assert.Equal(t, "data", result[1]["name"])
	assert.Equal(t, "/data", result[1]["mountPath"])
}

func TestResolveContainerMounts_NoStorages(t *testing.T) {
	tt := &task.Task{}
	result := ResolveContainerMounts(nil, tt)
	assert.Nil(t, result)
}

func TestResolveContainerMounts_StoragesNoMounts(t *testing.T) {
	tt := &task.Task{
		Storages: []task.Storage{
			{Name: "data", PVC: "pvc1", MountPath: "/data"},
			{Name: "cache", HostPath: "/tmp/cache", MountPath: "/cache"},
		},
	}
	result := ResolveContainerMounts(nil, tt)
	require.Len(t, result, 2)
	m0 := result[0].(map[string]interface{})
	assert.Equal(t, "data", m0["name"])
	assert.Equal(t, "/data", m0["mountPath"])
	m1 := result[1].(map[string]interface{})
	assert.Equal(t, "cache", m1["name"])
	assert.Equal(t, "/cache", m1["mountPath"])
}

func TestResolveContainerMounts_StoragesWithMounts(t *testing.T) {
	tt := &task.Task{
		Storages: []task.Storage{
			{Name: "data", PVC: "pvc1", MountPath: "/data"},
			{Name: "logs", PVC: "pvc2", MountPath: "/logs"},
		},
	}
	mounts := []task.Mount{
		{Name: "logs", MountPath: "/tb/logs"},
	}
	result := ResolveContainerMounts(mounts, tt)
	require.Len(t, result, 1)
	m0 := result[0].(map[string]interface{})
	assert.Equal(t, "logs", m0["name"])
	assert.Equal(t, "/tb/logs", m0["mountPath"])
}

func TestBuildSyncInitContainers_NewNaming(t *testing.T) {
	tt := &task.Task{
		Sync: []task.SyncEntry{
			{Git: "https://github.com/org/repo.git", LocalPath: "/code"},
			{Rsync: "10.0.0.1::data", LocalPath: "/data"},
			{HDFS: "hdfs://namenode/model", LocalPath: "/model"},
		},
	}
	result := buildSyncInitContainers(tt)
	require.Len(t, result, 3)
	assert.Equal(t, "arena-sync-0", result[0]["name"])
	assert.Equal(t, "arena-sync-1", result[1]["name"])
	assert.Equal(t, "arena-sync-2", result[2]["name"])
}

func TestBuildSyncInitContainers_GitUsesLocalPath(t *testing.T) {
	tt := &task.Task{
		Storages: []task.Storage{{Name: "code", MountPath: "/workspace", Tmp: "5Gi"}},
		Sync: []task.SyncEntry{
			{
				Git:       "https://github.com/org/repo.git",
				Branch:    "main",
				LocalPath: "/code",
				Mounts:    []task.Mount{{Name: "code"}},
			},
		},
	}
	result := buildSyncInitContainers(tt)
	require.Len(t, result, 1)

	// GIT_SYNC_ROOT uses local_path, not mount_path
	envs := result[0]["env"].([]map[string]interface{})
	found := map[string]string{}
	for _, e := range envs {
		found[e["name"].(string)] = e["value"].(string)
	}
	assert.Equal(t, "/code", found["GIT_SYNC_ROOT"])

	// Volume mount references storage by name, uses storage's mount_path
	mounts := result[0]["volumeMounts"].([]interface{})
	require.Len(t, mounts, 1)
	m0 := mounts[0].(map[string]interface{})
	assert.Equal(t, "code", m0["name"])
	assert.Equal(t, "/workspace", m0["mountPath"])
}

func TestBuildSyncInitContainers_RsyncUsesLocalPath(t *testing.T) {
	tt := &task.Task{
		Storages: []task.Storage{{Name: "data", MountPath: "/data-vol", PVC: "data-pvc"}},
		Sync: []task.SyncEntry{
			{
				Rsync:     "10.0.0.1::data/train.zip",
				LocalPath: "/old-path",
				Mounts:    []task.Mount{{Name: "data"}},
			},
		},
	}
	result := buildSyncInitContainers(tt)
	require.Len(t, result, 1)

	// Rsync dest uses local_path, not mount_path
	cmd := result[0]["command"].([]interface{})
	require.Len(t, cmd, 4)
	assert.Equal(t, "/old-path", cmd[3])
}

func TestBuildSyncInitContainers_HDFSUsesLocalPath(t *testing.T) {
	tt := &task.Task{
		Storages: []task.Storage{{Name: "models", MountPath: "/models-vol", PVC: "models-pvc"}},
		Sync: []task.SyncEntry{
			{
				HDFS:      "hdfs://namenode:9000/data/model",
				LocalPath: "/local-models",
				Mounts:    []task.Mount{{Name: "models"}},
			},
		},
	}
	result := buildSyncInitContainers(tt)
	require.Len(t, result, 1)

	// HDFS dest uses local_path, not mount_path
	cmd := result[0]["command"].([]interface{})
	require.Len(t, cmd, 5)
	assert.Equal(t, "/local-models", cmd[4])
}

func TestBuildSyncInitContainers_NoMounts(t *testing.T) {
	tt := &task.Task{
		Sync: []task.SyncEntry{
			{Git: "https://github.com/org/repo.git", LocalPath: "/code"},
		},
	}
	result := buildSyncInitContainers(tt)
	require.Len(t, result, 1)
	assert.Equal(t, "arena-sync-0", result[0]["name"])

	// No volumeMounts when sync has no mounts
	_, hasMounts := result[0]["volumeMounts"]
	assert.False(t, hasMounts)
}

func TestBuildSyncInitContainers_FallsBackToAllStorages(t *testing.T) {
	tt := &task.Task{
		Storages: []task.Storage{
			{Name: "data", PVC: "data-pvc", MountPath: "/data"},
			{Name: "cache", HostPath: "/tmp/cache", MountPath: "/cache"},
		},
		Sync: []task.SyncEntry{
			{Git: "https://github.com/org/repo.git", LocalPath: "/code"},
		},
	}
	result := buildSyncInitContainers(tt)
	require.Len(t, result, 1)

	// With storages but no mounts, all storages should be mounted
	mounts, ok := result[0]["volumeMounts"].([]interface{})
	require.True(t, ok, "sync container should have volumeMounts when storages exist")
	require.Len(t, mounts, 2)
	m0 := mounts[0].(map[string]interface{})
	assert.Equal(t, "data", m0["name"])
	assert.Equal(t, "/data", m0["mountPath"])
	m1 := mounts[1].(map[string]interface{})
	assert.Equal(t, "cache", m1["name"])
	assert.Equal(t, "/cache", m1["mountPath"])
}

func TestBuildInitContainers_FallsBackToAllStorages(t *testing.T) {
	tt := &task.Task{
		Storages: []task.Storage{
			{Name: "data", PVC: "data-pvc", MountPath: "/data"},
		},
		Init: []task.InitContainer{
			{Name: "setup", Image: "busybox", Run: "echo setup"},
		},
	}
	result := buildInitContainers(tt)
	require.Len(t, result, 1)

	// With storages but no mounts, all storages should be mounted
	mounts, ok := result[0]["volumeMounts"].([]interface{})
	require.True(t, ok, "init container should have volumeMounts when storages exist")
	require.Len(t, mounts, 1)
	m0 := mounts[0].(map[string]interface{})
	assert.Equal(t, "data", m0["name"])
	assert.Equal(t, "/data", m0["mountPath"])
}

// --- Exported wrapper tests ---

func TestBuildVolumes_Exported(t *testing.T) {
	tk := &task.Task{
		Storages: []task.Storage{
			{Name: "data", PVC: "my-pvc", MountPath: "/data"},
		},
	}
	volumes, mounts := BuildVolumes(tk)
	assert.Len(t, volumes, 1)
	assert.Len(t, mounts, 1)
}

func TestBuildVolumes_ExportedEmpty(t *testing.T) {
	tk := &task.Task{}
	volumes, mounts := BuildVolumes(tk)
	assert.Nil(t, volumes)
	assert.Nil(t, mounts)
}

func TestResolveMounts_Exported(t *testing.T) {
	mounts := []task.Mount{
		{Name: "data", MountPath: "/override"},
	}
	storages := []task.Storage{
		{Name: "data", MountPath: "/original"},
	}
	result := ResolveMounts(mounts, storages)
	assert.Len(t, result, 1)
	assert.Equal(t, "/override", result[0]["mountPath"])
}

func TestResolveMounts_ExportedEmpty(t *testing.T) {
	result := ResolveMounts(nil, nil)
	assert.Nil(t, result)
}

func TestBuildSyncInitContainers_WarningLocalPathMismatch(t *testing.T) {
	// Capture stderr
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	tt := &task.Task{
		Storages: []task.Storage{{Name: "code", MountPath: "/workspace", Tmp: "5Gi"}},
		Sync: []task.SyncEntry{
			{
				Git:       "https://github.com/org/repo.git",
				LocalPath: "/code",
				Mounts:    []task.Mount{{Name: "code"}},
			},
		},
	}
	buildSyncInitContainers(tt)

	w.Close()
	os.Stderr = old
	var buf bytes.Buffer
	io.Copy(&buf, r)
	assert.Contains(t, buf.String(), "warning")
	assert.Contains(t, buf.String(), "/code")
}

func TestEffectiveRun(t *testing.T) {
	tk := &task.Task{Run: "python train.py"}

	// roleRun empty → falls back to top-level
	assert.Equal(t, "python train.py", effectiveRun(tk, ""))

	// roleRun non-empty → overrides top-level
	assert.Equal(t, "python ps.py", effectiveRun(tk, "python ps.py"))
}
