package task

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/kubeflow/arena/pkg/constants"

	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/util/validation"
)

// versionRegex validates MAJOR.MINOR.PATCH format for schema versions.
var versionRegex = regexp.MustCompile(`^\d+\.\d+\.\d+$`)

// defaultSchemaVersion is the schema version used when none is specified.
const defaultSchemaVersion = "0.1.0"

// knownSchemaVersions maps MAJOR.MINOR to whether the CLI can parse that schema.
var knownSchemaVersions = map[string]bool{
	"0.1": true,
}

// validFrameworks lists all supported training frameworks.
var validFrameworks = map[string]bool{
	constants.FrameworkPyTorch: true, constants.FrameworkTensorFlow: true,
	constants.FrameworkMPI: true, constants.FrameworkHorovod: true,
	constants.FrameworkDeepSpeed: true, constants.FrameworkRay: true,
}

// Validation lookup maps for enum-like fields. Declared at package level
// to avoid repeated allocations on every Validate call.
var (
	cleanPodPolicies    = map[string]bool{constants.CleanPodPolicyNone: true, constants.CleanPodPolicyRunning: true, constants.CleanPodPolicyAll: true}
	restartPolicies     = map[string]bool{constants.RestartPolicyAlways: true, constants.RestartPolicyOnFailure: true, constants.RestartPolicyNever: true}
	imagePullPolicies   = map[string]bool{"Always": true, "IfNotPresent": true, "Never": true}
	successPolicies     = map[string]bool{constants.SuccessPolicyChiefWorkerAlias: true, constants.SuccessPolicyAllWorkers: true}
	mpiImplementations  = map[string]bool{"OpenMPI": true, "Intel": true, "MPICH": true}
	launcherPolicies    = map[string]bool{"AtStartup": true, "WaitForWorkersReady": true}
	affinityPolicies    = map[string]bool{"spread": true, "binpack": true, "none": true}
	affinityConstraints = map[string]bool{"preferred": true, "required": true}
)

// Task is the top-level configuration schema for an Arena v2 training job.
// YAML field names use snake_case per the arena-v2 design spec.
type Task struct {
	// Schema version for forward compatibility.
	Version string `yaml:"version,omitempty"`

	// Identity
	Name        string            `yaml:"name"`
	Namespace   string            `yaml:"namespace,omitempty"`
	Description string            `yaml:"description,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`

	// Task
	Image      string              `yaml:"image"`
	Run        string              `yaml:"run"`
	Shell      string              `yaml:"shell,omitempty"`
	WorkingDir string              `yaml:"working_dir,omitempty"`
	Framework  Framework           `yaml:"framework"`
	Envs       map[string]EnvValue `yaml:"envs,omitempty"`

	// Scale
	Worker *Worker `yaml:"worker,omitempty"`

	// Optional roles — provider-specific
	Master    *RoleConfig `yaml:"master,omitempty"`    // PyTorch
	Chief     *RoleConfig `yaml:"chief,omitempty"`     // TFJob
	PS        *RoleConfig `yaml:"ps,omitempty"`        // TFJob
	Evaluator *RoleConfig `yaml:"evaluator,omitempty"` // TFJob
	Launcher  *RoleConfig `yaml:"launcher,omitempty"`  // MPIJob

	// Sync & Init
	Sync []SyncEntry     `yaml:"sync,omitempty"`
	Init []InitContainer `yaml:"init,omitempty"`

	// Storage
	Storages []Storage `yaml:"storages,omitempty"`

	// Scheduling
	Scheduling Scheduling `yaml:"scheduling,omitempty"`

	// Lifecycle
	Lifecycle Lifecycle `yaml:"lifecycle,omitempty"`

	// Runtime
	ImagePullPolicy  string   `yaml:"image_pull_policy,omitempty"`
	ImagePullSecrets []string `yaml:"image_pull_secrets,omitempty"`
	ServiceAccount   string   `yaml:"service_account,omitempty"`
	Restart          string   `yaml:"restart,omitempty"`
	HostNetwork      bool     `yaml:"host_network,omitempty"`
	HostIPC          bool     `yaml:"host_ipc,omitempty"`
	HostPID          bool     `yaml:"host_pid,omitempty"`

	// Logging
	Logging Logging `yaml:"logging,omitempty"`
}

// Framework selects the training provider and carries its typed options.
type Framework struct {
	Name    string          `yaml:"name"`
	Options FrameworkConfig `yaml:"options,omitempty"`
}

// FrameworkConfig holds provider-specific options. Only the fields matching the
// selected provider are populated; the rest remain nil/zero.
type FrameworkConfig struct {
	// PyTorch
	NprocPerNode string `yaml:"nproc_per_node,omitempty"`

	// MPI
	SlotsPerWorker         int    `yaml:"slots_per_worker,omitempty"`
	MountsOnLauncher       bool   `yaml:"mounts_on_launcher,omitempty"`
	RunLauncherAsWorker    bool   `yaml:"run_launcher_as_worker,omitempty"`
	GPUTopology            bool   `yaml:"gpu_topology,omitempty"`
	MPIImplementation      string `yaml:"mpi_implementation,omitempty"`
	LauncherCreationPolicy string `yaml:"launcher_creation_policy,omitempty"`
	SSHAuthMountPath       string `yaml:"ssh_auth_mount_path,omitempty"`
}

// Worker defines the GPU-using replica configuration.
type Worker struct {
	Replicas  int                 `yaml:"replicas"`
	Resources Resources           `yaml:"resources,omitempty"`
	Envs      map[string]EnvValue `yaml:"envs,omitempty"`
	Run       string              `yaml:"run,omitempty"`
}

// RoleConfig defines per-role resource and environment overrides.
// Used by optional roles (master, chief, ps, evaluator, launcher).
type RoleConfig struct {
	// Replicas is only meaningful for unconstrained roles (PS).
	// Constrained roles (master, chief, launcher, evaluator) force replicas=1.
	Replicas  int                 `yaml:"replicas,omitempty"`
	Resources Resources           `yaml:"resources,omitempty"`
	Envs      map[string]EnvValue `yaml:"envs,omitempty"`
	Run       string              `yaml:"run,omitempty"`
}

// Resources maps K8s resource names to quantities. Values are applied to
// both requests and limits (Guaranteed QoS).
type Resources map[string]string

// GangConfig controls gang scheduling behavior.
type GangConfig struct {
	Enabled bool `yaml:"enabled,omitempty"`
}

// Scheduling controls pod placement.
type Scheduling struct {
	Priority          int               `yaml:"priority,omitempty"`
	PriorityClassName string            `yaml:"priority_class_name,omitempty"`
	Gang              GangConfig        `yaml:"gang,omitempty"`
	SchedulerName     string            `yaml:"scheduler_name,omitempty"`
	Queue             string            `yaml:"queue,omitempty"`
	NodeSelector      map[string]string `yaml:"node_selector,omitempty"`
	Tolerations       []Toleration      `yaml:"tolerations,omitempty"`
	Affinity          *Affinity         `yaml:"affinity,omitempty"`
}

// Toleration mirrors the K8s toleration spec.
type Toleration struct {
	Key               string `yaml:"key,omitempty"`
	Operator          string `yaml:"operator,omitempty"`
	Value             string `yaml:"value,omitempty"`
	Effect            string `yaml:"effect,omitempty"`
	TolerationSeconds *int64 `yaml:"toleration_seconds,omitempty"`
}

// Affinity uses orthogonal policy x constraint x target dimensions.
type Affinity struct {
	Policy     string         `yaml:"policy,omitempty"`
	Constraint string         `yaml:"constraint,omitempty"`
	Target     string         `yaml:"target,omitempty"`
	Rules      []AffinityRule `yaml:"rules,omitempty"`
}

// AffinityRule maps directly to K8s affinity rule fields.
type AffinityRule struct {
	TopologyKey       string            `yaml:"topology_key,omitempty"`
	Weight            int               `yaml:"weight,omitempty"`
	MatchExpressions  []MatchExpression `yaml:"match_expressions,omitempty"`
	MatchFields       []MatchExpression `yaml:"match_fields,omitempty"`
	MatchLabels       map[string]string `yaml:"match_labels,omitempty"`
	Namespaces        []string          `yaml:"namespaces,omitempty"`
	NamespaceSelector *LabelSelector    `yaml:"namespace_selector,omitempty"`
}

// MatchExpression mirrors the K8s LabelSelectorRequirement.
type MatchExpression struct {
	Key      string   `yaml:"key"`
	Operator string   `yaml:"operator"`
	Values   []string `yaml:"values,omitempty"`
}

// LabelSelector mirrors the K8s LabelSelector.
type LabelSelector struct {
	MatchLabels      map[string]string `yaml:"match_labels,omitempty"`
	MatchExpressions []MatchExpression `yaml:"match_expressions,omitempty"`
}

// Lifecycle wraps training-operator RunPolicy (job-level policies).
type Lifecycle struct {
	CleanPodPolicy   string `yaml:"clean_pod_policy,omitempty"`
	ActiveDeadline   string `yaml:"active_deadline,omitempty"`
	TTLAfterFinished string `yaml:"ttl_after_finished,omitempty"`
	BackoffLimit     *int   `yaml:"backoff_limit,omitempty"`
	SuccessPolicy    string `yaml:"success_policy,omitempty"`
	Suspend          *bool  `yaml:"suspend,omitempty"`
	ManagedBy        string `yaml:"managed_by,omitempty"`
}

// Logging configures sidecar logging resources.
type Logging struct {
	TensorBoard *TensorBoardConfig `yaml:"tensorboard,omitempty"`
}

// TensorBoardConfig enables a TensorBoard Deployment+Service.
type TensorBoardConfig struct {
	Enabled bool    `yaml:"enabled"`
	LogDir  string  `yaml:"logdir,omitempty"`
	Image   string  `yaml:"image,omitempty"`
	Mounts  []Mount `yaml:"mounts,omitempty"`
}

// Storage declares a volume mount. Exactly one of PVC, SHM, Tmp, HostPath,
// ConfigMap, or Secret should be non-empty.
type Storage struct {
	Name      string `yaml:"name"`
	MountPath string `yaml:"mount_path,omitempty"`
	SubPath   string `yaml:"sub_path,omitempty"`
	PVC       string `yaml:"pvc,omitempty"`
	SHM       string `yaml:"shm,omitempty"`
	Tmp       string `yaml:"tmp,omitempty"`
	HostPath  string `yaml:"hostpath,omitempty"`
	ConfigMap string `yaml:"configmap,omitempty"`
	Secret    string `yaml:"secret,omitempty"`
	Key       string `yaml:"key,omitempty"`
}

// Validate checks that the Storage entry has a name, mount path, and exactly
// one storage type declared.
func (s *Storage) Validate() error {
	if s.Name == "" {
		return errors.New("storage name must not be empty")
	}
	if s.MountPath == "" {
		return fmt.Errorf("storage %q: mountPath must not be empty", s.Name)
	}

	types := []struct {
		name string
		val  string
	}{
		{name: "pvc", val: s.PVC},
		{name: "shm", val: s.SHM},
		{name: "tmp", val: s.Tmp},
		{name: "hostpath", val: s.HostPath},
		{name: "configmap", val: s.ConfigMap},
		{name: "secret", val: s.Secret},
	}

	set := make([]string, 0, 6)
	for _, t := range types {
		if t.val != "" {
			set = append(set, t.name)
		}
	}

	switch len(set) {
	case 0:
		return fmt.Errorf("storage %q: must specify exactly one of pvc, shm, tmp, hostpath, configmap, or secret", s.Name)
	case 1:
		// OK
	default:
		return fmt.Errorf("storage %q: cannot specify multiple storage types (%s)", s.Name, strings.Join(set, ", "))
	}

	hasKey := s.Key != ""
	hasConfigMap := s.ConfigMap != ""
	hasSecret := s.Secret != ""
	if hasKey && !hasConfigMap && !hasSecret {
		return fmt.Errorf("storage %q: key can only be used with configmap or secret", s.Name)
	}

	return nil
}

// Mount is used by SyncEntry to override storage mount points.
type Mount struct {
	Name      string `yaml:"name,omitempty"`
	MountPath string `yaml:"mount_path,omitempty"`
	SubPath   string `yaml:"sub_path,omitempty"`
}

// SyncEntry declares a code/data injection source.
type SyncEntry struct {
	Git       string  `yaml:"git,omitempty"`
	Rsync     string  `yaml:"rsync,omitempty"`
	HDFS      string  `yaml:"hdfs,omitempty"`
	Branch    string  `yaml:"branch,omitempty"`
	LocalPath string  `yaml:"local_path,omitempty"`
	Image     string  `yaml:"image,omitempty"`
	Mounts    []Mount `yaml:"mounts,omitempty"`
}

// InitContainer defines a generic init container with run+shell semantics.
type InitContainer struct {
	Name   string  `yaml:"name"`
	Image  string  `yaml:"image"`
	Run    string  `yaml:"run"`
	Shell  string  `yaml:"shell,omitempty"`
	Mounts []Mount `yaml:"mounts,omitempty"`
}

// EnvValue holds an environment variable value: a plain string, a secret
// reference, or a configmap reference.
type EnvValue struct {
	Value     string
	Secret    *EnvFrom
	ConfigMap *EnvFrom
}

// EnvFrom references a key in a K8s Secret or ConfigMap.
type EnvFrom struct {
	Name string
	Key  string
}

// UnmarshalYAML handles the three envs value forms:
//
//	plain:    NCCL_DEBUG: "INFO"
//	secret:   HF_TOKEN: {secret: my-hf-creds, key: token}
//	configmap: DB_HOST: {configmap: db-config, key: host}
func (e *EnvValue) UnmarshalYAML(node *yaml.Node) error {
	*e = EnvValue{}
	if node.Kind == yaml.ScalarNode {
		e.Value = node.Value
		return nil
	}
	if node.Kind == yaml.MappingNode {
		var raw map[string]string
		if err := node.Decode(&raw); err != nil {
			return fmt.Errorf("envs: failed to decode mapping: %w", err)
		}

		_, hasSecret := raw["secret"]
		_, hasConfigMap := raw["configmap"]

		if hasSecret && hasConfigMap {
			return errors.New("envs: mapping must specify exactly one of 'secret' or 'configmap', not both")
		}
		if !hasSecret && !hasConfigMap {
			return errors.New("envs: mapping must contain 'secret' or 'configmap' key")
		}

		if hasSecret {
			name := raw["secret"]
			key := raw["key"]
			if name == "" {
				return errors.New("envs: secret name must not be empty")
			}
			if key == "" {
				return fmt.Errorf("envs: key must not be empty for secret reference %q", name)
			}
			e.Secret = &EnvFrom{Name: name, Key: key}
			return nil
		}

		name := raw["configmap"]
		key := raw["key"]
		if name == "" {
			return errors.New("envs: configmap name must not be empty")
		}
		if key == "" {
			return fmt.Errorf("envs: key must not be empty for configmap reference %q", name)
		}
		e.ConfigMap = &EnvFrom{Name: name, Key: key}
		return nil
	}
	return fmt.Errorf("envs: unsupported value type (kind=%d)", node.Kind)
}

// MarshalYAML produces the canonical form for serialization.
func (e EnvValue) MarshalYAML() (interface{}, error) {
	if e.Secret != nil && e.ConfigMap != nil {
		return nil, errors.New("envs: cannot set both secret and configmap on the same EnvValue")
	}
	if e.Secret != nil {
		return map[string]string{"secret": e.Secret.Name, "key": e.Secret.Key}, nil
	}
	if e.ConfigMap != nil {
		return map[string]string{"configmap": e.ConfigMap.Name, "key": e.ConfigMap.Key}, nil
	}
	return e.Value, nil
}

// validateVersion checks the Task version field.
// Empty version is accepted as-is (SetDefaults will fill it in).
// Non-empty version must match MAJOR.MINOR.PATCH format.
// Unknown MAJOR.MINOR returns an error prompting CLI upgrade.
func validateVersion(t *Task) error {
	if t.Version == "" {
		return nil
	}
	if !versionRegex.MatchString(t.Version) {
		return fmt.Errorf("invalid version: %q (must be MAJOR.MINOR.PATCH, e.g. 0.1.0)", t.Version)
	}
	parts := strings.Split(t.Version, ".")
	majorMinor := parts[0] + "." + parts[1]
	if !knownSchemaVersions[majorMinor] {
		return fmt.Errorf("version %q is newer than supported (current: 0.1.x)", t.Version)
	}
	return nil
}

// ParseDuration parses a duration string with optional day suffix.
// Supports standard Go duration formats ("2h", "30m", "10s") plus day suffix "d"
// (e.g., "7d", "1.5d") which is useful for Kubernetes job deadlines.
func ParseDuration(s string) (time.Duration, error) {
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
		return 0, fmt.Errorf("invalid duration format: %q (expected a number followed by 'd', e.g. '7d')", s)
	}

	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("invalid duration %q: %w", s, err)
	}
	return d, nil
}

// SetDefaults fills in zero-value fields with their default values.
// Call this before Validate to ensure the Task is fully populated.
func (t *Task) SetDefaults() {
	if t.Version == "" {
		t.Version = defaultSchemaVersion
	}
}

// Validate checks that the Task has all required fields and valid values.
func Validate(t *Task) error {
	var errs []error
	errs = append(errs, validateIdentity(t))
	errs = append(errs, validateFramework(t))
	errs = append(errs, validateLifecycle(t))
	errs = append(errs, validateScheduling(t))
	errs = append(errs, validateSync(t))
	errs = append(errs, validateRoles(t))
	errs = append(errs, validateStorages(t))
	return errors.Join(errs...)
}

// validateIdentity validates version, name, image, and run.
func validateIdentity(t *Task) error {
	if err := validateVersion(t); err != nil {
		return err
	}
	if t.Name == "" {
		return errors.New("name is required")
	}
	if errs := validation.IsDNS1123Label(t.Name); len(errs) > 0 {
		return fmt.Errorf("invalid name %q: %s", t.Name, strings.Join(errs, ", "))
	}
	if t.Image == "" {
		return errors.New("image is required")
	}
	if t.Run == "" {
		return errors.New("run is required")
	}
	if t.Namespace != "" {
		if errs := validation.IsDNS1123Label(t.Namespace); len(errs) > 0 {
			return fmt.Errorf("invalid namespace %q: %s", t.Namespace, strings.Join(errs, ", "))
		}
	}
	return nil
}

// validateFramework validates framework name, worker requirements, and framework-specific options.
func validateFramework(t *Task) error {
	if !validFrameworks[t.Framework.Name] {
		return fmt.Errorf("unsupported framework: %q (must be pytorch, tensorflow, mpi, horovod, deepspeed, or ray)", t.Framework.Name)
	}

	if t.Worker == nil {
		if t.Framework.Name != constants.FrameworkPyTorch {
			return fmt.Errorf("worker is required for %q framework", t.Framework.Name)
		}
		if t.Master == nil {
			return errors.New("pytorch requires worker or master (at least one must be specified)")
		}
	} else if t.Worker.Replicas < 1 {
		return fmt.Errorf("worker.replicas must be > 0, got %d", t.Worker.Replicas)
	}

	if t.Framework.Name == constants.FrameworkPyTorch && t.Framework.Options.NprocPerNode != "" {
		v := t.Framework.Options.NprocPerNode
		switch v {
		case "auto", "gpu", "cpu":
			// valid
		default:
			if n, err := strconv.Atoi(v); err != nil || n < 1 {
				return fmt.Errorf("pytorch nproc_per_node must be 'auto', 'gpu', 'cpu', or a positive integer, got %q", v)
			}
		}
	}

	if t.Lifecycle.SuccessPolicy != "" && t.Framework.Name != constants.FrameworkTensorFlow {
		return errors.New("success_policy is only valid for tensorflow framework")
	}
	if t.Lifecycle.SuccessPolicy != "" {
		if !successPolicies[t.Lifecycle.SuccessPolicy] {
			return fmt.Errorf("invalid success_policy: %q (must be ChiefWorker or AllWorkers; ChiefWorker is an alias for the default \"\")", t.Lifecycle.SuccessPolicy)
		}
	}

	if t.Framework.Name == constants.FrameworkMPI {
		if t.Framework.Options.MPIImplementation != "" {
			if !mpiImplementations[t.Framework.Options.MPIImplementation] {
				return fmt.Errorf("invalid mpi_implementation: %q (must be OpenMPI, Intel, or MPICH)", t.Framework.Options.MPIImplementation)
			}
		}
		if t.Framework.Options.LauncherCreationPolicy != "" {
			if !launcherPolicies[t.Framework.Options.LauncherCreationPolicy] {
				return fmt.Errorf("invalid launcher_creation_policy: %q (must be AtStartup or WaitForWorkersReady)", t.Framework.Options.LauncherCreationPolicy)
			}
		}
	}
	return nil
}

// validateLifecycle validates activeDeadline, ttlAfterFinished, restartPolicy, cleanPodPolicy, and imagePullPolicy.
func validateLifecycle(t *Task) error {
	if t.Lifecycle.CleanPodPolicy != "" {
		if !cleanPodPolicies[t.Lifecycle.CleanPodPolicy] {
			return fmt.Errorf("invalid clean_pod_policy: %q (must be None, Running, or All)", t.Lifecycle.CleanPodPolicy)
		}
	}

	if t.Lifecycle.ActiveDeadline != "" {
		if _, err := ParseDuration(t.Lifecycle.ActiveDeadline); err != nil {
			return fmt.Errorf("invalid active_deadline: %q (must be a valid duration like 30s, 5m, 1h, or 7d)", t.Lifecycle.ActiveDeadline)
		}
	}

	if t.Lifecycle.TTLAfterFinished != "" {
		if _, err := ParseDuration(t.Lifecycle.TTLAfterFinished); err != nil {
			return fmt.Errorf("invalid ttl_after_finished: %q (must be a valid duration like 30s, 5m, 1h, or 7d)", t.Lifecycle.TTLAfterFinished)
		}
	}

	if t.Restart != "" {
		if !restartPolicies[t.Restart] {
			return fmt.Errorf("invalid restart: %q (must be Always, OnFailure, or Never)", t.Restart)
		}
	}

	if t.ImagePullPolicy != "" {
		if !imagePullPolicies[t.ImagePullPolicy] {
			return fmt.Errorf("invalid image_pull_policy: %q (must be Always, IfNotPresent, or Never)", t.ImagePullPolicy)
		}
	}
	return nil
}

// validateScheduling validates scheduler, queue, priorityClassName, gang, affinity, tolerations, and nodeSelector.
func validateScheduling(t *Task) error {
	if a := t.Scheduling.Affinity; a != nil {
		// rules requires target
		if len(a.Rules) > 0 && a.Target == "" {
			return errors.New("affinity.target is required when affinity.rules is specified")
		}
		// target valid values
		if a.Target != "" {
			switch a.Target {
			case "pod", "node":
				// valid
			default:
				return fmt.Errorf("affinity.target must be 'pod' or 'node', got %q", a.Target)
			}
		}
		// policy validation
		if a.Policy != "" && a.Policy != "none" {
			if !affinityPolicies[a.Policy] {
				return fmt.Errorf("affinity.policy must be 'spread', 'binpack', or 'none', got %q", a.Policy)
			}
			if len(a.Rules) == 0 {
				return fmt.Errorf("affinity.policy %q requires at least one rule", a.Policy)
			}
		}
		// constraint validation
		if a.Constraint != "" {
			if !affinityConstraints[a.Constraint] {
				return fmt.Errorf("affinity.constraint must be 'preferred' or 'required', got %q", a.Constraint)
			}
		}
		// weight validation: preferred mode requires weight 1-100 (skip for 'none' policy)
		if a.Policy != "none" {
			constraint := a.Constraint
			if constraint == "" {
				constraint = "preferred"
			}
			if constraint == "preferred" {
				for i, rule := range a.Rules {
					if rule.Weight < 1 || rule.Weight > 100 {
						return fmt.Errorf("affinity rule[%d]: weight must be 1-100 for preferred scheduling, got %d", i, rule.Weight)
					}
				}
			}
		}
	}
	// toleration validation
	for i, tol := range t.Scheduling.Tolerations {
		if tol.Operator != "" && tol.Operator != "Equal" && tol.Operator != "Exists" {
			return fmt.Errorf("toleration[%d]: invalid operator %q (must be Equal or Exists)", i, tol.Operator)
		}
		if tol.Effect != "" && tol.Effect != "NoSchedule" && tol.Effect != "PreferNoSchedule" && tol.Effect != "NoExecute" {
			return fmt.Errorf("toleration[%d]: invalid effect %q (must be NoSchedule, PreferNoSchedule, or NoExecute)", i, tol.Effect)
		}
		if tol.Operator == "Exists" && tol.Value != "" {
			return fmt.Errorf("toleration[%d]: Exists operator must not have a value", i)
		}
	}
	return nil
}

// validateStorages validates storage names, duplicates, and individual storage entries.
func validateStorages(t *Task) error {
	var errs []error
	storageMap := make(map[string]Storage, len(t.Storages))
	for _, st := range t.Storages {
		if st.Name == "" {
			errs = append(errs, errors.New("storages: storage name must not be empty"))
			continue
		}
		if _, exists := storageMap[st.Name]; exists {
			errs = append(errs, fmt.Errorf("storages: duplicate storage name %q", st.Name))
			continue
		}
		storageMap[st.Name] = st
	}

	for i := range t.Storages {
		if err := (&t.Storages[i]).Validate(); err != nil {
			errs = append(errs, fmt.Errorf("storages: %w", err))
		}
	}
	return errors.Join(errs...)
}

// validateSync validates sync entries and TensorBoard mount references.
func validateSync(t *Task) error {
	// Build storage lookup once (O(M)) before iterating sync entries (O(N)),
	// so the combined validation is O(N+M) rather than O(N*M).
	storageMap := make(map[string]Storage, len(t.Storages))
	for _, st := range t.Storages {
		storageMap[st.Name] = st
	}

	for i, s := range t.Sync {
		count := 0
		if s.Git != "" {
			count++
		}
		if s.Rsync != "" {
			count++
		}
		if s.HDFS != "" {
			count++
		}
		if count == 0 {
			return fmt.Errorf("sync[%d]: must specify git, rsync, or hdfs", i)
		}
		if count > 1 {
			return fmt.Errorf("sync[%d]: can only specify one of git, rsync, or hdfs", i)
		}
		// local_path is required: it is the sync command's target path inside the container
		if s.LocalPath == "" {
			return fmt.Errorf("sync[%d]: local_path is required (git-sync/rsync/hdfs target path)", i)
		}
		// Validate that sync mount names reference existing storages and that
		// each mount resolves to a non-empty mount_path (mount.MountPath takes
		// precedence; falls back to the referenced Storage.MountPath).
		for j, m := range s.Mounts {
			if m.Name != "" {
				st, ok := storageMap[m.Name]
				if !ok {
					return fmt.Errorf("sync[%d].mounts[%d].name %q not found in storages", i, j, m.Name)
				}
				effectiveMountPath := m.MountPath
				if effectiveMountPath == "" {
					effectiveMountPath = st.MountPath
				}
				if effectiveMountPath == "" {
					return fmt.Errorf("sync[%d].mounts[%d].mount_path is required (neither mount nor storage %q defines mount_path)", i, j, m.Name)
				}
			}
		}
	}

	// TensorBoard mounts validation (identical to sync mount validation)
	if t.Logging.TensorBoard != nil {
		for j, m := range t.Logging.TensorBoard.Mounts {
			if m.Name != "" {
				st, ok := storageMap[m.Name]
				if !ok {
					return fmt.Errorf("logging.tensorboard.mounts[%d].name %q not found in storages", j, m.Name)
				}
				effectiveMountPath := m.MountPath
				if effectiveMountPath == "" {
					effectiveMountPath = st.MountPath
				}
				if effectiveMountPath == "" {
					return fmt.Errorf("logging.tensorboard.mounts[%d].mount_path is required (neither mount nor storage %q defines mount_path)", j, m.Name)
				}
			}
		}
	}
	return nil
}

// validateRoles validates master, chief, launcher, evaluator, and PS roles.
func validateRoles(t *Task) error {
	// Framework-specific roles are validated against their respective frameworks.
	if t.Master != nil && t.Framework.Name != constants.FrameworkPyTorch {
		return errors.New("master role is only valid for pytorch framework")
	}
	if t.Chief != nil && t.Framework.Name != constants.FrameworkTensorFlow {
		return errors.New("chief role is only valid for tensorflow framework")
	}
	if t.PS != nil && t.Framework.Name != constants.FrameworkTensorFlow {
		return errors.New("ps role is only valid for tensorflow framework")
	}
	if t.Evaluator != nil && t.Framework.Name != constants.FrameworkTensorFlow {
		return errors.New("evaluator role is only valid for tensorflow framework")
	}
	isMPIFamily := t.Framework.Name == constants.FrameworkMPI ||
		t.Framework.Name == constants.FrameworkHorovod ||
		t.Framework.Name == constants.FrameworkDeepSpeed
	if t.Launcher != nil && !isMPIFamily {
		return errors.New("launcher role is only valid for mpi, horovod, and deepspeed frameworks")
	}

	// Constrained roles (master, chief, launcher, evaluator) are limited to replicas=1.
	constrained := []struct {
		name string
		role *RoleConfig
	}{
		{name: "master", role: t.Master},
		{name: "chief", role: t.Chief},
		{name: "launcher", role: t.Launcher},
		{name: "evaluator", role: t.Evaluator},
	}
	for _, c := range constrained {
		if c.role != nil && c.role.Replicas > 1 {
			return fmt.Errorf("%s role is constrained to replicas=1, got %d", c.name, c.role.Replicas)
		}
	}

	// PS is unconstrained but requires at least 1 replica if present.
	if t.PS != nil && t.PS.Replicas < 1 {
		return fmt.Errorf("ps.replicas must be > 0, got %d", t.PS.Replicas)
	}
	return nil
}

// EffectiveShell returns the shell to use. Bare executable names (e.g. "bash",
// "sh") and absolute paths (e.g. "/bin/bash") are returned as-is. An empty or
// whitespace-only shell defaults to DefaultShell.
func (t *Task) EffectiveShell() string {
	s := strings.TrimSpace(t.Shell)
	if s == "" {
		return constants.DefaultShell
	}
	return s
}
