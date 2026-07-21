package provider

import (
	"strings"
	"testing"

	"github.com/kubeflow/arena/pkg/task"
)

func TestBuildAffinity_Nil(t *testing.T) {
	result, err := buildAffinity(nil, "test-job")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil for nil affinity, got %v", result)
	}
}

func TestBuildAffinity_Empty(t *testing.T) {
	a := &task.Affinity{}
	result, err := buildAffinity(a, "test-job")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil for empty affinity, got %v", result)
	}
}

func TestBuildAffinity_PolicyOnlyNoRules(t *testing.T) {
	// Policy without rules should return an error
	a := &task.Affinity{
		Policy: "spread",
		Target: "pod",
	}
	result, err := buildAffinity(a, "test-job")
	if err == nil {
		t.Fatalf("expected error for policy without rules, got result: %v", result)
	}
	if result != nil {
		t.Errorf("expected nil result on error, got %v", result)
	}
}

func TestBuildAffinity_RulesWithPodTarget(t *testing.T) {
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
	result, err := buildAffinity(a, "test-job")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result for rules with pod target")
	}
	// Should generate podAffinity or podAntiAffinity
	if _, ok := result["podAffinity"]; !ok {
		if _, ok := result["podAntiAffinity"]; !ok {
			t.Errorf("expected podAffinity or podAntiAffinity, got %v", result)
		}
	}
}

func TestBuildAffinity_RulesWithNodeTarget(t *testing.T) {
	a := &task.Affinity{
		Target: "node",
		Rules: []task.AffinityRule{
			{
				MatchExpressions: []task.MatchExpression{
					{Key: "gpu-type", Operator: "In", Values: []string{"A100"}},
				},
			},
		},
	}
	result, err := buildAffinity(a, "test-job")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result for rules with node target")
	}
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
	terms, err := buildPodAffinityTerms(rules, "preferred")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(terms) != 1 {
		t.Fatalf("expected 1 term, got %d", len(terms))
	}
	weightedTerm, ok := terms[0].(map[string]interface{})
	if !ok {
		t.Fatal("expected map[string]interface{} for weighted term")
	}
	if w, ok := weightedTerm["weight"].(int); !ok || w != 50 {
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

func TestBuildPodAffinityTerms_PreferredWithInvalidWeight_Zero(t *testing.T) {
	rules := []task.AffinityRule{
		{TopologyKey: "kubernetes.io/hostname", Weight: 0},
	}
	_, err := buildPodAffinityTerms(rules, "preferred")
	if err == nil {
		t.Fatal("expected error for weight=0, got nil")
	}
	if !strings.Contains(err.Error(), "must be 1-100") {
		t.Errorf("error should mention valid range, got: %v", err)
	}
}

func TestBuildPodAffinityTerms_PreferredWithInvalidWeight_TooHigh(t *testing.T) {
	rules := []task.AffinityRule{
		{TopologyKey: "kubernetes.io/hostname", Weight: 101},
	}
	_, err := buildPodAffinityTerms(rules, "preferred")
	if err == nil {
		t.Fatal("expected error for weight=101, got nil")
	}
	if !strings.Contains(err.Error(), "101") {
		t.Errorf("error should mention the invalid weight value, got: %v", err)
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
	terms, err := buildPodAffinityTerms(rules, "required")
	if err != nil {
		t.Fatalf("unexpected error in required mode: %v", err)
	}
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
	result, err := buildAffinity(a, "test-job")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
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
	if w, ok := wt["weight"].(int); !ok || w != 50 {
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

func TestBuildAffinity_PreferredInvalidWeight_ReturnsError(t *testing.T) {
	a := &task.Affinity{
		Policy:     "spread",
		Constraint: "preferred",
		Target:     "pod",
		Rules: []task.AffinityRule{
			{TopologyKey: "kubernetes.io/hostname", Weight: 0},
		},
	}
	_, err := buildAffinity(a, "test-job")
	if err == nil {
		t.Fatal("expected error for invalid weight through buildAffinity, got nil")
	}
	if !strings.Contains(err.Error(), "must be 1-100") {
		t.Errorf("error should mention valid range, got: %v", err)
	}
}
