package task

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestApplySetOverrides_EmptyExpressions(t *testing.T) {
	yamlData := []byte("name: test\nimage: pytorch:2.1\n")
	result, err := ApplySetOverrides(yamlData, nil)
	require.NoError(t, err)

	var m map[string]interface{}
	require.NoError(t, yaml.Unmarshal(result, &m))
	assert.Equal(t, "test", m["name"])
	assert.Equal(t, "pytorch:2.1", m["image"])
}

func TestApplySetOverrides_EmptyExpressionSlice(t *testing.T) {
	yamlData := []byte("name: test\n")
	result, err := ApplySetOverrides(yamlData, []string{})
	require.NoError(t, err)

	var m map[string]interface{}
	require.NoError(t, yaml.Unmarshal(result, &m))
	assert.Equal(t, "test", m["name"])
}

func TestApplySetOverrides_TopLevelString(t *testing.T) {
	yamlData := []byte("name: old-name\nimage: pytorch:2.1\n")
	result, err := ApplySetOverrides(yamlData, []string{"name=new-name"})
	require.NoError(t, err)

	var m map[string]interface{}
	require.NoError(t, yaml.Unmarshal(result, &m))
	assert.Equal(t, "new-name", m["name"])
	assert.Equal(t, "pytorch:2.1", m["image"])
}

func TestApplySetOverrides_TopLevelInt(t *testing.T) {
	yamlData := []byte("name: test\nimage: x\nrun: echo\nframework:\n  name: pytorch\nworker:\n  replicas: 2\n")
	result, err := ApplySetOverrides(yamlData, []string{"worker.replicas=8"})
	require.NoError(t, err)

	var m map[string]interface{}
	require.NoError(t, yaml.Unmarshal(result, &m))
	worker := m["worker"].(map[string]interface{})
	assert.Equal(t, 8, worker["replicas"])
}

func TestApplySetOverrides_GangEnabled(t *testing.T) {
	yamlData := []byte("name: test\nimage: x\nrun: echo\nframework:\n  name: pytorch\nworker:\n  replicas: 1\nscheduling:\n  gang:\n    enabled: false\n")
	result, err := ApplySetOverrides(yamlData, []string{"scheduling.gang.enabled=true"})
	require.NoError(t, err)

	var m map[string]interface{}
	require.NoError(t, yaml.Unmarshal(result, &m))
	sched := m["scheduling"].(map[string]interface{})
	assert.Equal(t, map[string]interface{}{"enabled": true}, sched["gang"])
}

func TestApplySetOverrides_SchedulingFields(t *testing.T) {
	yamlData := []byte(`name: test
image: test:latest
framework:
  name: pytorch
worker:
  replicas: 1
`)
	merged, err := ApplySetOverrides(yamlData, []string{
		"scheduling.priority=100",
		"scheduling.priority_class_name=high",
		"scheduling.gang.enabled=true",
		"scheduling.scheduler_name=volcano",
	})
	require.NoError(t, err)

	var result map[string]interface{}
	err = yaml.Unmarshal(merged, &result)
	require.NoError(t, err)

	sched := result["scheduling"].(map[string]interface{})
	assert.Equal(t, 100, sched["priority"])
	assert.Equal(t, "high", sched["priority_class_name"])
	assert.Equal(t, map[string]interface{}{"enabled": true}, sched["gang"])
	assert.Equal(t, "volcano", sched["scheduler_name"])
}

func TestApplySetOverrides_NestedPath(t *testing.T) {
	yamlData := []byte("name: test\nimage: x\nrun: echo\nframework:\n  name: pytorch\nworker:\n  replicas: 1\n  resources:\n    cpu: \"2\"\n")
	result, err := ApplySetOverrides(yamlData, []string{"worker.resources.memory=16Gi"})
	require.NoError(t, err)

	var m map[string]interface{}
	require.NoError(t, yaml.Unmarshal(result, &m))
	worker := m["worker"].(map[string]interface{})
	resources := worker["resources"].(map[string]interface{})
	assert.Equal(t, "16Gi", resources["memory"])
	assert.Equal(t, "2", resources["cpu"])
}

func TestApplySetOverrides_MapValueOverride(t *testing.T) {
	yamlData := []byte("name: test\nimage: x\nrun: echo\nframework:\n  name: pytorch\nworker:\n  replicas: 1\nenvs:\n  EXISTING: old\n")
	result, err := ApplySetOverrides(yamlData, []string{"envs.NEW_VAR=hello"})
	require.NoError(t, err)

	var m map[string]interface{}
	require.NoError(t, yaml.Unmarshal(result, &m))
	envs := m["envs"].(map[string]interface{})
	assert.Equal(t, "hello", envs["NEW_VAR"])
	assert.Equal(t, "old", envs["EXISTING"])
}

func TestApplySetOverrides_MultipleExpressionsLastWins(t *testing.T) {
	yamlData := []byte("name: test\nimage: x\nrun: echo\nframework:\n  name: pytorch\nworker:\n  replicas: 1\n")
	result, err := ApplySetOverrides(yamlData, []string{"worker.replicas=2", "worker.replicas=4"})
	require.NoError(t, err)

	var m map[string]interface{}
	require.NoError(t, yaml.Unmarshal(result, &m))
	worker := m["worker"].(map[string]interface{})
	assert.Equal(t, 4, worker["replicas"])
}

func TestApplySetOverrides_InvalidExpression(t *testing.T) {
	yamlData := []byte("name: test\n")
	_, err := ApplySetOverrides(yamlData, []string{"=nokey"})
	require.Error(t, err)
}

func TestApplySetOverrides_InvalidYAML(t *testing.T) {
	yamlData := []byte("not: valid: yaml: {{{")
	_, err := ApplySetOverrides(yamlData, []string{"name=test"})
	require.Error(t, err)
}

func TestApplySetOverrides_ArrayIndexOverride(t *testing.T) {
	yamlData := []byte(`name: test
image: x
run: echo
framework:
  name: pytorch
worker:
  replicas: 1
storages:
  - name: old-ds
    mount_path: /data
    pvc: my-pvc
  - name: other
    mount_path: /other
    pvc: other-pvc
`)
	result, err := ApplySetOverrides(yamlData, []string{"storages[0].name=new-ds"})
	require.NoError(t, err)

	var m map[string]interface{}
	require.NoError(t, yaml.Unmarshal(result, &m))
	storages := m["storages"].([]interface{})
	first := storages[0].(map[string]interface{})
	assert.Equal(t, "new-ds", first["name"])
	assert.Equal(t, "/data", first["mount_path"])
	second := storages[1].(map[string]interface{})
	assert.Equal(t, "other", second["name"])
}

func TestApplySetOverrides_EnvValueSecretRoundTrip(t *testing.T) {
	yamlData := []byte(`name: test
image: x
run: echo
framework:
  name: pytorch
worker:
  replicas: 1
envs:
  HF_TOKEN:
    secret: my-creds
    key: token
  PLAIN_VAR: hello
`)
	result, err := ApplySetOverrides(yamlData, []string{"envs.NEW_KEY=world"})
	require.NoError(t, err)

	tk, err := LoadFromBytes(result)
	require.NoError(t, err)

	assert.Equal(t, "world", tk.Envs["NEW_KEY"].Value)
	assert.Equal(t, "hello", tk.Envs["PLAIN_VAR"].Value)
	require.NotNil(t, tk.Envs["HF_TOKEN"].Secret)
	assert.Equal(t, "my-creds", tk.Envs["HF_TOKEN"].Secret.Name)
	assert.Equal(t, "token", tk.Envs["HF_TOKEN"].Secret.Key)
}

func TestApplySetOverrides_EnvValueConfigMapRoundTrip(t *testing.T) {
	yamlData := []byte(`name: test
image: x
run: echo
framework:
  name: pytorch
worker:
  replicas: 1
envs:
  DB_HOST:
    configmap: db-config
    key: host
`)
	result, err := ApplySetOverrides(yamlData, []string{"envs.NEW=added"})
	require.NoError(t, err)

	tk, err := LoadFromBytes(result)
	require.NoError(t, err)

	require.NotNil(t, tk.Envs["DB_HOST"].ConfigMap)
	assert.Equal(t, "db-config", tk.Envs["DB_HOST"].ConfigMap.Name)
	assert.Equal(t, "host", tk.Envs["DB_HOST"].ConfigMap.Key)
	assert.Equal(t, "added", tk.Envs["NEW"].Value)
}

func TestApplySetOverrides_NewKeysCreated(t *testing.T) {
	yamlData := []byte("name: test\nimage: x\nrun: echo\nframework:\n  name: pytorch\nworker:\n  replicas: 1\n")
	result, err := ApplySetOverrides(yamlData, []string{"labels.team=ml", "labels.env=prod"})
	require.NoError(t, err)

	var m map[string]interface{}
	require.NoError(t, yaml.Unmarshal(result, &m))
	labels := m["labels"].(map[string]interface{})
	assert.Equal(t, "ml", labels["team"])
	assert.Equal(t, "prod", labels["env"])
}

func TestApplySetOverrides_ResourceQuantityWithSlash(t *testing.T) {
	yamlData := []byte("name: test\nimage: x\nrun: echo\nframework:\n  name: pytorch\nworker:\n  replicas: 1\n  resources:\n    cpu: \"2\"\n")
	result, err := ApplySetOverrides(yamlData, []string{"worker.resources.nvidia.com/gpu=4"})
	require.NoError(t, err)

	var m map[string]interface{}
	require.NoError(t, yaml.Unmarshal(result, &m))
	worker := m["worker"].(map[string]interface{})
	resources := worker["resources"].(map[string]interface{})
	assert.Equal(t, "2", resources["cpu"])
	// strvals treats the dot in nvidia.com as a path separator, producing
	// resources.nvidia["com/gpu"]=4 rather than resources["nvidia.com/gpu"]=4.
	// This validates that the slash in com/gpu is preserved as a key segment.
	nvidia := resources["nvidia"].(map[string]interface{})
	assert.Equal(t, 4, nvidia["com/gpu"])
}

func TestApplySetOverrides_NilYAMLWithExpressions(t *testing.T) {
	result, err := ApplySetOverrides(nil, []string{"name=test"})
	require.NoError(t, err)

	var m map[string]interface{}
	require.NoError(t, yaml.Unmarshal(result, &m))
	assert.Equal(t, "test", m["name"])
}

func TestApplySetOverrides_QuotedKey(t *testing.T) {
	yamlData := []byte("name: test\nimage: x\nrun: echo\nframework:\n  name: pytorch\nworker:\n  replicas: 1\n")
	result, err := ApplySetOverrides(yamlData, []string{"worker.resources.'nvidia.com/gpu'=4"})
	require.NoError(t, err)

	var m map[string]interface{}
	require.NoError(t, yaml.Unmarshal(result, &m))
	worker := m["worker"].(map[string]interface{})
	resources := worker["resources"].(map[string]interface{})
	assert.Equal(t, 4, resources["nvidia.com/gpu"])
}

func TestApplySetOverrides_NestedQuotedKey(t *testing.T) {
	yamlData := []byte("name: test\nimage: x\nrun: echo\nframework:\n  name: pytorch\nworker:\n  replicas: 1\n  resources:\n    cpu: \"2\"\n")
	result, err := ApplySetOverrides(yamlData, []string{"worker.resources.'nvidia.com/gpu'=4"})
	require.NoError(t, err)

	tk, err := LoadFromBytes(result)
	require.NoError(t, err)
	assert.Equal(t, "4", tk.Worker.Resources["nvidia.com/gpu"])
	assert.Equal(t, "2", tk.Worker.Resources["cpu"])
}

func TestApplySetOverrides_MultipleQuotedKeys(t *testing.T) {
	yamlData := []byte("name: test\nimage: x\nrun: echo\nframework:\n  name: pytorch\nworker:\n  replicas: 1\n")
	result, err := ApplySetOverrides(yamlData, []string{"'a.b'.'c.d'=value"})
	require.NoError(t, err)

	var m map[string]interface{}
	require.NoError(t, yaml.Unmarshal(result, &m))
	ab := m["a.b"].(map[string]interface{})
	assert.Equal(t, "value", ab["c.d"])
}

func TestApplySetOverrides_QuotedKeyUnquotedKey(t *testing.T) {
	yamlData := []byte("name: test\nimage: x\nrun: echo\nframework:\n  name: pytorch\nworker:\n  replicas: 1\n")
	result, err := ApplySetOverrides(yamlData, []string{"worker.replicas=8", "worker.resources.'nvidia.com/gpu'=2"})
	require.NoError(t, err)

	tk, err := LoadFromBytes(result)
	require.NoError(t, err)
	assert.Equal(t, 8, tk.Worker.Replicas)
	assert.Equal(t, "2", tk.Worker.Resources["nvidia.com/gpu"])
}

func TestApplySetOverrides_TopLevelQuotedKey(t *testing.T) {
	yamlData := []byte("name: test\nimage: x\nrun: echo\nframework:\n  name: pytorch\nworker:\n  replicas: 1\n")
	result, err := ApplySetOverrides(yamlData, []string{"'nvidia.com/gpu'=4"})
	require.NoError(t, err)

	var m map[string]interface{}
	require.NoError(t, yaml.Unmarshal(result, &m))
	assert.Equal(t, 4, m["nvidia.com/gpu"])
}

func TestApplySetOverrides_EmptyQuotedSegment(t *testing.T) {
	yamlData := []byte("name: test\n")
	_, err := ApplySetOverrides(yamlData, []string{"''.foo=bar"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty quoted segment")
}

func TestApplySetOverrides_MismatchedQuote(t *testing.T) {
	yamlData := []byte("name: test\n")
	_, err := ApplySetOverrides(yamlData, []string{"'abc.foo=bar"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mismatched single quote")
}

func TestApplySetOverrides_DeeplyNestedOverride(t *testing.T) {
	yamlData := []byte(`name: test
image: x
run: echo
framework:
  name: pytorch
worker:
  replicas: 1
scheduling:
  affinity:
    policy: spread
    constraint: preferred
    target: node
`)
	result, err := ApplySetOverrides(yamlData, []string{"scheduling.affinity.policy=binpack"})
	require.NoError(t, err)

	var m map[string]interface{}
	require.NoError(t, yaml.Unmarshal(result, &m))
	sched := m["scheduling"].(map[string]interface{})
	affinity := sched["affinity"].(map[string]interface{})
	assert.Equal(t, "binpack", affinity["policy"])
	assert.Equal(t, "preferred", affinity["constraint"])
}

func TestApplySetOverrides_QuotedValueSideIgnored(t *testing.T) {
	yamlData := []byte("name: test\nimage: x\nrun: echo\nframework:\n  name: pytorch\nworker:\n  replicas: 1\n  resources:\n    cpu: \"2\"\n")
	result, err := ApplySetOverrides(yamlData, []string{"worker.resources.gpu='nvidia.com/gpu'"})
	require.NoError(t, err)

	var m map[string]interface{}
	require.NoError(t, yaml.Unmarshal(result, &m))
	worker := m["worker"].(map[string]interface{})
	resources := worker["resources"].(map[string]interface{})
	// The value should be the literal string 'nvidia.com/gpu', not a placeholder
	assert.Equal(t, "'nvidia.com/gpu'", resources["gpu"])
}

func TestApplySetOverrides_QuotedKeyInArrayElement(t *testing.T) {
	yamlData := []byte(`name: test
image: x
run: echo
framework:
  name: pytorch
worker:
  replicas: 1
storages:
  - name: data-vol
    mount_path: /data
    pvc: my-pvc
`)
	result, err := ApplySetOverrides(yamlData, []string{"storages[0].'volume.beta.kubernetes.io/storage-class'=fast"})
	require.NoError(t, err)

	var m map[string]interface{}
	require.NoError(t, yaml.Unmarshal(result, &m))

	storages := m["storages"].([]interface{})
	first := storages[0].(map[string]interface{})
	assert.Equal(t, "fast", first["volume.beta.kubernetes.io/storage-class"],
		"quoted key inside array element should be restored, not left as placeholder")
	assert.Equal(t, "data-vol", first["name"])
	assert.Equal(t, "/data", first["mount_path"])
}

func TestApplySetOverrides_WhitespaceInValue(t *testing.T) {
	yamlData := []byte("name: old\nimage: x\nrun: echo\nframework:\n  name: pytorch\nworker:\n  replicas: 1\n")
	result, err := ApplySetOverrides(yamlData, []string{"name=hello world"})
	require.NoError(t, err)

	var m map[string]interface{}
	require.NoError(t, yaml.Unmarshal(result, &m))
	assert.Equal(t, "hello world", m["name"])
}

func TestApplySetOverrides_MultipleEquals(t *testing.T) {
	yamlData := []byte("name: old\nimage: x\nrun: echo\nframework:\n  name: pytorch\nworker:\n  replicas: 1\n")
	result, err := ApplySetOverrides(yamlData, []string{"name=a=b"})
	require.NoError(t, err)

	var m map[string]interface{}
	require.NoError(t, yaml.Unmarshal(result, &m))
	assert.Equal(t, "a=b", m["name"])
}

func TestApplySetOverrides_NegativeNumber(t *testing.T) {
	yamlData := []byte("name: test\nimage: x\nrun: echo\nframework:\n  name: pytorch\nworker:\n  replicas: 1\n")
	result, err := ApplySetOverrides(yamlData, []string{"worker.replicas=-1"})
	require.NoError(t, err)

	var m map[string]interface{}
	require.NoError(t, yaml.Unmarshal(result, &m))
	worker := m["worker"].(map[string]interface{})
	assert.Equal(t, -1, worker["replicas"])
}

func TestApplySetOverrides_BooleanValueType(t *testing.T) {
	yamlData := []byte("name: test\nimage: x\nrun: echo\nframework:\n  name: pytorch\nworker:\n  replicas: 1\nscheduling:\n  gang:\n    enabled: false\n")
	result, err := ApplySetOverrides(yamlData, []string{"scheduling.gang.enabled=true"})
	require.NoError(t, err)

	var m map[string]interface{}
	require.NoError(t, yaml.Unmarshal(result, &m))
	sched := m["scheduling"].(map[string]interface{})
	gang := sched["gang"].(map[string]interface{})
	assert.Equal(t, true, gang["enabled"])
	assert.IsType(t, true, gang["enabled"], "strvals should parse 'true' as bool, not string")
}

func TestApplySetOverrides_EmptyValue(t *testing.T) {
	yamlData := []byte("name: old\nimage: x\nrun: echo\nframework:\n  name: pytorch\nworker:\n  replicas: 1\n")
	result, err := ApplySetOverrides(yamlData, []string{"name="})
	require.NoError(t, err)

	var m map[string]interface{}
	require.NoError(t, yaml.Unmarshal(result, &m))
	assert.Equal(t, "", m["name"])
}

func TestApplySetOverrides_QuotedKeyAdjacentToUnquoted(t *testing.T) {
	yamlData := []byte("name: test\n")
	result, err := ApplySetOverrides(yamlData, []string{"foo'bar.baz'qux=value"})
	require.NoError(t, err)

	var m map[string]interface{}
	require.NoError(t, yaml.Unmarshal(result, &m))
	assert.Equal(t, "value", m["foobar.bazqux"])
}

func TestApplySetOverrides_StrvalsParseError(t *testing.T) {
	yamlData := []byte("name: test\n")
	_, err := ApplySetOverrides(yamlData, []string{"foo[abc]=value"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse --set")
}
