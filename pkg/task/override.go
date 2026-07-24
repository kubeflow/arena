package task

import (
	"fmt"
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
// Flag keys match the CLI flag names (without the leading --).
// Returns an error when a flag value has an unexpected type.
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

// applyResourceOverrides handles image, workers, gpus, gpu-type, cpus, memory,
// shm, device, envs, data, data-dir, and config-file.
func applyResourceOverrides(t *Task, flags map[string]interface{}) error {
	if err := setStrFlag(flags, "image", &t.Image); err != nil {
		return err
	}

	ensureWorker := func() {
		if t.Worker == nil {
			t.Worker = &Worker{}
		}
	}

	// ensureResourceTarget returns the Resources map that resource overrides
	// should be applied to. When Worker exists, it returns Worker.Resources.
	// When Worker is nil but Master exists (single-node PyTorch), it returns
	// Master.Resources. Otherwise it creates a Worker with Replicas=1.
	ensureResourceTarget := func() Resources {
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
		target := ensureResourceTarget()
		target["nvidia.com/gpu"] = strconv.Itoa(gpus)
	}
	if gpuType, ok := flags["gpu-type"].(string); ok && gpuType != "" {
		if t.Scheduling.NodeSelector == nil {
			t.Scheduling.NodeSelector = make(map[string]string)
		}
		t.Scheduling.NodeSelector["nvidia.com/gpu.product"] = gpuType
	}

	if v, ok := flags["cpus"].(string); ok && v != "" {
		target := ensureResourceTarget()
		target["cpu"] = v
	}
	if v, ok := flags["mem"].(string); ok && v != "" {
		target := ensureResourceTarget()
		target["memory"] = v
	}

	if shm, ok := flags["shm"].(string); ok && shm != "" {
		t.Storages = append(t.Storages, Storage{
			Name:      "shm",
			SHM:       shm,
			MountPath: constants.DefaultSHMMountPath,
		})
	}

	if devices, ok := flags["device"].([]string); ok {
		target := ensureResourceTarget()
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
			parts := strings.SplitN(d, ":", 3)
			if len(parts) == 3 {
				t.Storages = append(t.Storages, Storage{
					Name:      parts[0],
					MountPath: parts[1],
					PVC:       parts[2],
				})
			}
		}
	}

	// Data-dir (hostpath volumes)
	if dataDirs, ok := flags["data-dir"].([]string); ok {
		for _, d := range dataDirs {
			parts := strings.SplitN(d, ":", 3)
			if len(parts) == 3 {
				t.Storages = append(t.Storages, Storage{
					Name:      parts[0],
					MountPath: parts[1],
					HostPath:  parts[2],
				})
			}
		}
	}

	// Config-file (configmap volumes)
	if configFiles, ok := flags["config-file"].([]string); ok {
		for _, c := range configFiles {
			parts := strings.SplitN(c, ":", 3)
			if len(parts) == 3 {
				t.Storages = append(t.Storages, Storage{
					Name:      parts[0],
					MountPath: parts[1],
					ConfigMap: parts[2],
				})
			}
		}
	}
	return nil
}

// applySchedulingOverrides handles scheduler, queue, priority, gang, affinity,
// selector, and toleration.
func applySchedulingOverrides(t *Task, flags map[string]interface{}) error {
	if err := setIntFlag(flags, "priority", &t.Scheduling.Priority); err != nil {
		return err
	}
	if err := setStrFlag(flags, "priority-class-name", &t.Scheduling.PriorityClassName); err != nil {
		return err
	}
	if err := setBoolFlag(flags, "gang", &t.Scheduling.Gang.Enabled); err != nil {
		return err
	}
	if err := setStrFlag(flags, "scheduler-name", &t.Scheduling.SchedulerName); err != nil {
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

	if err := setStrFlag(flags, "queue", &t.Scheduling.Queue); err != nil {
		return err
	}
	return nil
}

// applyLifecycleOverrides handles clean-pod-policy, active-deadline,
// ttl-after-finished, success-policy, and backoff-limit.
func applyLifecycleOverrides(t *Task, flags map[string]interface{}) error {
	if err := setStrFlag(flags, "clean-pod-policy", &t.Lifecycle.CleanPodPolicy); err != nil {
		return err
	}
	if err := setStrFlag(flags, "active-deadline", &t.Lifecycle.ActiveDeadline); err != nil {
		return err
	}
	if err := setStrFlag(flags, "ttl-after-finished", &t.Lifecycle.TTLAfterFinished); err != nil {
		return err
	}
	if err := setStrFlag(flags, "success-policy", &t.Lifecycle.SuccessPolicy); err != nil {
		return err
	}

	if v, exists := flags["backoff-limit"]; exists {
		n, ok := v.(int)
		if !ok {
			return fmt.Errorf("flag %q: expected int, got %T", "backoff-limit", v)
		}
		if n >= 0 {
			t.Lifecycle.BackoffLimit = &n
		}
	}
	return nil
}

// applyRuntimeOverrides handles run, shell, working-dir, image-pull-policy,
// service-account, restart, host-network, host-ipc, host-pid, and image-pull-secret.
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
	if err := setStrFlag(flags, "restart", &t.Restart); err != nil {
		return err
	}
	if err := setBoolFlag(flags, "host-network", &t.HostNetwork); err != nil {
		return err
	}
	if err := setBoolFlag(flags, "host-ipc", &t.HostIPC); err != nil {
		return err
	}
	if err := setBoolFlag(flags, "host-pid", &t.HostPID); err != nil {
		return err
	}

	if secrets, ok := flags["image-pull-secret"].([]string); ok {
		t.ImagePullSecrets = append(t.ImagePullSecrets, secrets...)
	}
	return nil
}

// applyLoggingOverrides handles tensorboard, tensorboard-logdir, and tensorboard-image.
func applyLoggingOverrides(t *Task, flags map[string]interface{}) error {
	if v, ok := flags["tensorboard"].(bool); ok {
		if t.Logging.TensorBoard == nil {
			t.Logging.TensorBoard = &TensorBoardConfig{}
		}
		t.Logging.TensorBoard.Enabled = v
	}
	if v, ok := flags["tensorboard-logdir"].(string); ok && v != "" {
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

// applyFrameworkOverrides handles nproc-per-node, slots-per-worker, gpu-topology,
// mounts-on-launcher, chief, evaluator, and ps-count.
func applyFrameworkOverrides(t *Task, flags map[string]interface{}) error {
	if err := setStrFlag(flags, "nproc-per-node", &t.Framework.Options.NprocPerNode); err != nil {
		return err
	}
	if err := setIntFlag(flags, "slots-per-worker", &t.Framework.Options.SlotsPerWorker); err != nil {
		return err
	}
	if err := setBoolFlag(flags, "gpu-topology", &t.Framework.Options.GPUTopology); err != nil {
		return err
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
	if v, ok := flags["ps-count"].(int); ok && v > 0 {
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
