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
	// The parser treats the dot in nvidia.com as a path separator, producing
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
	assert.IsType(t, true, gang["enabled"], "coerceValue should parse 'true' as bool, not string")
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

func TestApplySetOverrides_ParseError(t *testing.T) {
	yamlData := []byte("name: test\n")
	_, err := ApplySetOverrides(yamlData, []string{"foo[abc]=value"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse --set")
}

// Tests for standalone parser functions

func TestParseKeyPath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		want    []pathSegment
		wantErr bool
	}{
		{
			name: "simple key",
			path: "name",
			want: []pathSegment{{key: "name"}},
		},
		{
			name: "dotted path",
			path: "worker.replicas",
			want: []pathSegment{{key: "worker"}, {key: "replicas"}},
		},
		{
			name: "deeply nested",
			path: "scheduling.gang.enabled",
			want: []pathSegment{{key: "scheduling"}, {key: "gang"}, {key: "enabled"}},
		},
		{
			name: "array index",
			path: "storages[0]",
			want: []pathSegment{{key: "storages"}, {index: 0, isArr: true}},
		},
		{
			name: "array index with key",
			path: "storages[0].name",
			want: []pathSegment{{key: "storages"}, {index: 0, isArr: true}, {key: "name"}},
		},
		{
			name: "multiple array indices",
			path: "matrix[0][1]",
			want: []pathSegment{{key: "matrix"}, {index: 0, isArr: true}, {index: 1, isArr: true}},
		},
		{
			name: "complex path",
			path: "a.b[0].c[1].d",
			want: []pathSegment{
				{key: "a"}, {key: "b"}, {index: 0, isArr: true},
				{key: "c"}, {index: 1, isArr: true}, {key: "d"},
			},
		},
		{
			name:    "unclosed bracket",
			path:    "a[0",
			wantErr: true,
		},
		{
			name:    "invalid array index",
			path:    "a[abc]",
			wantErr: true,
		},
		{
			name:    "negative array index",
			path:    "a[-1]",
			wantErr: true,
		},
		{
			name:    "index exceeds MaxIndex",
			path:    "a[65536]",
			wantErr: true,
		},
		{
			name: "placeholder key",
			path: "__ARENA_QK_0__",
			want: []pathSegment{{key: "__ARENA_QK_0__"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseKeyPath(tt.path)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCoerceValue(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  interface{}
	}{
		{"true", "true", true},
		{"false", "false", false},
		{"null", "null", nil},
		{"positive int", "42", 42},
		{"negative int", "-1", -1},
		{"zero", "0", 0},
		{"string", "hello", "hello"},
		{"string with spaces", "hello world", "hello world"},
		{"resource quantity", "16Gi", "16Gi"},
		{"empty string", "", ""},
		{"string with equals", "a=b", "a=b"},
		{"leading zero", "007", 7}, // strconv.Atoi accepts leading zeros
		{"float-like", "1.5", "1.5"}, // not parsed as int
		{"True capital", "True", true},
		{"FALSE capital", "FALSE", false},
		{"TRUE all caps", "TRUE", true},
		{"False mixed", "False", false},
		{"Null capital", "Null", "Null"}, // null is exact-match only
		{"NULL all caps", "NULL", "NULL"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := coerceValue(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPreprocessQuotedKeys(t *testing.T) {
	tests := []struct {
		name       string
		expr       string
		wantExpr   string
		wantKeys   map[string]string
		wantErr    bool
		errContain string
	}{
		{
			name:     "no quotes",
			expr:     "a.b.c=value",
			wantExpr: "a.b.c=value",
			wantKeys: map[string]string{},
		},
		{
			name:     "single quoted key",
			expr:     "a.'b.c'.d=value",
			wantExpr: "a.__ARENA_QK_0__.d=value",
			wantKeys: map[string]string{"__ARENA_QK_0__": "b.c"},
		},
		{
			name:     "quoted key with slash",
			expr:     "worker.resources.'nvidia.com/gpu'=4",
			wantExpr: "worker.resources.__ARENA_QK_0__=4",
			wantKeys: map[string]string{"__ARENA_QK_0__": "nvidia.com/gpu"},
		},
		{
			name:     "multiple quoted keys",
			expr:     "'a.b'.'c.d'=value",
			wantExpr: "__ARENA_QK_0__.__ARENA_QK_1__=value",
			wantKeys: map[string]string{"__ARENA_QK_0__": "a.b", "__ARENA_QK_1__": "c.d"},
		},
		{
			name:     "quotes in value ignored",
			expr:     "key='value.with.dots'",
			wantExpr: "key='value.with.dots'",
			wantKeys: map[string]string{},
		},
		{
			name:       "mismatched quote",
			expr:       "'abc.def=value",
			wantErr:    true,
			errContain: "mismatched single quote",
		},
		{
			name:       "empty quoted segment",
			expr:       "a.''.b=value",
			wantErr:    true,
			errContain: "empty quoted segment",
		},
		{
			name:     "no equals sign",
			expr:     "a.'b.c'",
			wantExpr: "a.__ARENA_QK_0__",
			wantKeys: map[string]string{"__ARENA_QK_0__": "b.c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotExpr, gotKeys, err := preprocessQuotedKeys(tt.expr)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContain != "" {
					assert.Contains(t, err.Error(), tt.errContain)
				}
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantExpr, gotExpr)
			assert.Equal(t, tt.wantKeys, gotKeys)
		})
	}
}

func TestApplySetOverrides_ArrayCreation(t *testing.T) {
	yamlData := []byte("name: test\n")
	result, err := ApplySetOverrides(yamlData, []string{"items[0]=first"})
	require.NoError(t, err)

	var m map[string]interface{}
	require.NoError(t, yaml.Unmarshal(result, &m))
	items := m["items"].([]interface{})
	assert.Equal(t, 1, len(items))
	assert.Equal(t, "first", items[0])
}

func TestApplySetOverrides_ArrayGapFilledWithNil(t *testing.T) {
	yamlData := []byte("name: test\n")
	result, err := ApplySetOverrides(yamlData, []string{"items[2]=third"})
	require.NoError(t, err)

	var m map[string]interface{}
	require.NoError(t, yaml.Unmarshal(result, &m))
	items := m["items"].([]interface{})
	assert.Equal(t, 3, len(items))
	assert.Nil(t, items[0])
	assert.Nil(t, items[1])
	assert.Equal(t, "third", items[2])
}

func TestApplySetOverrides_NullValue(t *testing.T) {
	yamlData := []byte("name: test\nimage: pytorch:2.1\n")
	result, err := ApplySetOverrides(yamlData, []string{"image=null"})
	require.NoError(t, err)

	var m map[string]interface{}
	require.NoError(t, yaml.Unmarshal(result, &m))
	assert.Nil(t, m["image"])
}

func TestApplySetOverrides_OverwriteExistingQuotedKey(t *testing.T) {
	// Regression test: when the YAML already contains a key with dots (e.g.
	// nvidia.com/gpu: 1) and --set uses a quoted key to overwrite it, the
	// restoreQuotedKeys function must ensure the new value wins regardless of
	// Go's randomized map iteration order. Before the fix, the old value
	// would win ~50% of the time.
	yamlData := []byte("worker:\n  resources:\n    nvidia.com/gpu: 1\n")
	for i := 0; i < 100; i++ {
		result, err := ApplySetOverrides(yamlData, []string{"worker.resources.'nvidia.com/gpu'=2"})
		require.NoError(t, err)

		var m map[string]interface{}
		require.NoError(t, yaml.Unmarshal(result, &m))
		worker := m["worker"].(map[string]interface{})
		resources := worker["resources"].(map[string]interface{})
		assert.Equal(t, 2, resources["nvidia.com/gpu"], "iteration %d: override should win", i)
	}
}

func TestApplySetOverrides_MaxIndexExceeded(t *testing.T) {
	yamlData := []byte("name: test\n")
	_, err := ApplySetOverrides(yamlData, []string{"items[65536]=value"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds maximum")
}

func TestApplySetOverrides_CaseInsensitiveBool(t *testing.T) {
	yamlData := []byte("name: test\n")
	result, err := ApplySetOverrides(yamlData, []string{"flag=True", "other=FALSE"})
	require.NoError(t, err)

	var m map[string]interface{}
	require.NoError(t, yaml.Unmarshal(result, &m))
	assert.Equal(t, true, m["flag"])
	assert.Equal(t, false, m["other"])
}

// ---------------------------------------------------------------------------
// Schema coverage tests: verify --set can override every major YAML field.
// ---------------------------------------------------------------------------

// --- Framework options ---

func TestApplySetOverrides_FrameworkOptionsPyTorch(t *testing.T) {
	yamlData := []byte(`name: test
image: x
run: echo
framework:
  name: pytorch
worker:
  replicas: 1
`)
	result, err := ApplySetOverrides(yamlData, []string{
		"framework.options.nproc_per_node=auto",
	})
	require.NoError(t, err)

	tk, err := LoadFromBytes(result)
	require.NoError(t, err)
	assert.Equal(t, "auto", tk.Framework.Options.NprocPerNode)
}

func TestApplySetOverrides_FrameworkOptionsMPI(t *testing.T) {
	yamlData := []byte(`name: test
image: x
run: echo
framework:
  name: mpi
worker:
  replicas: 2
`)
	result, err := ApplySetOverrides(yamlData, []string{
		"framework.options.slots_per_worker=4",
		"framework.options.mpi_implementation=Intel",
		"framework.options.mounts_on_launcher=true",
		"framework.options.gpu_topology=true",
		"framework.options.run_launcher_as_worker=true",
		"framework.options.launcher_creation_policy=WaitForWorkersReady",
		"framework.options.ssh_auth_mount_path=/etc/ssh",
	})
	require.NoError(t, err)

	tk, err := LoadFromBytes(result)
	require.NoError(t, err)
	assert.Equal(t, 4, tk.Framework.Options.SlotsPerWorker)
	assert.Equal(t, "Intel", tk.Framework.Options.MPIImplementation)
	assert.True(t, tk.Framework.Options.MountsOnLauncher)
	assert.True(t, tk.Framework.Options.GPUTopology)
	assert.True(t, tk.Framework.Options.RunLauncherAsWorker)
	assert.Equal(t, "WaitForWorkersReady", tk.Framework.Options.LauncherCreationPolicy)
	assert.Equal(t, "/etc/ssh", tk.Framework.Options.SSHAuthMountPath)
}

// --- Roles (master, launcher, ps, chief, evaluator) ---

func TestApplySetOverrides_MasterRole(t *testing.T) {
	yamlData := []byte(`name: test
image: x
run: echo
framework:
  name: pytorch
master:
  replicas: 1
`)
	result, err := ApplySetOverrides(yamlData, []string{
		"master.resources.cpu=8",
		"master.resources.memory=32Gi",
		"master.envs.TRAINING=pytorch",
	})
	require.NoError(t, err)

	tk, err := LoadFromBytes(result)
	require.NoError(t, err)
	require.NotNil(t, tk.Master)
	assert.Equal(t, "8", tk.Master.Resources["cpu"])
	assert.Equal(t, "32Gi", tk.Master.Resources["memory"])
	assert.Equal(t, "pytorch", tk.Master.Envs["TRAINING"].Value)
}

func TestApplySetOverrides_LauncherRole(t *testing.T) {
	yamlData := []byte(`name: test
image: x
run: echo
framework:
  name: mpi
worker:
  replicas: 2
launcher:
  replicas: 1
`)
	result, err := ApplySetOverrides(yamlData, []string{
		"launcher.resources.cpu=4",
		"launcher.resources.memory=16Gi",
		"launcher.envs.OMPI_MCA_orte_rsh_agent=ssh",
	})
	require.NoError(t, err)

	tk, err := LoadFromBytes(result)
	require.NoError(t, err)
	require.NotNil(t, tk.Launcher)
	assert.Equal(t, "4", tk.Launcher.Resources["cpu"])
	assert.Equal(t, "16Gi", tk.Launcher.Resources["memory"])
	assert.Equal(t, "ssh", tk.Launcher.Envs["OMPI_MCA_orte_rsh_agent"].Value)
}

func TestApplySetOverrides_TFRoles(t *testing.T) {
	yamlData := []byte(`name: test
image: x
run: echo
framework:
  name: tensorflow
worker:
  replicas: 2
ps:
  replicas: 1
chief:
  replicas: 1
evaluator:
  replicas: 1
`)
	result, err := ApplySetOverrides(yamlData, []string{
		"ps.replicas=3",
		"ps.resources.cpu=2",
		"chief.resources.memory=8Gi",
		"evaluator.envs.MODE=test",
	})
	require.NoError(t, err)

	tk, err := LoadFromBytes(result)
	require.NoError(t, err)
	require.NotNil(t, tk.PS)
	assert.Equal(t, 3, tk.PS.Replicas)
	assert.Equal(t, "2", tk.PS.Resources["cpu"])
	require.NotNil(t, tk.Chief)
	assert.Equal(t, "8Gi", tk.Chief.Resources["memory"])
	require.NotNil(t, tk.Evaluator)
	assert.Equal(t, "test", tk.Evaluator.Envs["MODE"].Value)
}

func TestApplySetOverrides_WorkerRoleFields(t *testing.T) {
	yamlData := []byte(`name: test
image: x
run: echo
framework:
  name: pytorch
worker:
  replicas: 2
`)
	result, err := ApplySetOverrides(yamlData, []string{
		"worker.replicas=4",
		"worker.resources.cpu=8",
		"worker.resources.memory=32Gi",
		"worker.envs.NCCL_DEBUG=INFO",
		"worker.run=python train.py --epochs 100",
	})
	require.NoError(t, err)

	tk, err := LoadFromBytes(result)
	require.NoError(t, err)
	require.NotNil(t, tk.Worker)
	assert.Equal(t, 4, tk.Worker.Replicas)
	assert.Equal(t, "8", tk.Worker.Resources["cpu"])
	assert.Equal(t, "32Gi", tk.Worker.Resources["memory"])
	assert.Equal(t, "INFO", tk.Worker.Envs["NCCL_DEBUG"].Value)
	assert.Equal(t, "python train.py --epochs 100", tk.Worker.Run)
}

// --- Sync entries ---

func TestApplySetOverrides_SyncGit(t *testing.T) {
	yamlData := []byte(`name: test
image: x
run: echo
framework:
  name: pytorch
worker:
  replicas: 1
storages:
  - name: code
    mount_path: /workspace
    pvc: code-pvc
sync:
  - git: https://github.com/example/repo.git
    local_path: /workspace
`)
	result, err := ApplySetOverrides(yamlData, []string{
		"sync[0].git=https://github.com/new/repo.git",
		"sync[0].branch=dev",
		"sync[0].image=ghcr.io/fluxcd/git-sync:v4",
	})
	require.NoError(t, err)

	tk, err := LoadFromBytes(result)
	require.NoError(t, err)
	require.Len(t, tk.Sync, 1)
	assert.Equal(t, "https://github.com/new/repo.git", tk.Sync[0].Git)
	assert.Equal(t, "dev", tk.Sync[0].Branch)
	assert.Equal(t, "ghcr.io/fluxcd/git-sync:v4", tk.Sync[0].Image)
}

func TestApplySetOverrides_SyncRsync(t *testing.T) {
	yamlData := []byte(`name: test
image: x
run: echo
framework:
  name: pytorch
worker:
  replicas: 1
sync:
  - rsync: rsync://old-server:/data
    local_path: /data
`)
	result, err := ApplySetOverrides(yamlData, []string{
		"sync[0].rsync=rsync://new-server:/data",
		"sync[0].local_path=/workspace",
	})
	require.NoError(t, err)

	tk, err := LoadFromBytes(result)
	require.NoError(t, err)
	require.Len(t, tk.Sync, 1)
	assert.Equal(t, "rsync://new-server:/data", tk.Sync[0].Rsync)
	assert.Equal(t, "/workspace", tk.Sync[0].LocalPath)
}

func TestApplySetOverrides_SyncHDFS(t *testing.T) {
	yamlData := []byte(`name: test
image: x
run: echo
framework:
  name: pytorch
worker:
  replicas: 1
sync:
  - hdfs: hdfs://old-nn:8020/data
    local_path: /data
`)
	result, err := ApplySetOverrides(yamlData, []string{
		"sync[0].hdfs=hdfs://new-nn:8020/data",
	})
	require.NoError(t, err)

	tk, err := LoadFromBytes(result)
	require.NoError(t, err)
	require.Len(t, tk.Sync, 1)
	assert.Equal(t, "hdfs://new-nn:8020/data", tk.Sync[0].HDFS)
}

func TestApplySetOverrides_SyncAddNewEntry(t *testing.T) {
	yamlData := []byte(`name: test
image: x
run: echo
framework:
  name: pytorch
worker:
  replicas: 1
storages:
  - name: code
    mount_path: /workspace
    pvc: code-pvc
`)
	result, err := ApplySetOverrides(yamlData, []string{
		"sync[0].git=https://github.com/example/repo.git",
		"sync[0].local_path=/workspace",
		"sync[0].branch=main",
	})
	require.NoError(t, err)

	tk, err := LoadFromBytes(result)
	require.NoError(t, err)
	require.Len(t, tk.Sync, 1)
	assert.Equal(t, "https://github.com/example/repo.git", tk.Sync[0].Git)
	assert.Equal(t, "/workspace", tk.Sync[0].LocalPath)
	assert.Equal(t, "main", tk.Sync[0].Branch)
}

func TestApplySetOverrides_SyncMounts(t *testing.T) {
	yamlData := []byte(`name: test
image: x
run: echo
framework:
  name: pytorch
worker:
  replicas: 1
storages:
  - name: code-vol
    mount_path: /workspace
    pvc: code-pvc
sync:
  - git: https://github.com/example/repo.git
    local_path: /workspace
`)
	result, err := ApplySetOverrides(yamlData, []string{
		"sync[0].mounts[0].name=code-vol",
		"sync[0].mounts[0].mount_path=/code",
	})
	require.NoError(t, err)

	tk, err := LoadFromBytes(result)
	require.NoError(t, err)
	require.Len(t, tk.Sync, 1)
	require.Len(t, tk.Sync[0].Mounts, 1)
	assert.Equal(t, "code-vol", tk.Sync[0].Mounts[0].Name)
	assert.Equal(t, "/code", tk.Sync[0].Mounts[0].MountPath)
}

// --- Init containers ---

func TestApplySetOverrides_InitContainers(t *testing.T) {
	yamlData := []byte(`name: test
image: x
run: echo
framework:
  name: pytorch
worker:
  replicas: 1
init:
  - name: setup
    image: busybox:latest
    run: echo initializing
`)
	result, err := ApplySetOverrides(yamlData, []string{
		"init[0].image=alpine:3.18",
		"init[0].shell=/bin/sh",
		"init[0].run=echo ready",
	})
	require.NoError(t, err)

	tk, err := LoadFromBytes(result)
	require.NoError(t, err)
	require.Len(t, tk.Init, 1)
	assert.Equal(t, "alpine:3.18", tk.Init[0].Image)
	assert.Equal(t, "/bin/sh", tk.Init[0].Shell)
	assert.Equal(t, "echo ready", tk.Init[0].Run)
}

func TestApplySetOverrides_InitAddNewContainer(t *testing.T) {
	yamlData := []byte(`name: test
image: x
run: echo
framework:
  name: pytorch
worker:
  replicas: 1
`)
	result, err := ApplySetOverrides(yamlData, []string{
		"init[0].name=warmup",
		"init[0].image=busybox:latest",
		"init[0].run=echo warming up",
	})
	require.NoError(t, err)

	tk, err := LoadFromBytes(result)
	require.NoError(t, err)
	require.Len(t, tk.Init, 1)
	assert.Equal(t, "warmup", tk.Init[0].Name)
	assert.Equal(t, "busybox:latest", tk.Init[0].Image)
	assert.Equal(t, "echo warming up", tk.Init[0].Run)
}

// --- Storages ---

func TestApplySetOverrides_StoragePVC(t *testing.T) {
	yamlData := []byte(`name: test
image: x
run: echo
framework:
  name: pytorch
worker:
  replicas: 1
storages:
  - name: data
    mount_path: /data
    pvc: old-pvc
`)
	result, err := ApplySetOverrides(yamlData, []string{
		"storages[0].pvc=new-pvc",
		"storages[0].mount_path=/workspace",
	})
	require.NoError(t, err)

	tk, err := LoadFromBytes(result)
	require.NoError(t, err)
	require.Len(t, tk.Storages, 1)
	assert.Equal(t, "new-pvc", tk.Storages[0].PVC)
	assert.Equal(t, "/workspace", tk.Storages[0].MountPath)
}

func TestApplySetOverrides_StorageHostPath(t *testing.T) {
	yamlData := []byte(`name: test
image: x
run: echo
framework:
  name: pytorch
worker:
  replicas: 1
storages:
  - name: data
    mount_path: /data
    pvc: my-pvc
`)
	result, err := ApplySetOverrides(yamlData, []string{
		"storages[0].pvc=",
		"storages[0].hostpath=/host/data",
	})
	require.NoError(t, err)

	tk, err := LoadFromBytes(result)
	require.NoError(t, err)
	require.Len(t, tk.Storages, 1)
	assert.Equal(t, "/host/data", tk.Storages[0].HostPath)
}

func TestApplySetOverrides_StorageConfigMap(t *testing.T) {
	yamlData := []byte(`name: test
image: x
run: echo
framework:
  name: pytorch
worker:
  replicas: 1
storages:
  - name: config
    mount_path: /config
    pvc: dummy
`)
	result, err := ApplySetOverrides(yamlData, []string{
		"storages[0].pvc=",
		"storages[0].configmap=my-config",
		"storages[0].key=app.conf",
	})
	require.NoError(t, err)

	tk, err := LoadFromBytes(result)
	require.NoError(t, err)
	require.Len(t, tk.Storages, 1)
	assert.Equal(t, "my-config", tk.Storages[0].ConfigMap)
	assert.Equal(t, "app.conf", tk.Storages[0].Key)
}

func TestApplySetOverrides_StorageSecret(t *testing.T) {
	yamlData := []byte(`name: test
image: x
run: echo
framework:
  name: pytorch
worker:
  replicas: 1
storages:
  - name: secrets
    mount_path: /secrets
    pvc: dummy
`)
	result, err := ApplySetOverrides(yamlData, []string{
		"storages[0].pvc=",
		"storages[0].secret=my-secret",
	})
	require.NoError(t, err)

	tk, err := LoadFromBytes(result)
	require.NoError(t, err)
	require.Len(t, tk.Storages, 1)
	assert.Equal(t, "my-secret", tk.Storages[0].Secret)
}

func TestApplySetOverrides_StorageSHMandTmp(t *testing.T) {
	yamlData := []byte(`name: test
image: x
run: echo
framework:
  name: pytorch
worker:
  replicas: 1
storages:
  - name: data
    mount_path: /data
    pvc: my-pvc
`)
	result, err := ApplySetOverrides(yamlData, []string{
		"storages[0].pvc=",
		"storages[0].shm=2Gi",
	})
	require.NoError(t, err)

	tk, err := LoadFromBytes(result)
	require.NoError(t, err)
	require.Len(t, tk.Storages, 1)
	assert.Equal(t, "2Gi", tk.Storages[0].SHM)
}

func TestApplySetOverrides_StorageAddNewEntry(t *testing.T) {
	yamlData := []byte(`name: test
image: x
run: echo
framework:
  name: pytorch
worker:
  replicas: 1
`)
	result, err := ApplySetOverrides(yamlData, []string{
		"storages[0].name=data",
		"storages[0].mount_path=/data",
		"storages[0].pvc=my-pvc",
	})
	require.NoError(t, err)

	tk, err := LoadFromBytes(result)
	require.NoError(t, err)
	require.Len(t, tk.Storages, 1)
	assert.Equal(t, "data", tk.Storages[0].Name)
	assert.Equal(t, "/data", tk.Storages[0].MountPath)
	assert.Equal(t, "my-pvc", tk.Storages[0].PVC)
}

// --- Scheduling: tolerations, affinity rules, node_selector ---

func TestApplySetOverrides_Tolerations(t *testing.T) {
	yamlData := []byte(`name: test
image: x
run: echo
framework:
  name: pytorch
worker:
  replicas: 1
`)
	result, err := ApplySetOverrides(yamlData, []string{
		"scheduling.tolerations[0].key=nvidia.com/gpu",
		"scheduling.tolerations[0].operator=Exists",
		"scheduling.tolerations[0].effect=NoSchedule",
	})
	require.NoError(t, err)

	tk, err := LoadFromBytes(result)
	require.NoError(t, err)
	require.Len(t, tk.Scheduling.Tolerations, 1)
	assert.Equal(t, "nvidia.com/gpu", tk.Scheduling.Tolerations[0].Key)
	assert.Equal(t, "Exists", tk.Scheduling.Tolerations[0].Operator)
	assert.Equal(t, "NoSchedule", tk.Scheduling.Tolerations[0].Effect)
}

func TestApplySetOverrides_AffinityRules(t *testing.T) {
	yamlData := []byte(`name: test
image: x
run: echo
framework:
  name: pytorch
worker:
  replicas: 1
`)
	result, err := ApplySetOverrides(yamlData, []string{
		"scheduling.affinity.policy=spread",
		"scheduling.affinity.constraint=preferred",
		"scheduling.affinity.target=node",
		"scheduling.affinity.rules[0].topology_key=kubernetes.io/hostname",
		"scheduling.affinity.rules[0].weight=100",
	})
	require.NoError(t, err)

	tk, err := LoadFromBytes(result)
	require.NoError(t, err)
	require.NotNil(t, tk.Scheduling.Affinity)
	assert.Equal(t, "spread", tk.Scheduling.Affinity.Policy)
	assert.Equal(t, "preferred", tk.Scheduling.Affinity.Constraint)
	assert.Equal(t, "node", tk.Scheduling.Affinity.Target)
	require.Len(t, tk.Scheduling.Affinity.Rules, 1)
	assert.Equal(t, "kubernetes.io/hostname", tk.Scheduling.Affinity.Rules[0].TopologyKey)
	assert.Equal(t, 100, tk.Scheduling.Affinity.Rules[0].Weight)
}

func TestApplySetOverrides_NodeSelector(t *testing.T) {
	yamlData := []byte(`name: test
image: x
run: echo
framework:
  name: pytorch
worker:
  replicas: 1
`)
	result, err := ApplySetOverrides(yamlData, []string{
		"scheduling.node_selector.accelerator=nvidia",
		"scheduling.node_selector.zone=us-east-1",
	})
	require.NoError(t, err)

	tk, err := LoadFromBytes(result)
	require.NoError(t, err)
	assert.Equal(t, "nvidia", tk.Scheduling.NodeSelector["accelerator"])
	assert.Equal(t, "us-east-1", tk.Scheduling.NodeSelector["zone"])
}

// --- Labels, annotations ---

func TestApplySetOverrides_LabelsAndAnnotations(t *testing.T) {
	yamlData := []byte(`name: test
image: x
run: echo
framework:
  name: pytorch
worker:
  replicas: 1
`)
	result, err := ApplySetOverrides(yamlData, []string{
		"labels.team=ml-platform",
		"labels.env=production",
		"annotations.managed-by=arena",
	})
	require.NoError(t, err)

	tk, err := LoadFromBytes(result)
	require.NoError(t, err)
	assert.Equal(t, "ml-platform", tk.Labels["team"])
	assert.Equal(t, "production", tk.Labels["env"])
	assert.Equal(t, "arena", tk.Annotations["managed-by"])
}

// --- Runtime fields ---

func TestApplySetOverrides_RuntimeFields(t *testing.T) {
	yamlData := []byte(`name: test
image: x
run: echo
framework:
  name: pytorch
worker:
  replicas: 1
`)
	result, err := ApplySetOverrides(yamlData, []string{
		"image_pull_policy=Always",
		"image_pull_secrets[0]=my-secret",
		"service_account=my-sa",
		"restart=OnFailure",
		"host_network=true",
		"host_ipc=true",
		"host_pid=true",
	})
	require.NoError(t, err)

	tk, err := LoadFromBytes(result)
	require.NoError(t, err)
	assert.Equal(t, "Always", tk.ImagePullPolicy)
	assert.Equal(t, []string{"my-secret"}, tk.ImagePullSecrets)
	assert.Equal(t, "my-sa", tk.ServiceAccount)
	assert.Equal(t, "OnFailure", tk.Restart)
	assert.True(t, tk.HostNetwork)
	assert.True(t, tk.HostIPC)
	assert.True(t, tk.HostPID)
}

// --- Logging / TensorBoard ---

func TestApplySetOverrides_LoggingTensorBoard(t *testing.T) {
	yamlData := []byte(`name: test
image: x
run: echo
framework:
  name: pytorch
worker:
  replicas: 1
`)
	result, err := ApplySetOverrides(yamlData, []string{
		"logging.tensorboard.enabled=true",
		"logging.tensorboard.logdir=/logs",
		"logging.tensorboard.image=tensorboard/tensorboard:latest",
	})
	require.NoError(t, err)

	tk, err := LoadFromBytes(result)
	require.NoError(t, err)
	require.NotNil(t, tk.Logging.TensorBoard)
	assert.True(t, tk.Logging.TensorBoard.Enabled)
	assert.Equal(t, "/logs", tk.Logging.TensorBoard.LogDir)
	assert.Equal(t, "tensorboard/tensorboard:latest", tk.Logging.TensorBoard.Image)
}

// --- Identity fields ---

func TestApplySetOverrides_IdentityFields(t *testing.T) {
	yamlData := []byte(`name: old-name
image: x
run: echo
framework:
  name: pytorch
worker:
  replicas: 1
`)
	result, err := ApplySetOverrides(yamlData, []string{
		"name=new-name",
		"namespace=ml-team",
		"description=training job v2",
		"working_dir=/app",
		"shell=/bin/bash",
	})
	require.NoError(t, err)

	tk, err := LoadFromBytes(result)
	require.NoError(t, err)
	assert.Equal(t, "new-name", tk.Name)
	assert.Equal(t, "ml-team", tk.Namespace)
	assert.Equal(t, "training job v2", tk.Description)
	assert.Equal(t, "/app", tk.WorkingDir)
	assert.Equal(t, "/bin/bash", tk.Shell)
}
