package provider

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"

	"github.com/kubeflow/arena/pkg/task"
)

// assertK8sAffinityValid validates that the generated affinity map deserializes
// cleanly into corev1.Affinity with no unknown fields. This catches structural
// errors like missing "preference" wrappers or wrong nesting levels.
func assertK8sAffinityValid(t *testing.T, affinityMap map[string]interface{}) corev1.Affinity {
	t.Helper()

	data, err := json.Marshal(affinityMap)
	require.NoError(t, err, "failed to marshal affinity map")

	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()

	var affinity corev1.Affinity
	require.NoError(t, dec.Decode(&affinity),
		"affinity map should be structurally valid as corev1.Affinity")
	return affinity
}

func TestBuildAffinity_Nil(t *testing.T) {
	result := buildAffinity(nil, "test-job")
	if result != nil {
		t.Errorf("expected nil for nil affinity, got %v", result)
	}
}

func TestBuildAffinity_Empty(t *testing.T) {
	a := &task.Affinity{}
	result := buildAffinity(a, "test-job")
	if result != nil {
		t.Errorf("expected nil for empty affinity, got %v", result)
	}
}

func TestBuildAffinity_PolicyOnlyNoRules(t *testing.T) {
	// Policy without rules is a no-op in buildAffinity; task.Validate rejects this configuration.
	a := &task.Affinity{
		Policy: "spread",
		Target: "pod",
	}
	result := buildAffinity(a, "test-job")
	if result != nil {
		t.Errorf("expected nil result for policy without rules, got %v", result)
	}
}

func TestBuildAffinity_RulesWithPodTarget(t *testing.T) {
	a := &task.Affinity{
		Target: "pod",
		Policy: "binpack",
		Rules: []task.AffinityRule{
			{
				TopologyKey: "kubernetes.io/hostname",
				Weight:      50,
				MatchLabels: map[string]string{"app": "test"},
			},
		},
	}
	result := buildAffinity(a, "test-job")
	if result == nil {
		t.Fatal("expected non-nil result for rules with pod target")
	}
	assertK8sAffinityValid(t, result)
	if _, ok := result["podAffinity"]; !ok {
		t.Errorf("expected podAffinity for binpack policy, got %v", result)
	}
}

func TestBuildAffinity_UnsetPolicyGeneratesNothing(t *testing.T) {
	a := &task.Affinity{
		Target: "pod",
		Rules: []task.AffinityRule{
			{
				TopologyKey: "kubernetes.io/hostname",
				Weight:      50,
				MatchLabels: map[string]string{"app": "test"},
			},
		},
	}
	if result := buildAffinity(a, "test-job"); result != nil {
		t.Errorf("expected nil for unset policy even with rules, got %v", result)
	}

	a.Target = "node"
	a.Rules = []task.AffinityRule{
		{MatchExpressions: []task.MatchExpression{{Key: "gpu-type", Operator: "Exists"}}},
	}
	if result := buildAffinity(a, "test-job"); result != nil {
		t.Errorf("expected nil for unset policy with node target, got %v", result)
	}
}

func TestBuildAffinity_RulesWithNodeTarget(t *testing.T) {
	a := &task.Affinity{
		Target: "node",
		Policy: "spread",
		Rules: []task.AffinityRule{
			{
				Weight: 1,
				MatchExpressions: []task.MatchExpression{
					{Key: "gpu-type", Operator: "In", Values: []string{"A100"}},
				},
			},
		},
	}
	result := buildAffinity(a, "test-job")
	if result == nil {
		t.Fatal("expected non-nil result for rules with node target")
	}
	assertK8sAffinityValid(t, result)
	// Should generate nodeAffinity
	if _, ok := result["nodeAffinity"]; !ok {
		t.Errorf("expected nodeAffinity, got %v", result)
	}
}

func TestBuildPodAffinityTerms_PreferredWithValidWeight(t *testing.T) {
	rules := []task.AffinityRule{
		{
			TopologyKey: "kubernetes.io/hostname",
			Weight:      50,
			MatchLabels: map[string]string{"app": "test"},
		},
	}
	terms := buildPodAffinityTerms(rules, "preferred")
	if len(terms) != 1 {
		t.Fatalf("expected 1 term, got %d", len(terms))
	}
	weightedTerm, ok := terms[0].(map[string]interface{})
	if !ok {
		t.Fatal("expected map[string]interface{} for weighted term")
	}
	if w, ok := weightedTerm["weight"].(int64); !ok || w != 50 {
		t.Errorf("expected outer weight=50, got %v", weightedTerm["weight"])
	}
	podTerm, ok := weightedTerm["podAffinityTerm"].(map[string]interface{})
	if !ok {
		t.Fatal("expected podAffinityTerm to be map[string]interface{}")
	}
	if _, hasWeight := podTerm["weight"]; hasWeight {
		t.Error("weight must not appear inside podAffinityTerm")
	}
	if tk, ok := podTerm["topologyKey"].(string); !ok || tk != "kubernetes.io/hostname" {
		t.Errorf("expected topologyKey=kubernetes.io/hostname, got %v", podTerm["topologyKey"])
	}
}

func TestBuildPodAffinityTerms_RequiredIgnoresWeight(t *testing.T) {
	rules := []task.AffinityRule{
		{
			TopologyKey: "kubernetes.io/hostname",
			Weight:      50,
			MatchLabels: map[string]string{"app": "test"},
		},
	}
	terms := buildPodAffinityTerms(rules, "required")
	if len(terms) != 1 {
		t.Fatalf("expected 1 term, got %d", len(terms))
	}
	term, ok := terms[0].(map[string]interface{})
	if !ok {
		t.Fatal("expected map[string]interface{}")
	}
	if _, hasWeight := term["weight"]; hasWeight {
		t.Error("weight must not appear in required mode term")
	}
	if _, hasWrapper := term["podAffinityTerm"]; hasWrapper {
		t.Error("required mode term must not have podAffinityTerm wrapper")
	}
}

func TestBuildAffinity_PreferredSpread_NoWeightInsidePodAffinityTerm(t *testing.T) {
	a := &task.Affinity{
		Policy:     "spread",
		Constraint: "preferred",
		Target:     "pod",
		Rules: []task.AffinityRule{
			{
				TopologyKey: "kubernetes.io/hostname",
				Weight:      50,
				MatchLabels: map[string]string{"app": "test"},
			},
		},
	}
	result := buildAffinity(a, "test-job")
	assertK8sAffinityValid(t, result)
	antiAffinity, ok := result["podAntiAffinity"].(map[string]interface{})
	if !ok {
		t.Fatal("expected podAntiAffinity for spread policy")
	}
	preferred, ok := antiAffinity["preferredDuringSchedulingIgnoredDuringExecution"].([]interface{})
	if !ok {
		t.Fatal("expected preferredDuringSchedulingIgnoredDuringExecution array")
	}
	if len(preferred) != 1 {
		t.Fatalf("expected 1 preferred term, got %d", len(preferred))
	}
	wt, ok := preferred[0].(map[string]interface{})
	if !ok {
		t.Fatal("expected weighted term map")
	}
	if w, ok := wt["weight"].(int64); !ok || w != 50 {
		t.Errorf("expected outer weight=50, got %v", wt["weight"])
	}
	podTerm, ok := wt["podAffinityTerm"].(map[string]interface{})
	if !ok {
		t.Fatal("expected podAffinityTerm map")
	}
	if _, hasWeight := podTerm["weight"]; hasWeight {
		t.Error("weight must not appear inside podAffinityTerm")
	}
}

func TestBuildAffinity_NonePolicySkipsRules(t *testing.T) {
	a := &task.Affinity{
		Policy: "none",
		Target: "pod",
		Rules: []task.AffinityRule{
			{
				TopologyKey: "kubernetes.io/hostname",
				Weight:      50,
				MatchLabels: map[string]string{"app": "test"},
			},
		},
	}
	result := buildAffinity(a, "test-job")
	if result != nil {
		t.Errorf("expected nil for policy=none even with rules, got %v", result)
	}
}

func TestBuildAffinity_NodeTargetSpreadPreservesOperators(t *testing.T) {
	a := &task.Affinity{
		Target: "node",
		Policy: "spread",
		Rules: []task.AffinityRule{
			{
				Weight: 1,
				MatchExpressions: []task.MatchExpression{
					{Key: "gpu-type", Operator: "In", Values: []string{"A100"}},
				},
				MatchLabels: map[string]string{"zone": "us-east-1"},
			},
		},
	}
	result := buildAffinity(a, "test-job")
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	assertK8sAffinityValid(t, result)

	nodeAffinity, ok := result["nodeAffinity"].(map[string]interface{})
	if !ok {
		t.Fatal("expected nodeAffinity")
	}
	preferred, ok := nodeAffinity["preferredDuringSchedulingIgnoredDuringExecution"].([]interface{})
	if !ok || len(preferred) != 1 {
		t.Fatalf("expected 1 preferred term, got %v", nodeAffinity)
	}
	term, ok := preferred[0].(map[string]interface{})
	if !ok {
		t.Fatal("expected term to be a map")
	}
	if w, ok := term["weight"].(int64); !ok || w != 1 {
		t.Errorf("expected weight=1, got %v", term["weight"])
	}
	pref, ok := term["preference"].(map[string]interface{})
	if !ok {
		t.Fatal("expected preference wrapper")
	}
	exprs, ok := pref["matchExpressions"].([]interface{})
	if !ok {
		t.Fatal("expected matchExpressions inside preference")
	}
	foundIn := false
	for _, e := range exprs {
		expr, ok := e.(map[string]interface{})
		if !ok {
			continue
		}
		if op, _ := expr["operator"].(string); op == "In" {
			foundIn = true
		}
		if op, _ := expr["operator"].(string); op == "NotIn" {
			t.Error("expected In operator to be preserved for spread policy, got NotIn")
		}
	}
	if !foundIn {
		t.Error("expected to find In operator preserved for spread policy")
	}
}

func TestBuildAffinity_NodeTargetBinpackPreservesOperators(t *testing.T) {
	a := &task.Affinity{
		Target: "node",
		Policy: "binpack",
		Rules: []task.AffinityRule{
			{
				Weight: 1,
				MatchExpressions: []task.MatchExpression{
					{Key: "gpu-type", Operator: "In", Values: []string{"A100"}},
				},
			},
		},
	}
	result := buildAffinity(a, "test-job")
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	assertK8sAffinityValid(t, result)

	nodeAffinity, ok := result["nodeAffinity"].(map[string]interface{})
	if !ok {
		t.Fatal("expected nodeAffinity")
	}
	preferred, ok := nodeAffinity["preferredDuringSchedulingIgnoredDuringExecution"].([]interface{})
	if !ok || len(preferred) != 1 {
		t.Fatalf("expected 1 preferred term, got %v", nodeAffinity)
	}
	term, ok := preferred[0].(map[string]interface{})
	if !ok {
		t.Fatal("expected term to be a map")
	}
	pref, ok := term["preference"].(map[string]interface{})
	if !ok {
		t.Fatal("expected preference wrapper")
	}
	exprs, ok := pref["matchExpressions"].([]interface{})
	if !ok {
		t.Fatal("expected matchExpressions inside preference")
	}
	found := false
	for _, e := range exprs {
		expr, ok := e.(map[string]interface{})
		if !ok {
			continue
		}
		if op, _ := expr["operator"].(string); op == "In" {
			found = true
		}
	}
	if !found {
		t.Error("expected In operator to be preserved for binpack policy")
	}
}

func TestBuildTopologySpreadConstraints_NilAffinity(t *testing.T) {
	result := buildTopologySpreadConstraints(nil, "test-job")
	if result != nil {
		t.Errorf("expected nil for nil affinity, got %v", result)
	}
}

func TestBuildTopologySpreadConstraints_NodeTargetSpread(t *testing.T) {
	a := &task.Affinity{
		Policy:     "spread",
		Constraint: "preferred",
		Target:     "node",
		Rules: []task.AffinityRule{
			{
				Weight: 1,
				MatchExpressions: []task.MatchExpression{
					{Key: "gpu-type", Operator: "In", Values: []string{"A100"}},
				},
			},
		},
	}
	result := buildTopologySpreadConstraints(a, "test-job")
	if result == nil {
		t.Fatal("expected non-nil result for node target spread")
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 constraint, got %d", len(result))
	}
	c, ok := result[0].(map[string]interface{})
	if !ok {
		t.Fatal("expected map[string]interface{}")
	}
	if m, ok := c["maxSkew"].(int64); !ok || m != 1 {
		t.Errorf("expected maxSkew=1, got %v", c["maxSkew"])
	}
	if tk, ok := c["topologyKey"].(string); !ok || tk != "kubernetes.io/hostname" {
		t.Errorf("expected topologyKey=kubernetes.io/hostname, got %v", c["topologyKey"])
	}
	if wu, ok := c["whenUnsatisfiable"].(string); !ok || wu != "ScheduleAnyway" {
		t.Errorf("expected whenUnsatisfiable=ScheduleAnyway for preferred, got %v", c["whenUnsatisfiable"])
	}
	ls, ok := c["labelSelector"].(map[string]interface{})
	if !ok {
		t.Fatal("expected labelSelector map")
	}
	ml, ok := ls["matchLabels"].(map[string]interface{})
	if !ok {
		t.Fatal("expected matchLabels map")
	}
	if v, ok := ml["training.kubeflow.org/job-name"].(string); !ok || v != "test-job" {
		t.Errorf("expected labelSelector matchLabels job-name=test-job, got %v", ml)
	}
}

func TestBuildTopologySpreadConstraints_NodeTargetSpreadRequired(t *testing.T) {
	a := &task.Affinity{
		Policy:     "spread",
		Constraint: "required",
		Target:     "node",
		Rules: []task.AffinityRule{
			{Weight: 1},
		},
	}
	result := buildTopologySpreadConstraints(a, "test-job")
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	c := result[0].(map[string]interface{})
	if wu, ok := c["whenUnsatisfiable"].(string); !ok || wu != "DoNotSchedule" {
		t.Errorf("expected whenUnsatisfiable=DoNotSchedule for required, got %v", c["whenUnsatisfiable"])
	}
}

func TestBuildTopologySpreadConstraints_NodeTargetSpreadCustomTopologyKey(t *testing.T) {
	// topology_key is pod-only (task.Validate rejects it for node target);
	// the builder tolerates it and TSC always uses kubernetes.io/hostname.
	a := &task.Affinity{
		Policy:     "spread",
		Constraint: "preferred",
		Target:     "node",
		Rules: []task.AffinityRule{
			{Weight: 1, TopologyKey: "topology.kubernetes.io/zone"},
		},
	}
	result := buildTopologySpreadConstraints(a, "test-job")
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 constraint, got %d", len(result))
	}
	c := result[0].(map[string]interface{})
	if tk, ok := c["topologyKey"].(string); !ok || tk != "kubernetes.io/hostname" {
		t.Errorf("expected topologyKey=kubernetes.io/hostname (hardcoded), got %v", c["topologyKey"])
	}
}

func TestBuildTopologySpreadConstraints_NodeTargetBinpack(t *testing.T) {
	a := &task.Affinity{
		Policy: "binpack",
		Target: "node",
		Rules:  []task.AffinityRule{{Weight: 1}},
	}
	result := buildTopologySpreadConstraints(a, "test-job")
	if result != nil {
		t.Errorf("expected nil for binpack policy, got %v", result)
	}
}

func TestBuildTopologySpreadConstraints_PodTarget(t *testing.T) {
	a := &task.Affinity{
		Policy: "spread",
		Target: "pod",
		Rules:  []task.AffinityRule{{Weight: 1}},
	}
	result := buildTopologySpreadConstraints(a, "test-job")
	if result != nil {
		t.Errorf("expected nil for pod target, got %v", result)
	}
}

func TestBuildTopologySpreadConstraints_MultipleRules(t *testing.T) {
	a := &task.Affinity{
		Policy: "spread",
		Target: "node",
		Rules: []task.AffinityRule{
			{Weight: 1, TopologyKey: "kubernetes.io/hostname"},
			{Weight: 1, TopologyKey: "topology.kubernetes.io/zone"},
		},
	}
	result := buildTopologySpreadConstraints(a, "test-job")
	if len(result) != 1 {
		t.Fatalf("expected 1 constraint (single TSC regardless of rule count), got %d", len(result))
	}
	c := result[0].(map[string]interface{})
	if tk, ok := c["topologyKey"].(string); !ok || tk != "kubernetes.io/hostname" {
		t.Errorf("expected topologyKey=kubernetes.io/hostname, got %v", c["topologyKey"])
	}
}

func TestBuildPodSpec_NodeTargetSpreadHasTopologySpreadConstraints(t *testing.T) {
	job := &task.Task{
		Name:  "test-job",
		Image: "test:latest",
		Framework: task.Framework{
			Name: "mpi",
		},
		Scheduling: task.Scheduling{
			Affinity: &task.Affinity{
				Policy:     "spread",
				Constraint: "preferred",
				Target:     "node",
				Rules: []task.AffinityRule{
					{
						Weight: 1,
						MatchExpressions: []task.MatchExpression{
							{Key: "gpu-type", Operator: "In", Values: []string{"A100"}},
						},
					},
				},
			},
		},
	}
	container := map[string]interface{}{
		"name":  "mpi",
		"image": "test:latest",
	}
	podSpec, err := buildPodSpec(job, container)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify nodeAffinity exists with positive (non-negated) operators
	affinity, ok := podSpec["affinity"].(map[string]interface{})
	if !ok {
		t.Fatal("expected affinity in pod spec")
	}
	nodeAffinity, ok := affinity["nodeAffinity"].(map[string]interface{})
	if !ok {
		t.Fatal("expected nodeAffinity")
	}
	preferred, ok := nodeAffinity["preferredDuringSchedulingIgnoredDuringExecution"].([]interface{})
	if !ok || len(preferred) != 1 {
		t.Fatalf("expected 1 preferred term, got %v", nodeAffinity)
	}
	term := preferred[0].(map[string]interface{})
	pref := term["preference"].(map[string]interface{})
	exprs := pref["matchExpressions"].([]interface{})
	for _, e := range exprs {
		expr := e.(map[string]interface{})
		if op, _ := expr["operator"].(string); op == "NotIn" {
			t.Error("expected In operator to be preserved, got NotIn")
		}
	}

	// Verify topologySpreadConstraints exists
	tsc, ok := podSpec["topologySpreadConstraints"].([]interface{})
	if !ok {
		t.Fatal("expected topologySpreadConstraints in pod spec")
	}
	if len(tsc) != 1 {
		t.Fatalf("expected 1 topology spread constraint, got %d", len(tsc))
	}
	c := tsc[0].(map[string]interface{})
	if tk, ok := c["topologyKey"].(string); !ok || tk != "kubernetes.io/hostname" {
		t.Errorf("expected topologyKey=kubernetes.io/hostname, got %v", c["topologyKey"])
	}
	if wu, ok := c["whenUnsatisfiable"].(string); !ok || wu != "ScheduleAnyway" {
		t.Errorf("expected whenUnsatisfiable=ScheduleAnyway, got %v", c["whenUnsatisfiable"])
	}
	ls := c["labelSelector"].(map[string]interface{})
	ml := ls["matchLabels"].(map[string]interface{})
	if v, ok := ml["training.kubeflow.org/job-name"].(string); !ok || v != "test-job" {
		t.Errorf("expected job-name=test-job, got %v", ml)
	}
}

func TestBuildPodSpec_NodeTargetBinpackNoTopologySpreadConstraints(t *testing.T) {
	job := &task.Task{
		Name:  "test-job",
		Image: "test:latest",
		Framework: task.Framework{
			Name: "mpi",
		},
		Scheduling: task.Scheduling{
			Affinity: &task.Affinity{
				Policy: "binpack",
				Target: "node",
				Rules: []task.AffinityRule{
					{Weight: 1},
				},
			},
		},
	}
	container := map[string]interface{}{
		"name":  "mpi",
		"image": "test:latest",
	}
	podSpec, err := buildPodSpec(job, container)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// affinity should exist with podAffinity (binpack packing)
	affinity, ok := podSpec["affinity"].(map[string]interface{})
	if !ok {
		t.Fatal("expected affinity in pod spec for binpack")
	}
	if _, ok := affinity["podAffinity"].(map[string]interface{}); !ok {
		t.Fatal("expected podAffinity for node binpack")
	}

	// topologySpreadConstraints should NOT exist
	if _, ok := podSpec["topologySpreadConstraints"]; ok {
		t.Error("expected no topologySpreadConstraints for binpack policy")
	}
}

func TestBuildAffinity_DefaultTopologyKeyForPodTarget(t *testing.T) {
	a := &task.Affinity{
		Target: "pod",
		Policy: "binpack",
		Rules: []task.AffinityRule{
			{
				Weight:      50,
				MatchLabels: map[string]string{"app": "test"},
				// TopologyKey intentionally omitted — should default to kubernetes.io/hostname
			},
		},
	}
	result := buildAffinity(a, "test-job")
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	assertK8sAffinityValid(t, result)

	// binpack → podAffinity; default constraint="" → preferred
	podAff, ok := result["podAffinity"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected podAffinity, got keys: %v", mapKeys(result))
	}
	preferred, ok := podAff["preferredDuringSchedulingIgnoredDuringExecution"].([]interface{})
	if !ok || len(preferred) != 1 {
		t.Fatalf("expected 1 preferred term, got %v", podAff)
	}
	wTerm := preferred[0].(map[string]interface{})
	term := wTerm["podAffinityTerm"].(map[string]interface{})
	if tk, ok := term["topologyKey"].(string); !ok || tk != "kubernetes.io/hostname" {
		t.Errorf("expected topologyKey=kubernetes.io/hostname (default), got %v", term["topologyKey"])
	}
}

func mapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func TestBuildTopologySpreadConstraints_NoRules(t *testing.T) {
	a := &task.Affinity{
		Policy:     "spread",
		Constraint: "preferred",
		Target:     "node",
		// No rules — should still generate 1 constraint
	}
	result := buildTopologySpreadConstraints(a, "test-job")
	if result == nil {
		t.Fatal("expected non-nil result for node spread without rules")
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 constraint without rules, got %d", len(result))
	}
	c := result[0].(map[string]interface{})
	if tk, ok := c["topologyKey"].(string); !ok || tk != "kubernetes.io/hostname" {
		t.Errorf("expected topologyKey=kubernetes.io/hostname, got %v", c["topologyKey"])
	}
}

func TestBuildAffinity_NodeBinpackNoRules(t *testing.T) {
	a := &task.Affinity{
		Policy: "binpack",
		Target: "node",
		// No rules — should still generate podAffinity
	}
	result := buildAffinity(a, "test-job")
	if result == nil {
		t.Fatal("expected non-nil result for node binpack without rules")
	}
	assertK8sAffinityValid(t, result)

	podAff, ok := result["podAffinity"].(map[string]interface{})
	if !ok {
		t.Fatal("expected podAffinity for node binpack")
	}
	preferred, ok := podAff["preferredDuringSchedulingIgnoredDuringExecution"].([]interface{})
	if !ok || len(preferred) != 1 {
		t.Fatalf("expected 1 preferred term, got %v", podAff)
	}
	wTerm := preferred[0].(map[string]interface{})
	if w, ok := wTerm["weight"].(int64); !ok || w != 100 {
		t.Errorf("expected default weight=100, got %v", wTerm["weight"])
	}
	term := wTerm["podAffinityTerm"].(map[string]interface{})
	if tk, ok := term["topologyKey"].(string); !ok || tk != "kubernetes.io/hostname" {
		t.Errorf("expected topologyKey=kubernetes.io/hostname, got %v", term["topologyKey"])
	}
	ls := term["labelSelector"].(map[string]interface{})
	ml := ls["matchLabels"].(map[string]interface{})
	if v, ok := ml["training.kubeflow.org/job-name"].(string); !ok || v != "test-job" {
		t.Errorf("expected job-name=test-job, got %v", ml)
	}
}

func TestBuildAffinity_NodeBinpackRequired(t *testing.T) {
	a := &task.Affinity{
		Policy:     "binpack",
		Constraint: "required",
		Target:     "node",
	}
	result := buildAffinity(a, "test-job")
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	assertK8sAffinityValid(t, result)

	podAff, ok := result["podAffinity"].(map[string]interface{})
	if !ok {
		t.Fatal("expected podAffinity")
	}
	required, ok := podAff["requiredDuringSchedulingIgnoredDuringExecution"].([]interface{})
	if !ok || len(required) != 1 {
		t.Fatalf("expected 1 required term, got %v", podAff)
	}
	term := required[0].(map[string]interface{})
	if tk, ok := term["topologyKey"].(string); !ok || tk != "kubernetes.io/hostname" {
		t.Errorf("expected topologyKey=kubernetes.io/hostname, got %v", term["topologyKey"])
	}
}

func TestBuildAffinity_NodeBinpackWithRules(t *testing.T) {
	a := &task.Affinity{
		Policy:     "binpack",
		Constraint: "preferred",
		Target:     "node",
		Rules: []task.AffinityRule{
			{
				Weight:      50,
				MatchLabels: map[string]string{"gpu": "a100"},
			},
		},
	}
	result := buildAffinity(a, "test-job")
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	assertK8sAffinityValid(t, result)

	// Should have both podAffinity (packing) and nodeAffinity (node selection)
	podAff, ok := result["podAffinity"].(map[string]interface{})
	if !ok {
		t.Fatal("expected podAffinity for binpack")
	}
	preferred := podAff["preferredDuringSchedulingIgnoredDuringExecution"].([]interface{})
	wTerm := preferred[0].(map[string]interface{})
	if w, ok := wTerm["weight"].(int64); !ok || w != 100 {
		t.Errorf("expected self-affinity weight=100 (decoupled from rule weight), got %v", wTerm["weight"])
	}

	nodeAff, ok := result["nodeAffinity"].(map[string]interface{})
	if !ok {
		t.Fatal("expected nodeAffinity from rules")
	}
	nodePreferred := nodeAff["preferredDuringSchedulingIgnoredDuringExecution"].([]interface{})
	if len(nodePreferred) != 1 {
		t.Fatalf("expected 1 node preferred term, got %d", len(nodePreferred))
	}
	nodeTerm := nodePreferred[0].(map[string]interface{})
	if w, ok := nodeTerm["weight"].(int64); !ok || w != 50 {
		t.Errorf("expected node preference weight=50 from rule, got %v", nodeTerm["weight"])
	}
}

func TestBuildAffinity_NodeSpreadNoRules(t *testing.T) {
	a := &task.Affinity{
		Policy: "spread",
		Target: "node",
		// No rules — buildAffinity returns nil (TSC generated separately in buildPodSpec)
	}
	result := buildAffinity(a, "test-job")
	if result != nil {
		t.Errorf("expected nil for node spread without rules (TSC generated separately), got %v", result)
	}
}
