package task

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/kubeflow/arena/pkg/constants"
	"gopkg.in/yaml.v3"
)

// versionRegex validates MAJOR.MINOR.PATCH format for schema versions.
var versionRegex = regexp.MustCompile(`^\d+\.\d+\.\d+$`)

// DefaultSchemaVersion is the schema version used when none is specified.
const DefaultSchemaVersion = "0.1.0"

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
	successPolicies     = map[string]bool{"ChiefWorker": true, "AllWorkers": true}
	mpiImplementations  = map[string]bool{"OpenMPI": true, "Intel": true, "MPICH": true}
	launcherPolicies     = map[string]bool{"AtStartup": true, "WaitForWorkersReady": true}
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
	Master    *RoleConfig `yaml:"master,omitempty"`      // PyTorch
	Chief     *RoleConfig `yaml:"chief,omitempty"`       // TFJob
	PS        *RoleConfig `yaml:"ps,omitempty"`          // TFJob
	Evaluator *RoleConfig `yaml:"evaluator,omitempty"`   // TFJob
	Launcher  *RoleConfig `yaml:"launcher,omitempty"`    // MPIJob

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
	Key      string `yaml:"key,omitempty"`
	Operator string `yaml:"operator,omitempty"`
	Value    string `yaml:"value,omitempty"`
	Effect   string `yaml:"effect,omitempty"`
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
func (s Storage) Validate() error {
	if s.Name == "" {
		return fmt.Errorf("storage name must not be empty")
	}
	if s.MountPath == "" {
		return fmt.Errorf("storage %q: mountPath must not be empty", s.Name)
	}

	types := []struct {
		name string
		val  string
	}{
		{"pvc", s.PVC},
		{"shm", s.SHM},
		{"tmp", s.Tmp},
		{"hostpath", s.HostPath},
		{"configmap", s.ConfigMap},
		{"secret", s.Secret},
	}

	var set []string
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

	if s.Key != "" && s.ConfigMap == "" && s.Secret == "" {
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
			return fmt.Errorf("envs: mapping must specify exactly one of 'secret' or 'configmap', not both")
		}
		if !hasSecret && !hasConfigMap {
			return fmt.Errorf("envs: mapping must contain 'secret' or 'configmap' key")
		}

		if hasSecret {
			name := raw["secret"]
			key := raw["key"]
			if name == "" {
				return fmt.Errorf("envs: secret name must not be empty")
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
			return fmt.Errorf("envs: configmap name must not be empty")
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
		return fmt.Errorf("version %s is newer than supported (current: 0.1.x). Upgrade arena CLI or use version: 0.1.0", t.Version)
	}
	return nil
}

// SetDefaults fills in zero-value fields with their default values.
// Call this before Validate to ensure the Task is fully populated.
func (t *Task) SetDefaults() {
	if t.Version == "" {
		t.Version = DefaultSchemaVersion
	}
}

// Validate checks that the Task has all required fields and valid values.
func Validate(t *Task) error {
	if err := validateVersion(t); err != nil {
		return err
	}
	if t.Name == "" {
		return fmt.Errorf("name is required")
	}
	if t.Image == "" {
		return fmt.Errorf("image is required")
	}
	if t.Run == "" {
		return fmt.Errorf("run is required")
	}

	if !validFrameworks[t.Framework.Name] {
		return fmt.Errorf("unsupported framework: %s (must be pytorch, tensorflow, mpi, horovod, deepspeed, or ray)", t.Framework.Name)
	}

	if t.Worker == nil {
		if t.Framework.Name != constants.FrameworkPyTorch {
			return fmt.Errorf("worker is required for %s framework", t.Framework.Name)
		}
		if t.Master == nil {
			return fmt.Errorf("pytorch requires worker or master (at least one must be specified)")
		}
	} else {
		if t.Worker.Replicas < 1 {
			return fmt.Errorf("worker.replicas must be > 0, got %d", t.Worker.Replicas)
		}
	}

	if t.Lifecycle.CleanPodPolicy != "" {
		if !cleanPodPolicies[t.Lifecycle.CleanPodPolicy] {
			return fmt.Errorf("invalid clean_pod_policy: %q (must be None, Running, or All)", t.Lifecycle.CleanPodPolicy)
		}
	}

	if t.Restart != "" {
		if !restartPolicies[t.Restart] {
			return fmt.Errorf("invalid restart: %q (must be Always, OnFailure, or Never)", t.Restart)
		}
	}

	if t.ImagePullPolicy != "" {
		if !imagePullPolicies[t.ImagePullPolicy] {
			return fmt.Errorf("invalid image_pull_policy: %q", t.ImagePullPolicy)
		}
	}

	if t.Framework.Name == constants.FrameworkPyTorch && t.Framework.Options.NprocPerNode != "" {
		v := t.Framework.Options.NprocPerNode
		if v != "auto" && v != "gpu" && v != "cpu" {
			n, err := strconv.Atoi(v)
			if err != nil || n < 1 {
				return fmt.Errorf("pytorch nproc_per_node must be 'auto', 'gpu', 'cpu', or a positive integer, got %q", v)
			}
		}
	}

	if t.Lifecycle.SuccessPolicy != "" && t.Framework.Name != constants.FrameworkTensorFlow {
		return fmt.Errorf("success_policy is only valid for tensorflow framework")
	}
	if t.Lifecycle.SuccessPolicy != "" {
		if !successPolicies[t.Lifecycle.SuccessPolicy] {
			return fmt.Errorf("invalid success_policy: %q (must be ChiefWorker or AllWorkers)", t.Lifecycle.SuccessPolicy)
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

	// Affinity validation
	if a := t.Scheduling.Affinity; a != nil {
		// rules requires target
		if len(a.Rules) > 0 && a.Target == "" {
			return fmt.Errorf("affinity.target is required when affinity.rules is specified")
		}
		// target valid values
		if a.Target != "" && a.Target != "pod" && a.Target != "node" {
			return fmt.Errorf("affinity.target must be 'pod' or 'node', got %q", a.Target)
		}
		// policy validation
		if a.Policy != "" && a.Policy != "none" {
			// policy + node target requires rules (no default node labels to match)
			if a.Target == "node" && len(a.Rules) == 0 {
				return fmt.Errorf("affinity.policy with target: node requires rules (no default node labels)")
			}
			if !affinityPolicies[a.Policy] {
				return fmt.Errorf("affinity.policy must be 'spread', 'binpack', or 'none', got %q", a.Policy)
			}
		}
		// constraint validation
		if a.Constraint != "" {
			if !affinityConstraints[a.Constraint] {
				return fmt.Errorf("affinity.constraint must be 'preferred' or 'required', got %q", a.Constraint)
			}
		}
	}

	// Sync validation
	// Build storage lookup once (O(M)) before iterating sync entries (O(N)),
	// so the combined validation is O(N+M) rather than O(N*M).
	storageMap := make(map[string]Storage, len(t.Storages))
	for _, st := range t.Storages {
		if st.Name == "" {
			return fmt.Errorf("storages: storage name must not be empty")
		}
		if _, exists := storageMap[st.Name]; exists {
			return fmt.Errorf("storages: duplicate storage name %q", st.Name)
		}
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

	// Role validation: framework-specific roles are only valid for their respective frameworks.
	// This prevents users from specifying invalid role combinations (e.g., chief with pytorch).
	// Each role maps to a specific Operator CRD replica type, and mismatched roles would
	// produce CRDs that the Operator cannot reconcile.
	if t.Master != nil && t.Framework.Name != constants.FrameworkPyTorch {
		return fmt.Errorf("master role is only valid for pytorch framework")
	}
	if t.Chief != nil && t.Framework.Name != constants.FrameworkTensorFlow {
		return fmt.Errorf("chief role is only valid for tensorflow framework")
	}
	if t.PS != nil && t.Framework.Name != constants.FrameworkTensorFlow {
		return fmt.Errorf("ps role is only valid for tensorflow framework")
	}
	if t.Evaluator != nil && t.Framework.Name != constants.FrameworkTensorFlow {
		return fmt.Errorf("evaluator role is only valid for tensorflow framework")
	}
	// Launcher is shared across MPI-family frameworks (mpi, horovod, deepspeed),
	// all of which use an MPIJob-style CRD with a launcher/worker split.
	if t.Launcher != nil && t.Framework.Name != constants.FrameworkMPI && t.Framework.Name != constants.FrameworkHorovod && t.Framework.Name != constants.FrameworkDeepSpeed {
		return fmt.Errorf("launcher role is only valid for mpi, horovod, and deepspeed frameworks")
	}

	// Constrained roles have a fixed replica count of 1 in the Operator CRD.
	// Replicas=0 means the field was not set (Go zero value) and is treated as unset.
	// Only master, chief, launcher, and evaluator are constrained; ps is unconstrained.
	constrained := []struct {
		name string
		role *RoleConfig
	}{
		{"master", t.Master},
		{"chief", t.Chief},
		{"launcher", t.Launcher},
		{"evaluator", t.Evaluator},
	}
	for _, c := range constrained {
		if c.role != nil && c.role.Replicas > 1 {
			return fmt.Errorf("%s role is constrained to replicas=1, got %d", c.name, c.role.Replicas)
		}
	}

	// PS (parameter server) is unconstrained: it can have any number of replicas,
	// but must have at least 1 if the section is present (replicas=0 means unset,
	// which would create a zero-replica PS deployment — an invalid configuration).
	if t.PS != nil && t.PS.Replicas < 1 {
		return fmt.Errorf("ps.replicas must be > 0, got %d", t.PS.Replicas)
	}

	// Storage validation
	for _, s := range t.Storages {
		if err := s.Validate(); err != nil {
			return fmt.Errorf("storages: %w", err)
		}
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
