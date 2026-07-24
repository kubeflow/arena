package provider

import (
	"fmt"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/kubeflow/arena/pkg/constants"
	log "github.com/kubeflow/arena/pkg/log"
	"github.com/kubeflow/arena/pkg/task"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Provider defines the interface for generating training job CRDs from a Task specification.
type Provider interface {
	// BuildCRD generates an unstructured Kubernetes CRD object for the given task.
	BuildCRD(task *task.Task) (*unstructured.Unstructured, error)
	// BuildRBAC returns auxiliary RBAC resources (ServiceAccount, Role, RoleBinding)
	// that need to be created alongside the CRD. Returns nil if no RBAC resources
	// are needed for this framework.
	BuildRBAC(task *task.Task, ownerRef metav1.OwnerReference) ([]*unstructured.Unstructured, error)
	// GetJobType returns the CRD kind this provider generates (e.g. "PyTorchJob").
	GetJobType() string
	// GetLogPodSelector returns the label selector string for finding the primary
	// pod to stream logs from (e.g. master, chief, or launcher).
	GetLogPodSelector(jobName string) string
	// GetJobPodSelector returns the label selector string for finding all pods
	// belonging to a job (across all replica types).
	GetJobPodSelector(jobName string) string
}

// buildLabelSelector constructs a Kubernetes label selector string from the
// given label key-value pairs using metav1.LabelSelector +
// metav1.FormatLabelSelector. This ensures label values are handled through
// the structured Kubernetes API rather than raw string interpolation, so
// escaping is applied consistently if values ever come from a less
// constrained source.
func buildLabelSelector(matchLabels map[string]string) string {
	selector := &metav1.LabelSelector{
		MatchLabels: matchLabels,
	}
	return metav1.FormatLabelSelector(selector)
}

// effectiveRun returns the role-level run if non-empty, otherwise falls back to the top-level run.
func effectiveRun(t *task.Task, roleRun string) string {
	if roleRun != "" {
		return roleRun
	}
	return t.Run
}

// buildCommandArgs returns the command and args arrays for a container using the
// run+shell model: command: [shell, "-c"], args: [run].
func buildCommandArgs(t *task.Task, run string) ([]interface{}, []interface{}) {
	shell := t.EffectiveShell()
	command := []interface{}{shell, "-c"}
	args := []interface{}{run}
	return command, args
}

// buildEnvVars merges global task envs with role-specific envs and returns a
// sorted (by name) list of K8s env var maps. Role envs override global envs.
func buildEnvVars(t *task.Task, roleEnvs map[string]task.EnvValue) []interface{} {
	merged := make(map[string]task.EnvValue)
	for k, v := range t.Envs {
		merged[k] = v
	}
	for k, v := range roleEnvs {
		merged[k] = v
	}
	if len(merged) == 0 {
		return nil
	}

	// Sort keys for deterministic output
	keys := make([]string, 0, len(merged))
	for k := range merged {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	env := make([]interface{}, 0, len(keys))
	for _, name := range keys {
		env = append(env, buildEnvVar(name, merged[name]))
	}
	return env
}

// buildEnvVar creates a single K8s env var map from an EnvValue.
func buildEnvVar(name string, e task.EnvValue) map[string]interface{} {
	envVar := map[string]interface{}{"name": name}
	switch {
	case e.Secret != nil:
		envVar["valueFrom"] = map[string]interface{}{
			"secretKeyRef": map[string]interface{}{
				"name": e.Secret.Name,
				"key":  e.Secret.Key,
			},
		}
	case e.ConfigMap != nil:
		envVar["valueFrom"] = map[string]interface{}{
			"configMapKeyRef": map[string]interface{}{
				"name": e.ConfigMap.Name,
				"key":  e.ConfigMap.Key,
			},
		}
	default:
		envVar["value"] = e.Value
	}
	return envVar
}

// BuildVolumes creates volumes and volumeMounts from the task's Storages.
func BuildVolumes(t *task.Task) ([]interface{}, []interface{}) {
	if len(t.Storages) == 0 {
		return nil, nil
	}
	volumes := make([]interface{}, 0, len(t.Storages))
	mounts := make([]interface{}, 0, len(t.Storages))
	for _, s := range t.Storages {
		vol := map[string]interface{}{"name": s.Name}
		mountPath := s.MountPath
		if mountPath == "" && s.SHM != "" {
			mountPath = constants.DefaultSHMMountPath
		}
		mount := map[string]interface{}{
			"name":      s.Name,
			"mountPath": mountPath,
		}
		if s.SubPath != "" {
			mount["subPath"] = s.SubPath
		}
		switch {
		case s.PVC != "":
			vol["persistentVolumeClaim"] = map[string]interface{}{
				"claimName": s.PVC,
			}
		case s.SHM != "":
			vol["emptyDir"] = map[string]interface{}{
				"medium":    constants.EmptyDirMediumMemory,
				"sizeLimit": s.SHM,
			}
		case s.Tmp != "":
			vol["emptyDir"] = map[string]interface{}{
				"sizeLimit": s.Tmp,
			}
		case s.HostPath != "":
			vol["hostPath"] = map[string]interface{}{
				"path": s.HostPath,
			}
		case s.ConfigMap != "":
			vol["configMap"] = map[string]interface{}{
				"name": s.ConfigMap,
			}
			if s.Key != "" {
				mount["subPath"] = s.Key
			}
		case s.Secret != "":
			vol["secret"] = map[string]interface{}{
				"secretName": s.Secret,
			}
			if s.Key != "" {
				mount["subPath"] = s.Key
			}
		}
		volumes = append(volumes, vol)
		mounts = append(mounts, mount)
	}
	return volumes, mounts
}

// resolveMounts builds K8s volumeMount specs from Mount entries by looking up
// the referenced Storage by name. Mount fields override the Storage defaults:
// - mount_path falls back to Storage.MountPath when empty
// - sub_path falls back to Storage.SubPath when empty
// - sub_path falls back to Storage.Key for ConfigMap/Secret storages
func resolveMounts(mounts []task.Mount, storages []task.Storage) []map[string]interface{} {
	if len(mounts) == 0 {
		return nil
	}
	storageMap := make(map[string]task.Storage, len(storages))
	for _, s := range storages {
		storageMap[s.Name] = s
	}
	result := make([]map[string]interface{}, 0, len(mounts))
	for _, m := range mounts {
		mountPath := m.MountPath
		subPath := m.SubPath
		if st, ok := storageMap[m.Name]; ok {
			if mountPath == "" {
				mountPath = st.MountPath
			}
			if subPath == "" {
				subPath = st.SubPath
			}
			subPathEmpty := subPath == ""
			hasRefStorage := st.ConfigMap != "" || st.Secret != ""
			hasKey := st.Key != ""
			if subPathEmpty && hasRefStorage && hasKey {
				subPath = st.Key
			}
		}
		entry := map[string]interface{}{
			"name":      m.Name,
			"mountPath": mountPath,
		}
		if subPath != "" {
			entry["subPath"] = subPath
		}
		result = append(result, entry)
	}
	return result
}

// ResolveContainerMounts returns volumeMounts for any container following the
// three-case contract:
// - No storages: nil
// - Storages, no mounts: all storages become mounts (BuildVolumes fallback)
// - Storages + mounts: only listed storages get mounts (resolveMounts overrides)
func ResolveContainerMounts(mounts []task.Mount, t *task.Task) []interface{} {
	if len(t.Storages) == 0 {
		return nil
	}
	if len(mounts) > 0 {
		resolved := resolveMounts(mounts, t.Storages)
		result := make([]interface{}, len(resolved))
		for i, m := range resolved {
			result[i] = m
		}
		return result
	}
	_, allMounts := BuildVolumes(t)
	return allMounts
}

// buildScheduling applies scheduling fields to a podSpec map.
func buildScheduling(t *task.Task, podSpec map[string]interface{}) {
	if len(t.Scheduling.NodeSelector) > 0 {
		ns := make(map[string]interface{}, len(t.Scheduling.NodeSelector))
		for k, v := range t.Scheduling.NodeSelector {
			ns[k] = v
		}
		podSpec["nodeSelector"] = ns
	}
	if len(t.Scheduling.Tolerations) > 0 {
		tolerations := make([]interface{}, 0, len(t.Scheduling.Tolerations))
		for _, tol := range t.Scheduling.Tolerations {
			tolMap := map[string]interface{}{}
			if tol.Key != "" {
				tolMap["key"] = tol.Key
			}
			if tol.Operator != "" {
				tolMap["operator"] = tol.Operator
			}
			if tol.Value != "" {
				tolMap["value"] = tol.Value
			}
			if tol.Effect != "" {
				tolMap["effect"] = tol.Effect
			}
			if tol.TolerationSeconds != nil {
				tolMap["tolerationSeconds"] = *tol.TolerationSeconds
			}
			tolerations = append(tolerations, tolMap)
		}
		podSpec["tolerations"] = tolerations
	}
	if t.Scheduling.PriorityClassName != "" {
		podSpec["priorityClassName"] = t.Scheduling.PriorityClassName
	}
	if t.Scheduling.Priority > 0 {
		podSpec["priority"] = int64(t.Scheduling.Priority)
	}
	if t.Scheduling.SchedulerName != "" {
		podSpec["schedulerName"] = t.Scheduling.SchedulerName
	}
	if t.HostNetwork {
		podSpec["hostNetwork"] = true
	}
	if t.HostIPC {
		podSpec["hostIPC"] = true
	}
	if t.HostPID {
		podSpec["hostPID"] = true
	}
	if t.ServiceAccount != "" {
		podSpec["serviceAccountName"] = t.ServiceAccount
	}
	if len(t.ImagePullSecrets) > 0 {
		secrets := make([]interface{}, 0, len(t.ImagePullSecrets))
		for _, s := range t.ImagePullSecrets {
			secrets = append(secrets, map[string]interface{}{"name": s})
		}
		podSpec["imagePullSecrets"] = secrets
	}
}

// buildRunPolicy creates a RunPolicy map from the task's Lifecycle settings.
// Returns an error if duration fields cannot be parsed.
func buildRunPolicy(t *task.Task) (map[string]interface{}, error) {
	policy := map[string]interface{}{}
	if t.Lifecycle.CleanPodPolicy != "" {
		policy["cleanPodPolicy"] = t.Lifecycle.CleanPodPolicy
	}
	if t.Lifecycle.ActiveDeadline != "" {
		d, err := task.ParseDuration(t.Lifecycle.ActiveDeadline)
		if err != nil {
			return nil, fmt.Errorf("invalid active_deadline: %w", err)
		}
		if d > 0 {
			policy["activeDeadlineSeconds"] = int64(d.Seconds())
		}
	}
	if t.Lifecycle.TTLAfterFinished != "" {
		d, err := task.ParseDuration(t.Lifecycle.TTLAfterFinished)
		if err != nil {
			return nil, fmt.Errorf("invalid ttl_after_finished: %w", err)
		}
		if d > 0 {
			policy["ttlSecondsAfterFinished"] = int64(d.Seconds())
		}
	}
	if t.Lifecycle.BackoffLimit != nil {
		policy["backoffLimit"] = int64(*t.Lifecycle.BackoffLimit)
	}
	if t.Lifecycle.Suspend != nil {
		policy["suspend"] = *t.Lifecycle.Suspend
	}
	if t.Lifecycle.ManagedBy != "" {
		policy["managedBy"] = t.Lifecycle.ManagedBy
	}
	if sp := buildSchedulingPolicy(t); sp != nil {
		policy["schedulingPolicy"] = sp
	}
	if len(policy) == 0 {
		return nil, nil
	}
	return policy, nil
}

// buildSchedulingPolicy creates a SchedulingPolicy map from the task's Scheduling settings.
// Maps gang scheduling parameters (queue, priorityClass) to CRD-native runPolicy.schedulingPolicy fields.
// When gang scheduling is enabled, minAvailable is set to the total number of replicas
// across all roles so the scheduler knows the full gang size.
func buildSchedulingPolicy(t *task.Task) map[string]interface{} {
	s := t.Scheduling
	gangDisabled := !s.Gang.Enabled
	noQueue := s.Queue == ""
	noPriorityClass := s.PriorityClassName == ""
	if gangDisabled && noQueue && noPriorityClass {
		return nil
	}
	sp := map[string]interface{}{}
	if s.Queue != "" {
		sp["queue"] = s.Queue
	}
	if s.PriorityClassName != "" {
		sp["priorityClass"] = s.PriorityClassName
	}
	if s.Gang.Enabled {
		total := totalReplicas(t)
		if total < 1 {
			total = 1
		}
		sp["minAvailable"] = total
	}
	if len(sp) == 0 {
		return nil
	}
	return sp
}

// totalReplicas returns the sum of replicas across all roles defined in the task.
// Constrained roles (master, chief, launcher, evaluator) contribute 1 when present.
// Unconstrained roles (worker, ps) contribute their configured replica count.
// Implicit roles that providers always create are also counted:
//   - MPI-family frameworks (mpi, horovod, deepspeed) always create a launcher (1 replica)
//     even when t.Launcher is nil.
//   - PyTorch always creates a master (1 replica) when worker is present, even when t.Master is nil.
func totalReplicas(t *task.Task) int64 {
	var total int64
	if t.Worker != nil {
		total += int64(t.Worker.Replicas)
	}
	if t.Master != nil {
		total++
	}
	if t.Chief != nil {
		total++
	}
	if t.PS != nil {
		total += int64(t.PS.Replicas)
	}
	if t.Evaluator != nil {
		total++
	}
	if t.Launcher != nil {
		total++
	}

	// Account for implicit roles that providers always create even when the
	// corresponding task section is nil. Without these corrections, gang
	// scheduling minAvailable would undercount the actual pod total.
	if t.Launcher == nil && isMPIFamily(t) {
		total++ // MPI-family providers always create a 1-replica launcher
	}
	masterNil := t.Master == nil
	workerPresent := t.Worker != nil
	isPyTorch := t.Framework.Name == constants.FrameworkPyTorch
	if masterNil && workerPresent && isPyTorch {
		total++ // PyTorch provider always creates a 1-replica master when worker is present
	}

	return total
}

// isMPIFamily reports whether the task uses an MPI-family framework
// (mpi, horovod, or deepspeed), all of which share the launcher/worker split.
func isMPIFamily(t *task.Task) bool {
	return t.Framework.Name == constants.FrameworkMPI ||
		t.Framework.Name == constants.FrameworkHorovod ||
		t.Framework.Name == constants.FrameworkDeepSpeed
}

// buildAffinity creates a K8s Affinity map from the task's Affinity settings.
// Supports two modes:
// - Policy mode: policy×constraint generates podAffinity/podAntiAffinity or nodeAffinity (requires rules)
// - Rules mode: custom rules applied to pod or node affinity based on target
// Policy without rules is a no-op; task.Validate rejects that configuration.
func buildAffinity(a *task.Affinity, _ string) (map[string]interface{}, error) {
	if a == nil {
		return nil, nil
	}
	if a.Policy == "none" {
		return nil, nil
	}
	if a.Policy == "" && len(a.Rules) == 0 {
		return nil, nil
	}

	affinity := map[string]interface{}{}

	if len(a.Rules) > 0 {
		switch a.Target {
		case "pod":
			if err := applyPodRules(affinity, a); err != nil {
				return nil, err
			}
		case "node":
			applyNodeRules(affinity, a)
		}
	}

	if len(affinity) == 0 {
		return nil, nil
	}
	return affinity, nil
}

// applyPodRules applies affinity rules to podAffinity or podAntiAffinity based on policy.
func applyPodRules(affinity map[string]interface{}, a *task.Affinity) error {
	constraint := a.Constraint
	if constraint == "" {
		constraint = "preferred"
	}

	affinityType := "podAffinity"
	if a.Policy == "spread" {
		affinityType = "podAntiAffinity"
	}
	mode := "required"
	fieldKey := "requiredDuringSchedulingIgnoredDuringExecution"
	if constraint == "preferred" {
		mode = "preferred"
		fieldKey = "preferredDuringSchedulingIgnoredDuringExecution"
	}
	terms, err := buildPodAffinityTerms(a.Rules, mode)
	if err != nil {
		return err
	}
	affinity[affinityType] = map[string]interface{}{fieldKey: terms}
	return nil
}

// applyNodeRules applies affinity rules to nodeAffinity based on policy and constraint.
// For binpack policy, rules attract pods to matching nodes (positive affinity).
// For spread policy, operators are negated to repel pods from matching nodes.
func applyNodeRules(affinity map[string]interface{}, a *task.Affinity) {
	constraint := a.Constraint
	if constraint == "" {
		constraint = "preferred"
	}

	terms := buildNodeSelectorTerms(a.Rules)
	if a.Policy == "spread" {
		negateNodeSelectorTerms(terms)
	}

	if constraint == "preferred" {
		affinity["nodeAffinity"] = map[string]interface{}{
			"preferredDuringSchedulingIgnoredDuringExecution": terms,
		}
	} else { // required
		affinity["nodeAffinity"] = map[string]interface{}{
			"requiredDuringSchedulingIgnoredDuringExecution": map[string]interface{}{
				"nodeSelectorTerms": terms,
			},
		}
	}
}

// negateNodeSelectorTerms negates operators in node selector terms for spread policy.
func negateNodeSelectorTerms(terms []interface{}) {
	for _, term := range terms {
		t, ok := term.(map[string]interface{})
		if !ok {
			continue
		}
		negateOperatorList(t, "matchExpressions")
		negateOperatorList(t, "matchFields")
	}
}

func negateOperatorList(term map[string]interface{}, field string) {
	exprs, ok := term[field].([]interface{})
	if !ok {
		return
	}
	for _, e := range exprs {
		expr, ok := e.(map[string]interface{})
		if !ok {
			continue
		}
		if op, ok := expr["operator"].(string); ok {
			expr["operator"] = negateNodeOperator(op)
		}
	}
}

func negateNodeOperator(op string) string {
	switch op {
	case "In":
		return "NotIn"
	case "NotIn":
		return "In"
	case "Exists":
		return "DoesNotExist"
	case "DoesNotExist":
		return "Exists"
	case "Gt":
		return "Lt"
	case "Lt":
		return "Gt"
	default:
		return op
	}
}

// buildAffinityTerm builds a single pod affinity term from an AffinityRule.
// In preferred mode, the term is wrapped in a WeightedPodAffinityTerm (weight at outer level).
// In required mode, the term is returned directly (no weight).
// Returns an error if preferred mode rules have weight outside 1-100.
func buildAffinityTerm(rule task.AffinityRule, mode string) (map[string]interface{}, error) {
	if mode == "preferred" {
		if rule.Weight < 1 || rule.Weight > 100 {
			return nil, fmt.Errorf("invalid affinity rule weight: %d (must be 1-100 for preferred scheduling)", rule.Weight)
		}
	}

	term := map[string]interface{}{
		"topologyKey": rule.TopologyKey,
	}
	if len(rule.MatchLabels) > 0 {
		term["labelSelector"] = map[string]interface{}{
			"matchLabels": rule.MatchLabels,
		}
	}
	// matchExpressions
	if len(rule.MatchExpressions) > 0 {
		labelSelector, _ := term["labelSelector"].(map[string]interface{})
		if labelSelector == nil {
			labelSelector = map[string]interface{}{}
			term["labelSelector"] = labelSelector
		}
		exprs := make([]interface{}, 0, len(rule.MatchExpressions))
		for _, e := range rule.MatchExpressions {
			expr := map[string]interface{}{
				"key":      e.Key,
				"operator": e.Operator,
			}
			if len(e.Values) > 0 {
				expr["values"] = e.Values
			}
			exprs = append(exprs, expr)
		}
		labelSelector["matchExpressions"] = exprs
	}
	// namespaces
	if len(rule.Namespaces) > 0 {
		term["namespaces"] = rule.Namespaces
	}
	// namespaceSelector
	if rule.NamespaceSelector != nil {
		ns := map[string]interface{}{}
		if len(rule.NamespaceSelector.MatchLabels) > 0 {
			ns["matchLabels"] = rule.NamespaceSelector.MatchLabels
		}
		if len(rule.NamespaceSelector.MatchExpressions) > 0 {
			exprs := make([]interface{}, 0, len(rule.NamespaceSelector.MatchExpressions))
			for _, e := range rule.NamespaceSelector.MatchExpressions {
				expr := map[string]interface{}{
					"key":      e.Key,
					"operator": e.Operator,
				}
				if len(e.Values) > 0 {
					expr["values"] = e.Values
				}
				exprs = append(exprs, expr)
			}
			ns["matchExpressions"] = exprs
		}
		term["namespaceSelector"] = ns
	}

	if mode == "preferred" {
		return map[string]interface{}{
			"weight":          rule.Weight,
			"podAffinityTerm": term,
		}, nil
	}
	return term, nil
}

// buildPodAffinityTerms converts AffinityRules to pod affinity terms.
// In preferred mode, returns WeightedPodAffinityTerm structures (weight at outer level).
// In required mode, returns PodAffinityTerm structures directly (no weight).
// Returns an error if preferred mode rules have weight outside 1-100.
func buildPodAffinityTerms(rules []task.AffinityRule, mode string) ([]interface{}, error) {
	terms := make([]interface{}, 0, len(rules))
	for _, rule := range rules {
		term, err := buildAffinityTerm(rule, mode)
		if err != nil {
			return nil, err
		}
		terms = append(terms, term)
	}
	return terms, nil
}

// buildNodeSelectorTerms converts AffinityRules to node selector terms.
func buildNodeSelectorTerms(rules []task.AffinityRule) []interface{} {
	terms := make([]interface{}, 0, len(rules))
	for _, rule := range rules {
		term := map[string]interface{}{}
		// Convert MatchExpressions
		if len(rule.MatchExpressions) > 0 {
			exprs := make([]interface{}, 0, len(rule.MatchExpressions))
			for _, e := range rule.MatchExpressions {
				expr := map[string]interface{}{
					"key":      e.Key,
					"operator": e.Operator,
				}
				if len(e.Values) > 0 {
					expr["values"] = e.Values
				}
				exprs = append(exprs, expr)
			}
			term["matchExpressions"] = exprs
		}
		// Convert MatchLabels to matchExpressions with In operator
		if len(rule.MatchLabels) > 0 {
			exprs := make([]interface{}, 0, len(rule.MatchLabels))
			for k, v := range rule.MatchLabels {
				exprs = append(exprs, map[string]interface{}{
					"key":      k,
					"operator": constants.AffinityOperatorIn,
					"values":   []string{v},
				})
			}
			if existing, ok := term["matchExpressions"]; ok {
				if existingArr, ok := existing.([]interface{}); ok {
					term["matchExpressions"] = append(existingArr, exprs...)
				} else {
					term["matchExpressions"] = exprs
				}
			} else {
				term["matchExpressions"] = exprs
			}
		}
		// Convert MatchFields
		if len(rule.MatchFields) > 0 {
			fields := make([]interface{}, 0, len(rule.MatchFields))
			for _, f := range rule.MatchFields {
				field := map[string]interface{}{
					"key":      f.Key,
					"operator": f.Operator,
				}
				if len(f.Values) > 0 {
					field["values"] = f.Values
				}
				fields = append(fields, field)
			}
			term["matchFields"] = fields
		}
		terms = append(terms, term)
	}
	return terms
}

// buildInitContainers creates init containers from task.Init and task.Sync.
// Sync-generated init containers (arena-sync-N) are added first, followed by user-defined init containers.
func buildInitContainers(t *task.Task) []map[string]interface{} {
	containers := make([]map[string]interface{}, 0, len(t.Init)+len(t.Sync))

	if syncInits := buildSyncInitContainers(t); len(syncInits) > 0 {
		containers = append(containers, syncInits...)
	}

	for _, init := range t.Init {
		shell := init.Shell
		if shell == "" {
			shell = t.Shell
		}
		if shell == "" {
			shell = constants.DefaultShell
		}
		container := map[string]interface{}{
			"name":    init.Name,
			"image":   init.Image,
			"command": []interface{}{shell, "-c"},
			"args":    []interface{}{init.Run},
		}

		// Add volume mounts from init[].mounts (resolved against storages)
		if mounts := ResolveContainerMounts(init.Mounts, t); len(mounts) > 0 {
			container["volumeMounts"] = mounts
		}

		containers = append(containers, container)
	}

	if len(containers) == 0 {
		return nil
	}
	return containers
}

// buildSyncInitContainers generates system init containers from all sync entries.
// Supports git-sync, rsync, and hdfs modes. Each sync entry gets its own init container
// named arena-sync-N. Volume mounts are resolved from storages via ResolveContainerMounts.
// The sync command target path is always local_path; a warning is printed to stderr
// when local_path does not match any resolved mount path.
func buildSyncInitContainers(t *task.Task) []map[string]interface{} {
	if len(t.Sync) == 0 {
		return nil
	}

	containers := make([]map[string]interface{}, 0, len(t.Sync))
	for i, s := range t.Sync {
		containerName := "arena-sync-" + strconv.Itoa(i)

		// Build volumeMounts from mounts (references storages by name)
		volumeMounts := ResolveContainerMounts(s.Mounts, t)

		switch {
		case s.Git != "":
			containers = append(containers, buildGitSyncContainer(s, containerName, volumeMounts, t.Envs))
		case s.Rsync != "":
			containers = append(containers, buildRsyncContainer(s, containerName, volumeMounts))
		case s.HDFS != "":
			containers = append(containers, buildHDFSSyncContainer(s, containerName, volumeMounts))
		}

		// Warn when local_path does not match any resolved mount path
		if s.LocalPath != "" && len(volumeMounts) > 0 {
			matched := false
			for _, vm := range volumeMounts {
				if m, ok := vm.(map[string]interface{}); ok {
					if mp, ok := m["mountPath"].(string); ok && mp == s.LocalPath {
						matched = true
						break
					}
				}
			}
			if !matched {
				log.Warning("sync local_path does not match any mount path; data will be written to ephemeral storage",
					"index", i, "local_path", s.LocalPath)
			}
		}
	}

	if len(containers) == 0 {
		return nil
	}
	return containers
}

// buildGitSyncContainer builds a git-sync init container from a SyncEntry.
func buildGitSyncContainer(sync task.SyncEntry, name string, volumeMounts []interface{}, userEnvs map[string]task.EnvValue) map[string]interface{} {
	projectName := extractGitProjectName(sync.Git)

	defaultEnvs := map[string]string{
		"GIT_SYNC_REPO":     sync.Git,
		"GIT_SYNC_DEST":     projectName,
		"GIT_SYNC_ROOT":     sync.LocalPath,
		"GIT_SYNC_ONE_TIME": "true",
	}
	if sync.Branch != "" {
		defaultEnvs["GIT_SYNC_BRANCH"] = sync.Branch
	}
	envList := buildEnvWithOverrides(defaultEnvs, userEnvs)

	image := sync.Image
	if image == "" {
		image = constants.DefaultGitSyncImage
	}

	c := map[string]interface{}{
		"name":            name,
		"image":           image,
		"imagePullPolicy": "Always",
		"env":             envList,
	}
	if len(volumeMounts) > 0 {
		c["volumeMounts"] = volumeMounts
	}
	return c
}

// buildRsyncContainer builds an rsync init container from a SyncEntry.
func buildRsyncContainer(sync task.SyncEntry, name string, volumeMounts []interface{}) map[string]interface{} {
	image := sync.Image
	if image == "" {
		image = constants.DefaultRsyncImage
	}

	c := map[string]interface{}{
		"name":            name,
		"image":           image,
		"imagePullPolicy": "Always",
		"command":         []interface{}{"rsync", "-avP", sync.Rsync, sync.LocalPath},
	}
	if len(volumeMounts) > 0 {
		c["volumeMounts"] = volumeMounts
	}
	return c
}

// buildHDFSSyncContainer builds an HDFS init container from a SyncEntry.
func buildHDFSSyncContainer(sync task.SyncEntry, name string, volumeMounts []interface{}) map[string]interface{} {
	image := sync.Image
	if image == "" {
		image = constants.DefaultHDFSImage
	}

	c := map[string]interface{}{
		"name":            name,
		"image":           image,
		"imagePullPolicy": "Always",
		"command":         []interface{}{"hdfs", "dfs", "-get", sync.HDFS, sync.LocalPath},
	}
	if len(volumeMounts) > 0 {
		c["volumeMounts"] = volumeMounts
	}
	return c
}

// buildEnvWithOverrides merges default envs with user envs.
// User envs with the same key override default values.
// User envs with secret/configmap refs are appended.
func buildEnvWithOverrides(defaults map[string]string, userEnvs map[string]task.EnvValue) []map[string]interface{} {
	// Start with defaults, userEnvs with same key override default value
	merged := make(map[string]string, len(defaults))
	for k, v := range defaults {
		merged[k] = v
	}

	// Track which keys are handled by user envs with refs
	refEnvs := make([]map[string]interface{}, 0, len(userEnvs))
	for k, e := range userEnvs {
		if e.Secret != nil || e.ConfigMap != nil {
			// Env with secret/configmap ref - add to ref list
			envVar := map[string]interface{}{"name": k}
			if e.Secret != nil {
				envVar["valueFrom"] = map[string]interface{}{
					"secretKeyRef": map[string]interface{}{
						"name": e.Secret.Name,
						"key":  e.Secret.Key,
					},
				}
			} else {
				envVar["valueFrom"] = map[string]interface{}{
					"configMapKeyRef": map[string]interface{}{
						"name": e.ConfigMap.Name,
						"key":  e.ConfigMap.Key,
					},
				}
			}
			refEnvs = append(refEnvs, envVar)
			// Remove from merged if it was a default (user ref overrides default plain value)
			delete(merged, k)
		} else {
			// Plain string value - override default
			merged[k] = e.Value
		}
	}

	// Output: merged plain values (sorted for deterministic output) + ref envs (sorted by name)
	result := make([]map[string]interface{}, 0, len(merged)+len(refEnvs))
	mergedKeys := make([]string, 0, len(merged))
	for k := range merged {
		mergedKeys = append(mergedKeys, k)
	}
	sort.Strings(mergedKeys)
	for _, k := range mergedKeys {
		result = append(result, map[string]interface{}{
			"name": k, "value": merged[k],
		})
	}
	sort.Slice(refEnvs, func(i, j int) bool {
		nameI, _ := refEnvs[i]["name"].(string)
		nameJ, _ := refEnvs[j]["name"].(string)
		return nameI < nameJ
	})
	result = append(result, refEnvs...)
	return result
}

// extractGitProjectName extracts project name from git URL.
// Example: "https://github.com/kubeflow/training-operator.git" -> "training-operator"
func extractGitProjectName(gitURL string) string {
	return path.Base(strings.TrimSuffix(gitURL, ".git"))
}

// buildMetadata creates a metadata map with name, namespace, labels, and annotations.
func buildMetadata(t *task.Task) map[string]interface{} {
	meta := map[string]interface{}{
		"name": t.Name,
	}
	if t.Namespace != "" {
		meta["namespace"] = t.Namespace
	}
	if len(t.Labels) > 0 {
		labels := make(map[string]interface{}, len(t.Labels))
		for k, v := range t.Labels {
			labels[k] = v
		}
		meta["labels"] = labels
	}
	if len(t.Annotations) > 0 {
		annotations := make(map[string]interface{}, len(t.Annotations))
		for k, v := range t.Annotations {
			annotations[k] = v
		}
		meta["annotations"] = annotations
	}
	return meta
}

// buildResources creates a Guaranteed QoS resource block (requests == limits).
func buildResources(r task.Resources) map[string]interface{} {
	if len(r) == 0 {
		return nil
	}
	reqs := make(map[string]interface{}, len(r))
	for k, v := range r {
		reqs[k] = v
	}
	lims := make(map[string]interface{}, len(r))
	for k, v := range r {
		lims[k] = v
	}
	return map[string]interface{}{
		"requests": reqs,
		"limits":   lims,
	}
}

// containerOptions holds parameters for buildContainer.
type containerOptions struct {
	Name      string
	Image     string
	Task      *task.Task
	Resources task.Resources
	RoleEnvs  map[string]task.EnvValue
	Run       string
	Mounts    []task.Mount
}

// buildContainer creates a container spec with the run+shell model.
func buildContainer(opts containerOptions) map[string]interface{} {
	container := map[string]interface{}{
		"name":  opts.Name,
		"image": opts.Image,
	}

	// Resources
	res := buildResources(opts.Resources)
	if res != nil {
		container["resources"] = res
	}

	// Command and args (run + shell model)
	if opts.Run != "" {
		cmd, args := buildCommandArgs(opts.Task, opts.Run)
		container["command"] = cmd
		container["args"] = args
	}

	// Environment variables
	envVars := buildEnvVars(opts.Task, opts.RoleEnvs)
	if len(envVars) > 0 {
		container["env"] = envVars
	}

	// Volume mounts (resolved via ResolveContainerMounts for per-container override support)
	containerMounts := ResolveContainerMounts(opts.Mounts, opts.Task)
	if len(containerMounts) > 0 {
		container["volumeMounts"] = containerMounts
	}

	// Image pull policy
	if opts.Task.ImagePullPolicy != "" {
		container["imagePullPolicy"] = opts.Task.ImagePullPolicy
	}

	// Working directory
	if opts.Task.WorkingDir != "" {
		container["workingDir"] = opts.Task.WorkingDir
	}

	return container
}

// buildPodSpec creates a pod spec with container, volumes, scheduling, affinity, and init containers.
func buildPodSpec(t *task.Task, container map[string]interface{}, includeVolumes bool) (map[string]interface{}, error) {
	podSpec := map[string]interface{}{
		"containers": []interface{}{container},
	}

	// Volumes from storages only (sync references storages, no separate sync volumes)
	volumes := []interface{}{}
	if includeVolumes {
		storageVols, _ := BuildVolumes(t)
		volumes = append(volumes, storageVols...)
	}
	if len(volumes) > 0 {
		podSpec["volumes"] = volumes
	}

	// Init containers (sync init + user-defined)
	initContainers := buildInitContainers(t)
	if len(initContainers) > 0 {
		podSpec["initContainers"] = initContainers
	}

	// Scheduling (nodeSelector, tolerations, priority, etc.)
	buildScheduling(t, podSpec)

	// Affinity
	affinity, err := buildAffinity(t.Scheduling.Affinity, t.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to build affinity: %w", err)
	}
	if affinity != nil {
		podSpec["affinity"] = affinity
	}

	return podSpec, nil
}

// replicaSpecOptions holds parameters for buildRoleReplicaSpec.
type replicaSpecOptions struct {
	ContainerName  string
	Task           *task.Task
	Resources      task.Resources
	Envs           map[string]task.EnvValue
	Replicas       int64
	RestartPolicy  string
	IncludeVolumes bool
	Run            string
}

// buildRoleReplicaSpec creates a replica spec with custom container name, resources, envs, and replicas.
// This is a shared helper used by all providers (PyTorch, TF, MPI) to eliminate duplication.
func buildRoleReplicaSpec(opts replicaSpecOptions) (map[string]interface{}, error) {
	container := map[string]interface{}{
		"name":  opts.ContainerName,
		"image": opts.Task.Image,
	}

	if res := buildResources(opts.Resources); res != nil {
		container["resources"] = res
	}

	if opts.Run != "" {
		cmd, args := buildCommandArgs(opts.Task, opts.Run)
		container["command"] = cmd
		container["args"] = args
	}

	if envVars := buildEnvVars(opts.Task, opts.Envs); len(envVars) > 0 {
		container["env"] = envVars
	}

	// Volume mounts (from storages only — sync volumes live in storages now)
	mounts := []interface{}{}
	if opts.IncludeVolumes {
		_, storageMounts := BuildVolumes(opts.Task)
		mounts = append(mounts, storageMounts...)
	}
	if len(mounts) > 0 {
		container["volumeMounts"] = mounts
	}

	if opts.Task.ImagePullPolicy != "" {
		container["imagePullPolicy"] = opts.Task.ImagePullPolicy
	}
	if opts.Task.WorkingDir != "" {
		container["workingDir"] = opts.Task.WorkingDir
	}

	podSpec, err := buildPodSpec(opts.Task, container, opts.IncludeVolumes)
	if err != nil {
		return nil, fmt.Errorf("failed to build replica spec for %s: %w", opts.ContainerName, err)
	}

	template := map[string]interface{}{
		"spec": podSpec,
	}

	return map[string]interface{}{
		"replicas":      opts.Replicas,
		"restartPolicy": opts.RestartPolicy,
		"template":      template,
	}, nil
}
