package provider

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/kubeflow/arena/pkg/constants"
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

// toInterfaceSlice converts a []string to []interface{} so that the values can
// be stored in an unstructured.Unstructured object (which requires JSON-compatible types).
func toInterfaceSlice(ss []string) []interface{} {
	out := make([]interface{}, len(ss))
	for i, s := range ss {
		out[i] = s
	}
	return out
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

	var env []interface{}
	for _, name := range keys {
		env = append(env, buildEnvVar(name, merged[name]))
	}
	return env
}

// buildEnvVar creates a single K8s env var map from an EnvValue.
func buildEnvVar(name string, e task.EnvValue) map[string]interface{} {
	envVar := map[string]interface{}{"name": name}
	if e.Secret != nil {
		envVar["valueFrom"] = map[string]interface{}{
			"secretKeyRef": map[string]interface{}{
				"name": e.Secret.Name,
				"key":  e.Secret.Key,
			},
		}
	} else if e.ConfigMap != nil {
		envVar["valueFrom"] = map[string]interface{}{
			"configMapKeyRef": map[string]interface{}{
				"name": e.ConfigMap.Name,
				"key":  e.ConfigMap.Key,
			},
		}
	} else {
		envVar["value"] = e.Value
	}
	return envVar
}

// BuildVolumes creates volumes and volumeMounts from the task's Storages.
// Exported for use by other v2 packages (e.g. lifecycle) and internal callers.
func BuildVolumes(t *task.Task) ([]interface{}, []interface{}) {
	if len(t.Storages) == 0 {
		return nil, nil
	}
	var volumes []interface{}
	var mounts []interface{}
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
		if s.PVC != "" {
			vol["persistentVolumeClaim"] = map[string]interface{}{
				"claimName": s.PVC,
			}
		} else if s.SHM != "" {
			vol["emptyDir"] = map[string]interface{}{
				"medium":    constants.EmptyDirMediumMemory,
				"sizeLimit": s.SHM,
			}
		} else if s.Tmp != "" {
			vol["emptyDir"] = map[string]interface{}{
				"sizeLimit": s.Tmp,
			}
		} else if s.HostPath != "" {
			vol["hostPath"] = map[string]interface{}{
				"path": s.HostPath,
			}
		} else if s.ConfigMap != "" {
			vol["configMap"] = map[string]interface{}{
				"name": s.ConfigMap,
			}
			if s.Key != "" {
				mount["subPath"] = s.Key
			}
		} else if s.Secret != "" {
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

// ResolveMounts builds K8s volumeMount specs from Mount entries by looking up
// the referenced Storage by name. Mount fields override the Storage defaults:
// - mount_path falls back to Storage.MountPath when empty
// - sub_path falls back to Storage.SubPath when empty
// - sub_path falls back to Storage.Key for ConfigMap/Secret storages
// Exported for use by other v2 packages (e.g. lifecycle) and internal callers.
func ResolveMounts(mounts []task.Mount, storages []task.Storage) []map[string]interface{} {
	if len(mounts) == 0 {
		return nil
	}
	storageMap := make(map[string]task.Storage, len(storages))
	for _, s := range storages {
		storageMap[s.Name] = s
	}
	var result []map[string]interface{}
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
			if subPath == "" && (st.ConfigMap != "" || st.Secret != "") && st.Key != "" {
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
// - Storages + mounts: only listed storages get mounts (ResolveMounts overrides)
// Exported for use by other v2 packages (e.g. lifecycle) and internal callers.
func ResolveContainerMounts(mounts []task.Mount, t *task.Task) []interface{} {
	if len(t.Storages) == 0 {
		return nil
	}
	if len(mounts) > 0 {
		resolved := ResolveMounts(mounts, t.Storages)
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
		var tolerations []interface{}
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
		var secrets []interface{}
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
		d, err := parseDuration(t.Lifecycle.ActiveDeadline)
		if err != nil {
			return nil, fmt.Errorf("invalid active_deadline: %w", err)
		}
		if d > 0 {
			policy["activeDeadlineSeconds"] = int64(d.Seconds())
		}
	}
	if t.Lifecycle.TTLAfterFinished != "" {
		d, err := parseDuration(t.Lifecycle.TTLAfterFinished)
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
	if !s.Gang.Enabled && s.Queue == "" && s.PriorityClassName == "" {
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
	if t.Master == nil && t.Worker != nil && t.Framework.Name == constants.FrameworkPyTorch {
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
// - Policy mode: policy×constraint generates podAffinity/podAntiAffinity or nodeAffinity (requires rules for node)
// - Rules mode: custom rules applied to pod or node affinity based on target
// When policy is set but rules is empty, returns nil (extension point for future default rules).
func buildAffinity(a *task.Affinity, jobName string) (map[string]interface{}, error) {
	if a == nil {
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
	} else if a.Policy != "" {
		_ = jobName
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

	if a.Policy == "spread" {
		if constraint == "preferred" {
			terms, err := buildPodAffinityTerms(a.Rules, "preferred")
			if err != nil {
				return err
			}
			affinity["podAntiAffinity"] = map[string]interface{}{
				"preferredDuringSchedulingIgnoredDuringExecution": terms,
			}
		} else {
			terms, err := buildPodAffinityTerms(a.Rules, "required")
			if err != nil {
				return err
			}
			affinity["podAntiAffinity"] = map[string]interface{}{
				"requiredDuringSchedulingIgnoredDuringExecution": terms,
			}
		}
	} else {
		if constraint == "preferred" {
			terms, err := buildPodAffinityTerms(a.Rules, "preferred")
			if err != nil {
				return err
			}
			affinity["podAffinity"] = map[string]interface{}{
				"preferredDuringSchedulingIgnoredDuringExecution": terms,
			}
		} else {
			terms, err := buildPodAffinityTerms(a.Rules, "required")
			if err != nil {
				return err
			}
			affinity["podAffinity"] = map[string]interface{}{
				"requiredDuringSchedulingIgnoredDuringExecution": terms,
			}
		}
	}
	return nil
}

// applyNodeRules applies affinity rules to nodeAffinity based on constraint.
func applyNodeRules(affinity map[string]interface{}, a *task.Affinity) {
	constraint := a.Constraint
	if constraint == "" {
		constraint = "preferred"
	}

	if constraint == "preferred" {
		affinity["nodeAffinity"] = map[string]interface{}{
			"preferredDuringSchedulingIgnoredDuringExecution": buildNodeSelectorTerms(a.Rules),
		}
	} else { // required
		affinity["nodeAffinity"] = map[string]interface{}{
			"requiredDuringSchedulingIgnoredDuringExecution": map[string]interface{}{
				"nodeSelectorTerms": buildNodeSelectorTerms(a.Rules),
			},
		}
	}
}

// buildPodAffinityTerms converts AffinityRules to pod affinity terms.
// In preferred mode, returns WeightedPodAffinityTerm structures (weight at outer level).
// In required mode, returns PodAffinityTerm structures directly (no weight).
// Returns an error if preferred mode rules have weight outside 1-100.
func buildPodAffinityTerms(rules []task.AffinityRule, mode string) ([]interface{}, error) {
	var terms []interface{}
	for _, rule := range rules {
		if mode == "preferred" {
			if rule.Weight < 1 || rule.Weight > 100 {
				return nil, fmt.Errorf("invalid affinity rule weight: %d (must be 1-100 for preferred scheduling)", rule.Weight)
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
				var exprs []interface{}
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
					var exprs []interface{}
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
			terms = append(terms, map[string]interface{}{
				"weight":          rule.Weight,
				"podAffinityTerm": term,
			})
		} else {
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
				var exprs []interface{}
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
					var exprs []interface{}
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
			terms = append(terms, term)
		}
	}
	return terms, nil
}

// buildNodeSelectorTerms converts AffinityRules to node selector terms.
func buildNodeSelectorTerms(rules []task.AffinityRule) []interface{} {
	var terms []interface{}
	for _, rule := range rules {
		term := map[string]interface{}{}
		// Convert MatchExpressions
		if len(rule.MatchExpressions) > 0 {
			var exprs []interface{}
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
			var exprs []interface{}
			for k, v := range rule.MatchLabels {
				exprs = append(exprs, map[string]interface{}{
					"key":      k,
					"operator": constants.AffinityOperatorIn,
					"values":   []string{v},
				})
			}
			if existing, ok := term["matchExpressions"]; ok {
				term["matchExpressions"] = append(existing.([]interface{}), exprs...)
			} else {
				term["matchExpressions"] = exprs
			}
		}
		// Convert MatchFields
		if len(rule.MatchFields) > 0 {
			var fields []interface{}
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
	var containers []map[string]interface{}

	// 1. System-generated init containers from sync (executed first)
	if syncInits := buildSyncInitContainers(t); len(syncInits) > 0 {
		containers = append(containers, syncInits...)
	}

	// 2. User-defined init containers
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

	var containers []map[string]interface{}
	for i, s := range t.Sync {
		containerName := fmt.Sprintf("arena-sync-%d", i)

		// Build volumeMounts from mounts (references storages by name)
		volumeMounts := ResolveContainerMounts(s.Mounts, t)

		if s.Git != "" {
			projectName := extractGitProjectName(s.Git)

			defaultEnvs := map[string]string{
				"GIT_SYNC_REPO":     s.Git,
				"GIT_SYNC_DEST":     projectName,
				"GIT_SYNC_ROOT":     s.LocalPath,
				"GIT_SYNC_ONE_TIME": "true",
			}
			if s.Branch != "" {
				defaultEnvs["GIT_SYNC_BRANCH"] = s.Branch
			}
			envList := buildEnvWithOverrides(defaultEnvs, t.Envs)

			image := s.Image
			if image == "" {
				image = constants.DefaultGitSyncImage
			}

			c := map[string]interface{}{
				"name":            containerName,
				"image":           image,
				"imagePullPolicy": "Always",
				"env":             envList,
			}
			if len(volumeMounts) > 0 {
				c["volumeMounts"] = volumeMounts
			}
			containers = append(containers, c)

		} else if s.Rsync != "" {
			image := s.Image
			if image == "" {
				image = constants.DefaultRsyncImage
			}

			c := map[string]interface{}{
				"name":            containerName,
				"image":           image,
				"imagePullPolicy": "Always",
				"command":         []interface{}{"rsync", "-avP", s.Rsync, s.LocalPath},
			}
			if len(volumeMounts) > 0 {
				c["volumeMounts"] = volumeMounts
			}
			containers = append(containers, c)

		} else if s.HDFS != "" {
			image := s.Image
			if image == "" {
				image = constants.DefaultHDFSImage
			}

			c := map[string]interface{}{
				"name":            containerName,
				"image":           image,
				"imagePullPolicy": "Always",
				"command":         []interface{}{"hdfs", "dfs", "-get", s.HDFS, s.LocalPath},
			}
			if len(volumeMounts) > 0 {
				c["volumeMounts"] = volumeMounts
			}
			containers = append(containers, c)
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
				fmt.Fprintf(os.Stderr,
					"warning: sync[%d] local_path %q does not match any mount path; synced data will be written to container ephemeral storage\n",
					i, s.LocalPath)
			}
		}
	}

	if len(containers) == 0 {
		return nil
	}
	return containers
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
	var refEnvs []map[string]interface{}
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
	var result []map[string]interface{}
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
		return refEnvs[i]["name"].(string) < refEnvs[j]["name"].(string)
	})
	result = append(result, refEnvs...)
	return result
}

// extractGitProjectName extracts project name from git URL.
// Example: "https://github.com/kubeflow/training-operator.git" -> "training-operator"
func extractGitProjectName(gitURL string) string {
	// Remove trailing .git if present
	name := strings.TrimSuffix(gitURL, ".git")
	// Extract last path component
	parts := strings.Split(name, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return "repo"
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

// buildContainer creates a container spec with the run+shell model.
func buildContainer(name, image string, t *task.Task, resources task.Resources, roleEnvs map[string]task.EnvValue, run string) map[string]interface{} {
	container := map[string]interface{}{
		"name":  name,
		"image": image,
	}

	// Resources
	res := buildResources(resources)
	if res != nil {
		container["resources"] = res
	}

	// Command and args (run + shell model)
	if run != "" {
		cmd, args := buildCommandArgs(t, run)
		container["command"] = cmd
		container["args"] = args
	}

	// Environment variables
	envVars := buildEnvVars(t, roleEnvs)
	if len(envVars) > 0 {
		container["env"] = envVars
	}

	// Volume mounts (from storages only — sync volumes live in storages now)
	_, mounts := BuildVolumes(t)
	if len(mounts) > 0 {
		container["volumeMounts"] = mounts
	}

	// Image pull policy
	if t.ImagePullPolicy != "" {
		container["imagePullPolicy"] = t.ImagePullPolicy
	}

	// Working directory
	if t.WorkingDir != "" {
		container["workingDir"] = t.WorkingDir
	}

	return container
}

// buildPodSpec creates a pod spec with container, volumes, scheduling, affinity, and init containers.
func buildPodSpec(t *task.Task, container map[string]interface{}, includeVolumes bool) (map[string]interface{}, error) {
	podSpec := map[string]interface{}{
		"containers": []interface{}{container},
	}

	// Volumes from storages only (sync references storages, no separate sync volumes)
	var volumes []interface{}
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

// buildRoleReplicaSpec creates a replica spec with custom container name, resources, envs, and replicas.
// This is a shared helper used by all providers (PyTorch, TF, MPI) to eliminate duplication.
func buildRoleReplicaSpec(containerName string, t *task.Task, resources task.Resources, envs map[string]task.EnvValue, replicas int64, restartPolicy string, includeVolumes bool, run string) (map[string]interface{}, error) {
	container := map[string]interface{}{
		"name":  containerName,
		"image": t.Image,
	}

	if res := buildResources(resources); res != nil {
		container["resources"] = res
	}

	if run != "" {
		cmd, args := buildCommandArgs(t, run)
		container["command"] = cmd
		container["args"] = args
	}

	if envVars := buildEnvVars(t, envs); len(envVars) > 0 {
		container["env"] = envVars
	}

	// Volume mounts (from storages only — sync volumes live in storages now)
	var mounts []interface{}
	if includeVolumes {
		_, storageMounts := BuildVolumes(t)
		mounts = append(mounts, storageMounts...)
	}
	if len(mounts) > 0 {
		container["volumeMounts"] = mounts
	}

	if t.ImagePullPolicy != "" {
		container["imagePullPolicy"] = t.ImagePullPolicy
	}
	if t.WorkingDir != "" {
		container["workingDir"] = t.WorkingDir
	}

	podSpec, err := buildPodSpec(t, container, includeVolumes)
	if err != nil {
		return nil, fmt.Errorf("failed to build replica spec for %s: %w", containerName, err)
	}

	template := map[string]interface{}{
		"spec": podSpec,
	}

	return map[string]interface{}{
		"replicas":      replicas,
		"restartPolicy": restartPolicy,
		"template":      template,
	}, nil
}

// parseDuration handles durations like "7d", "2h", "30m", "10s".
// Returns an error if the input cannot be parsed.
func parseDuration(s string) (time.Duration, error) {
	if s == "" {
		return 0, nil
	}
	s = strings.TrimSpace(s)

	// Handle day suffix "d"
	if strings.HasSuffix(s, "d") {
		numStr := strings.TrimSuffix(s, "d")
		var days float64
		if _, err := fmt.Sscanf(numStr, "%f", &days); err == nil {
			return time.Duration(days * 24 * float64(time.Hour)), nil
		}
		return 0, fmt.Errorf("invalid duration format: %s", s)
	}

	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("invalid duration %q: %w", s, err)
	}
	return d, nil
}
