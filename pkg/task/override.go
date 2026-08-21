package task

import (
	"fmt"
	"path"
	"strconv"
	"strings"

	"github.com/kubeflow/arena/pkg/constants"
)

// setStrFlag sets a string flag value into the target if the flag exists
// and the value is a non-empty string.
func setStrFlag(flags map[string]interface{}, flag string, target *string) error {
	v, exists := flags[flag]
	if !exists {
		return nil
	}
	s, ok := v.(string)
	if !ok {
		return fmt.Errorf("flag %q: expected string, got %T", flag, v)
	}
	if s != "" {
		*target = s
	}
	return nil
}

// setIntFlag sets an int flag value into the target if the flag exists
// and the value is a positive int.
func setIntFlag(flags map[string]interface{}, flag string, target *int) error {
	v, exists := flags[flag]
	if !exists {
		return nil
	}
	n, ok := v.(int)
	if !ok {
		return fmt.Errorf("flag %q: expected int, got %T", flag, v)
	}
	if n > 0 {
		*target = n
	}
	return nil
}

// setBoolFlag sets a bool flag value into the target if the flag exists.
func setBoolFlag(flags map[string]interface{}, flag string, target *bool) error {
	v, exists := flags[flag]
	if !exists {
		return nil
	}
	b, ok := v.(bool)
	if !ok {
		return fmt.Errorf("flag %q: expected bool, got %T", flag, v)
	}
	*target = b
	return nil
}

// ApplyOverrides applies flag-based overrides to a Task struct.
// Flag keys match the submit command's CLI flag names (v1 names, without the
// leading --). Returns an error when a flag value has an unexpected type.
func ApplyOverrides(t *Task, flags map[string]interface{}) error {
	if err := applyIdentityOverrides(t, flags); err != nil {
		return err
	}
	if err := applyResourceOverrides(t, flags); err != nil {
		return err
	}
	if err := applySchedulingOverrides(t, flags); err != nil {
		return err
	}
	if err := applyLifecycleOverrides(t, flags); err != nil {
		return err
	}
	if err := applyRuntimeOverrides(t, flags); err != nil {
		return err
	}
	if err := applyLoggingOverrides(t, flags); err != nil {
		return err
	}
	if err := applyFrameworkOverrides(t, flags); err != nil {
		return err
	}
	if err := applyRoleResourceOverrides(t, flags); err != nil {
		return err
	}
	if err := applySyncOverrides(t, flags); err != nil {
		return err
	}
	return nil
}

// applyIdentityOverrides handles name, namespace, labels, annotations, and framework name.
func applyIdentityOverrides(t *Task, flags map[string]interface{}) error {
	if err := setStrFlag(flags, "name", &t.Name); err != nil {
		return err
	}
	if err := setStrFlag(flags, "namespace", &t.Namespace); err != nil {
		return err
	}

	if labels, ok := flags["label"].([]string); ok {
		if t.Labels == nil {
			t.Labels = make(map[string]string)
		}
		for _, l := range labels {
			k, v := splitKV(l)
			if k != "" {
				t.Labels[k] = v
			}
		}
	}
	if annotations, ok := flags["annotation"].([]string); ok {
		if t.Annotations == nil {
			t.Annotations = make(map[string]string)
		}
		for _, a := range annotations {
			k, v := splitKV(a)
			if k != "" {
				t.Annotations[k] = v
			}
		}
	}

	if err := setStrFlag(flags, "framework", &t.Framework.Name); err != nil {
		return err
	}
	return nil
}

// applyResourceOverrides handles image, workers, gpus, gpu-type, cpu, memory,
// share-memory, device, envs, data, data-dir, and config-file.
func applyResourceOverrides(t *Task, flags map[string]interface{}) error {
	if err := setStrFlag(flags, "image", &t.Image); err != nil {
		return err
	}

	ensureWorker := func() {
		if t.Worker == nil {
			t.Worker = &Worker{}
		}
	}

	// Scale
	if v, ok := flags["workers"]; ok {
		ensureWorker()
		n, ok := v.(int)
		if !ok {
			return fmt.Errorf("flag %q: expected int, got %T", "workers", v)
		}
		if n > 0 {
			t.Worker.Replicas = n
		}
	}

	// Resources
	if gpus, ok := flags["gpus"].(int); ok && gpus > 0 {
		target := ensureResourceTarget(t)
		target["nvidia.com/gpu"] = strconv.Itoa(gpus)
	}
	if gpuType, ok := flags["gpu-type"].(string); ok && gpuType != "" {
		if t.Scheduling.NodeSelector == nil {
			t.Scheduling.NodeSelector = make(map[string]string)
		}
		t.Scheduling.NodeSelector["nvidia.com/gpu.product"] = gpuType
	}

	if v, ok := flags["cpu"].(string); ok && v != "" {
		target := ensureResourceTarget(t)
		target["cpu"] = v
	}
	if v, ok := flags["memory"].(string); ok && v != "" {
		target := ensureResourceTarget(t)
		target["memory"] = v
	}

	if shm, ok := flags["share-memory"].(string); ok && shm != "" {
		t.Storages = append(t.Storages, Storage{
			Name:      "shm",
			SHM:       shm,
			MountPath: constants.DefaultSHMMountPath,
		})
	}

	if devices, ok := flags["device"].([]string); ok {
		target := ensureResourceTarget(t)
		for _, d := range devices {
			k, v := splitKV(d)
			if k != "" {
				target[k] = v
			}
		}
	}

	// Environment
	if envs, ok := flags["env"].([]string); ok {
		if t.Envs == nil {
			t.Envs = make(map[string]EnvValue)
		}
		for _, e := range envs {
			k, v := splitKV(e)
			if k != "" {
				t.Envs[k] = EnvValue{Value: v}
			}
		}
	}

	// Data
	if dataEntries, ok := flags["data"].([]string); ok {
		for _, d := range dataEntries {
			s, err := parseDataEntry(d)
			if err != nil {
				return err
			}
			t.Storages = append(t.Storages, s)
		}
	}

	// Data-dir (hostpath volumes)
	if dataDirs, ok := flags["data-dir"].([]string); ok {
		for i, d := range dataDirs {
			s, err := parseDataDirEntry(d, i)
			if err != nil {
				return err
			}
			t.Storages = append(t.Storages, s)
		}
	}

	// Config-file (configmap volumes)
	if configFiles, ok := flags["config-file"].([]string); ok {
		for _, c := range configFiles {
			s, err := parseConfigFileEntry(c)
			if err != nil {
				return err
			}
			t.Storages = append(t.Storages, s)
		}
	}
	return nil
}

// validateAbsPath rejects non-absolute paths for v1-compat entries (v1
// validateHostPath behavior).
func validateAbsPath(what, p string) error {
	if !path.IsAbs(p) {
		return fmt.Errorf("%s %q must be an absolute path", what, p)
	}
	return nil
}

// parseDataEntry parses one --data value.
//   - v1 form   "<pvc>:<mount_path>"          mounts the PVC named <pvc>
//   - v2 form   "<name>:<mount_path>:<pvc>"   mounts <pvc> as volume <name>
func parseDataEntry(v string) (Storage, error) {
	parts := strings.Split(v, ":")
	switch len(parts) {
	case 2:
		if err := validateAbsPath("data mount path", parts[1]); err != nil {
			return Storage{}, err
		}
		return Storage{Name: parts[0], MountPath: parts[1], PVC: parts[0]}, nil
	case 3:
		if err := validateAbsPath("data mount path", parts[1]); err != nil {
			return Storage{}, err
		}
		return Storage{Name: parts[0], MountPath: parts[1], PVC: parts[2]}, nil
	default:
		return Storage{}, fmt.Errorf("invalid --data value %q: expected <pvc>:<mount_path> (v1) or <name>:<mount_path>:<pvc> (v2)", v)
	}
}

// parseDataDirEntry parses one --data-dir value.
//   - v1 forms  "<host_path>"                 host path mounted at the same path
//     "<host_path>:<container_path>"
//   - v2 form   "<name>:<mount_path>:<host_path>"
//
// v1 entries use the v1 volume naming convention training-data-<index>.
func parseDataDirEntry(v string, index int) (Storage, error) {
	parts := strings.Split(v, ":")
	name := fmt.Sprintf("training-data-%d", index)
	switch len(parts) {
	case 1:
		if err := validateAbsPath("data-dir host path", parts[0]); err != nil {
			return Storage{}, err
		}
		return Storage{Name: name, MountPath: parts[0], HostPath: parts[0]}, nil
	case 2:
		if err := validateAbsPath("data-dir host path", parts[0]); err != nil {
			return Storage{}, err
		}
		if err := validateAbsPath("data-dir container path", parts[1]); err != nil {
			return Storage{}, err
		}
		return Storage{Name: name, MountPath: parts[1], HostPath: parts[0]}, nil
	case 3:
		if err := validateAbsPath("data-dir container path", parts[1]); err != nil {
			return Storage{}, err
		}
		if err := validateAbsPath("data-dir host path", parts[2]); err != nil {
			return Storage{}, err
		}
		return Storage{Name: parts[0], MountPath: parts[1], HostPath: parts[2]}, nil
	default:
		return Storage{}, fmt.Errorf("invalid --data-dir value %q: expected <host_path>[:<container_path>] (v1) or <name>:<mount_path>:<host_path> (v2)", v)
	}
}

// parseConfigFileEntry parses one --config-file value. Only the v2 configmap
// form is supported; the v1 host-file form errors with guidance because v2
// cannot yet create a configmap from a local file.
func parseConfigFileEntry(v string) (Storage, error) {
	parts := strings.Split(v, ":")
	switch len(parts) {
	case 3:
		if err := validateAbsPath("config-file mount path", parts[1]); err != nil {
			return Storage{}, err
		}
		return Storage{Name: parts[0], MountPath: parts[1], ConfigMap: parts[2]}, nil
	case 1, 2:
		return Storage{}, fmt.Errorf("v1-style --config-file %q (local file mount) is not supported by arena-v2 yet: pre-create a configmap and use --config-file <name>:<mount_path>:<configmap>", v)
	default:
		return Storage{}, fmt.Errorf("invalid --config-file value %q: expected <name>:<mount_path>:<configmap>", v)
	}
}

// ensureResourceTarget returns the Resources map that worker-scoped resource
// overrides apply to. When Worker exists, it returns Worker.Resources. When
// Worker is nil but Master exists (single-node PyTorch), it returns
// Master.Resources. Otherwise it creates a Worker with Replicas=1.
func ensureResourceTarget(t *Task) Resources {
	if t.Worker != nil {
		if t.Worker.Resources == nil {
			t.Worker.Resources = Resources{}
		}
		return t.Worker.Resources
	}
	if t.Master != nil {
		if t.Master.Resources == nil {
			t.Master.Resources = Resources{}
		}
		return t.Master.Resources
	}
	t.Worker = &Worker{Replicas: 1}
	t.Worker.Resources = Resources{}
	return t.Worker.Resources
}

// applySchedulingOverrides handles scheduler, queue, priority, gang, affinity,
// selector, and toleration.
func applySchedulingOverrides(t *Task, flags map[string]interface{}) error {
	if err := setStrFlag(flags, "priority", &t.Scheduling.PriorityClassName); err != nil {
		return err
	}
	if err := setBoolFlag(flags, "gang", &t.Scheduling.Gang.Enabled); err != nil {
		return err
	}
	if err := setStrFlag(flags, "scheduler", &t.Scheduling.SchedulerName); err != nil {
		return err
	}

	// Affinity overrides: policy, constraint, and target
	if err := applyAffinityOverride(flags, "affinity-policy", &t.Scheduling.Affinity); err != nil {
		return err
	}
	if err := applyAffinityOverride(flags, "affinity-constraint", &t.Scheduling.Affinity); err != nil {
		return err
	}
	if err := applyAffinityOverride(flags, "affinity-target", &t.Scheduling.Affinity); err != nil {
		return err
	}

	if selectors, ok := flags["selector"].([]string); ok {
		if t.Scheduling.NodeSelector == nil {
			t.Scheduling.NodeSelector = make(map[string]string)
		}
		for _, s := range selectors {
			k, v := splitKV(s)
			if k != "" {
				t.Scheduling.NodeSelector[k] = v
			}
		}
	}

	if tolerations, ok := flags["toleration"].([]string); ok {
		for _, tol := range tolerations {
			parsed, err := parseTolerationFlag(tol)
			if err != nil {
				return err
			}
			if parsed != nil {
				t.Scheduling.Tolerations = append(t.Scheduling.Tolerations, *parsed)
			}
		}
	}

	// v1 --queue is a bool that suspends the job so an external queue
	// manager (kube-queue) can take over scheduling. The queue *name*
	// (scheduling.queue) is only settable via YAML on job run.
	if v, ok := flags["queue"].(bool); ok && v {
		suspend := true
		t.Lifecycle.Suspend = &suspend
	}
	return nil
}

// applyLifecycleOverrides handles clean-task-policy, running-timeout,
// ttl-after-finished, success-policy, and job-backoff-limit.
func applyLifecycleOverrides(t *Task, flags map[string]interface{}) error {
	if err := setStrFlag(flags, "clean-task-policy", &t.Lifecycle.CleanPodPolicy); err != nil {
		return err
	}
	if err := setStrFlag(flags, "running-timeout", &t.Lifecycle.ActiveDeadline); err != nil {
		return err
	}
	if err := setStrFlag(flags, "ttl-after-finished", &t.Lifecycle.TTLAfterFinished); err != nil {
		return err
	}
	if err := setStrFlag(flags, "success-policy", &t.Lifecycle.SuccessPolicy); err != nil {
		return err
	}

	if v, exists := flags["job-backoff-limit"]; exists {
		n, ok := v.(int)
		if !ok {
			return fmt.Errorf("flag %q: expected int, got %T", "job-backoff-limit", v)
		}
		if n >= 0 {
			t.Lifecycle.BackoffLimit = &n
		}
	}
	return nil
}

// applyRuntimeOverrides handles run, shell, working-dir, image-pull-policy,
// service-account, job-restart-policy, hostNetwork, hostIPC, hostPID, and image-pull-secret.
func applyRuntimeOverrides(t *Task, flags map[string]interface{}) error {
	if err := setStrFlag(flags, "run", &t.Run); err != nil {
		return err
	}
	if err := setStrFlag(flags, "shell", &t.Shell); err != nil {
		return err
	}
	if err := setStrFlag(flags, "working-dir", &t.WorkingDir); err != nil {
		return err
	}
	if err := setStrFlag(flags, "image-pull-policy", &t.ImagePullPolicy); err != nil {
		return err
	}
	if err := setStrFlag(flags, "service-account", &t.ServiceAccount); err != nil {
		return err
	}
	if err := setStrFlag(flags, "job-restart-policy", &t.Restart); err != nil {
		return err
	}
	if err := setBoolFlag(flags, "hostNetwork", &t.HostNetwork); err != nil {
		return err
	}
	if err := setBoolFlag(flags, "hostIPC", &t.HostIPC); err != nil {
		return err
	}
	if err := setBoolFlag(flags, "hostPID", &t.HostPID); err != nil {
		return err
	}

	if secrets, ok := flags["image-pull-secret"].([]string); ok {
		t.ImagePullSecrets = append(t.ImagePullSecrets, secrets...)
	}
	return nil
}

// applyLoggingOverrides handles tensorboard, logdir, and tensorboard-image.
func applyLoggingOverrides(t *Task, flags map[string]interface{}) error {
	if v, ok := flags["tensorboard"].(bool); ok {
		if t.Logging.TensorBoard == nil {
			t.Logging.TensorBoard = &TensorBoardConfig{}
		}
		t.Logging.TensorBoard.Enabled = v
	}
	if v, ok := flags["logdir"].(string); ok && v != "" {
		if t.Logging.TensorBoard == nil {
			t.Logging.TensorBoard = &TensorBoardConfig{}
		}
		t.Logging.TensorBoard.LogDir = v
	}
	if v, ok := flags["tensorboard-image"].(string); ok && v != "" {
		if t.Logging.TensorBoard == nil {
			t.Logging.TensorBoard = &TensorBoardConfig{}
		}
		t.Logging.TensorBoard.Image = v
	}
	return nil
}

// applyFrameworkOverrides handles nproc-per-node, slots-per-worker, gputopology,
// mounts-on-launcher, chief, evaluator, and ps.
func applyFrameworkOverrides(t *Task, flags map[string]interface{}) error {
	if err := setStrFlag(flags, "nproc-per-node", &t.Framework.Options.NprocPerNode); err != nil {
		return err
	}
	if err := setIntFlag(flags, "slots-per-worker", &t.Framework.Options.SlotsPerWorker); err != nil {
		return err
	}
	if v, ok := flags["gputopology"].(bool); ok && v {
		t.Framework.Options.GPUTopology = true
		// v1 semantics: gputopology implies host networking and topology labels.
		t.HostNetwork = true
		if t.Labels == nil {
			t.Labels = make(map[string]string)
		}
		t.Labels["gpu-topology"] = "true"
		t.Labels["gpu-topology-replica"] = "true"
	}
	if err := setBoolFlag(flags, "mounts-on-launcher", &t.Framework.Options.MountsOnLauncher); err != nil {
		return err
	}

	// Map CLI flags to role sections (migration from FrameworkConfig)
	if v, ok := flags["chief"].(bool); ok && v {
		if t.Chief == nil {
			t.Chief = &RoleConfig{}
		}
	}
	if v, ok := flags["evaluator"].(bool); ok && v {
		if t.Evaluator == nil {
			t.Evaluator = &RoleConfig{}
		}
	}
	if v, ok := flags["ps"].(int); ok && v > 0 {
		if t.PS == nil {
			t.PS = &RoleConfig{}
		}
		t.PS.Replicas = v
	}
	return nil
}

// applyAffinityOverride applies a single string affinity flag to the Affinity
// struct, creating it if necessary. Returns an error if the flag value is not
// a string.
func applyAffinityOverride(flags map[string]interface{}, flag string, affinity **Affinity) error {
	v, exists := flags[flag]
	if !exists {
		return nil
	}
	s, ok := v.(string)
	if !ok {
		return fmt.Errorf("flag %q: expected string, got %T", flag, v)
	}
	if s == "" {
		return nil
	}
	if *affinity == nil {
		*affinity = &Affinity{}
	}
	switch flag {
	case "affinity-policy":
		(*affinity).Policy = s
	case "affinity-constraint":
		(*affinity).Constraint = s
	case "affinity-target":
		(*affinity).Target = s
	}
	return nil
}

// splitKV splits "key=value" into (key, value).
func splitKV(s string) (string, string) {
	parts := strings.SplitN(s, "=", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return s, ""
}

// parseTolerationFlag parses a CLI toleration string into a Toleration.
// Format: key=value:effect (e.g., "gpu=true:NoSchedule")
// Or:     key:effect (e.g., "node.kubernetes.io/not-ready:NoExecute")
// Returns nil if s is empty. Returns an error if the input is non-empty but
// the toleration key cannot be parsed.
func parseTolerationFlag(s string) (*Toleration, error) {
	if s == "" {
		return nil, nil
	}

	tol := &Toleration{}

	parts := strings.SplitN(s, ":", 3)
	keyValue := parts[0]

	// Split by = to separate key from value
	eqIdx := strings.Index(keyValue, "=")
	if eqIdx >= 0 {
		tol.Key = keyValue[:eqIdx]
		tol.Value = keyValue[eqIdx+1:]
		tol.Operator = "Equal"
	} else {
		tol.Key = keyValue
		tol.Operator = "Exists"
	}

	if tol.Key == "" {
		return nil, fmt.Errorf("invalid toleration %q: key is required", s)
	}

	if len(parts) >= 2 {
		tol.Effect = parts[1]
	}
	if len(parts) >= 3 {
		seconds, err := strconv.ParseInt(parts[2], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid toleration %q: toleration seconds must be an integer, got %q", s, parts[2])
		}
		if seconds < 0 {
			return nil, fmt.Errorf("invalid toleration %q: toleration seconds must be non-negative", s)
		}
		tol.TolerationSeconds = &seconds
	}

	return tol, nil
}

// applyRoleResourceOverrides handles the v1 per-role resource flags:
// ps-cpu/ps-memory/ps-gpus, chief-cpu/chief-memory,
// evaluator-cpu/evaluator-memory, worker-cpu/worker-memory, and their
// -limit variants (which write the role's Limits overlay without touching
// requests). It runs after applyFrameworkOverrides (which creates the
// roles), so role-specific flags override the generic --cpu/--memory —
// the v1 precedence.
func applyRoleResourceOverrides(t *Task, flags map[string]interface{}) error {
	setRoleStr := func(key, resource string, role **RoleConfig) error {
		v, exists := flags[key]
		if !exists {
			return nil
		}
		s, ok := v.(string)
		if !ok {
			return fmt.Errorf("flag %q: expected string, got %T", key, v)
		}
		if s != "" {
			ensureRole(role).Resources[resource] = s
		}
		return nil
	}

	setRoleLimitStr := func(key, resource string, role **RoleConfig) error {
		v, exists := flags[key]
		if !exists {
			return nil
		}
		s, ok := v.(string)
		if !ok {
			return fmt.Errorf("flag %q: expected string, got %T", key, v)
		}
		if s != "" {
			ensureRoleLimits(role).Limits[resource] = s
		}
		return nil
	}

	if err := setRoleLimitStr("ps-cpu-limit", "cpu", &t.PS); err != nil {
		return err
	}
	if err := setRoleLimitStr("ps-memory-limit", "memory", &t.PS); err != nil {
		return err
	}
	if err := setRoleLimitStr("chief-cpu-limit", "cpu", &t.Chief); err != nil {
		return err
	}
	if err := setRoleLimitStr("chief-memory-limit", "memory", &t.Chief); err != nil {
		return err
	}
	if err := setRoleLimitStr("evaluator-cpu-limit", "cpu", &t.Evaluator); err != nil {
		return err
	}
	if err := setRoleLimitStr("evaluator-memory-limit", "memory", &t.Evaluator); err != nil {
		return err
	}
	if v, ok := flags["worker-cpu-limit"].(string); ok && v != "" {
		ensureLimitsTarget(t)["cpu"] = v
	}
	if v, ok := flags["worker-memory-limit"].(string); ok && v != "" {
		ensureLimitsTarget(t)["memory"] = v
	}

	if err := setRoleStr("ps-cpu", "cpu", &t.PS); err != nil {
		return err
	}
	if err := setRoleStr("ps-memory", "memory", &t.PS); err != nil {
		return err
	}
	if gpus, ok := flags["ps-gpus"].(int); ok && gpus > 0 {
		ensureRole(&t.PS).Resources["nvidia.com/gpu"] = strconv.Itoa(gpus)
	}
	if err := setRoleStr("chief-cpu", "cpu", &t.Chief); err != nil {
		return err
	}
	if err := setRoleStr("chief-memory", "memory", &t.Chief); err != nil {
		return err
	}
	if err := setRoleStr("evaluator-cpu", "cpu", &t.Evaluator); err != nil {
		return err
	}
	if err := setRoleStr("evaluator-memory", "memory", &t.Evaluator); err != nil {
		return err
	}

	// Worker flags target Worker, else Master (single-node PyTorch), else a
	// new Worker — the same resolution as the generic --cpu/--memory overrides.
	if v, ok := flags["worker-cpu"].(string); ok && v != "" {
		ensureResourceTarget(t)["cpu"] = v
	}
	if v, ok := flags["worker-memory"].(string); ok && v != "" {
		ensureResourceTarget(t)["memory"] = v
	}
	return nil
}

// ensureRole returns the role config, creating it (with an empty Resources
// map) when nil.
func ensureRole(rc **RoleConfig) *RoleConfig {
	if *rc == nil {
		*rc = &RoleConfig{}
	}
	if (*rc).Resources == nil {
		(*rc).Resources = Resources{}
	}
	return *rc
}

// ensureRoleLimits returns the role config, creating it (with an empty
// Limits map) when nil.
func ensureRoleLimits(rc **RoleConfig) *RoleConfig {
	if *rc == nil {
		*rc = &RoleConfig{}
	}
	if (*rc).Limits == nil {
		(*rc).Limits = Resources{}
	}
	return *rc
}

// ensureLimitsTarget returns the Limits map that worker-scoped limit
// overrides apply to, mirroring ensureResourceTarget's role resolution.
func ensureLimitsTarget(t *Task) Resources {
	if t.Worker != nil {
		if t.Worker.Limits == nil {
			t.Worker.Limits = Resources{}
		}
		return t.Worker.Limits
	}
	if t.Master != nil {
		if t.Master.Limits == nil {
			t.Master.Limits = Resources{}
		}
		return t.Master.Limits
	}
	t.Worker = &Worker{Replicas: 1}
	t.Worker.Limits = Resources{}
	return t.Worker.Limits
}

// code-sync storage constants. v1 wired sync through a shared code-sync
// emptyDir volume mounted by both the sync init container and the main
// container; v2 storages require a size for emptyDir volumes (v1 was
// unlimited).
const (
	codeSyncStorageName = "code-sync"
	codeSyncStorageSize = "10Gi"
)

// applySyncOverrides handles the v1 code-sync flags: sync-mode, sync-source,
// and sync-image. Synced code lands in $workingDir/code (default /root) and
// is exchanged through a shared code-sync emptyDir volume mounted by both
// the sync init container and the main container — without it the synced
// files die with the init container's filesystem.
func applySyncOverrides(t *Task, flags map[string]interface{}) error {
	mode, _ := flags["sync-mode"].(string)
	if mode == "" {
		return nil
	}
	source, _ := flags["sync-source"].(string)
	if source == "" {
		return fmt.Errorf("--sync-source is required when --sync-mode is set")
	}
	image, _ := flags["sync-image"].(string)

	entry := SyncEntry{Image: image}
	switch strings.ToLower(mode) {
	case "git":
		entry.Git = source
	case "rsync":
		entry.Rsync = source
	case "hdfs":
		entry.HDFS = source
	default:
		return fmt.Errorf("invalid --sync-mode %q: must be git, rsync, or hdfs", mode)
	}

	wd := t.WorkingDir
	if wd == "" {
		wd = "/root"
	}
	entry.LocalPath = path.Join(wd, "code")

	t.Storages = append(t.Storages, Storage{
		Name:      codeSyncStorageName,
		MountPath: entry.LocalPath,
		Tmp:       codeSyncStorageSize,
	})
	entry.Mounts = []Mount{{Name: codeSyncStorageName}}

	t.Sync = append(t.Sync, entry)
	return nil
}
