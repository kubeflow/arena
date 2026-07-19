package task

import (
	"fmt"
	"strings"

	"github.com/kubeflow/arena/pkg/constants"
)

// ApplyOverrides applies flag-based overrides to a Task struct.
// Flag keys match the CLI flag names (without the leading --).
// Returns an error when a flag value has an unexpected type.
func ApplyOverrides(t *Task, flags map[string]interface{}) error {
	setStr := func(flag string, target *string) error {
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
	setInt := func(flag string, target *int) error {
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
	setBool := func(flag string, target *bool) error {
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

	// Identity
	if err := setStr("name", &t.Name); err != nil {
		return err
	}
	if err := setStr("namespace", &t.Namespace); err != nil {
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

	// Task
	if err := setStr("framework", &t.Framework.Name); err != nil {
		return err
	}
	if err := setStr("image", &t.Image); err != nil {
		return err
	}
	if err := setStr("working-dir", &t.WorkingDir); err != nil {
		return err
	}
	if err := setStr("shell", &t.Shell); err != nil {
		return err
	}

	if err := setStr("run", &t.Run); err != nil {
		return err
	}

	// Scale + Resources — auto-create Worker only when overrides target it
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
			return fmt.Errorf("flag \"workers\": expected int, got %T", v)
		}
		if n > 0 {
			t.Worker.Replicas = n
		}
	}

	// Resources
	if gpus, ok := flags["gpus"].(int); ok && gpus > 0 {
		ensureWorker()
		if t.Worker.Resources == nil {
			t.Worker.Resources = Resources{}
		}
		t.Worker.Resources["nvidia.com/gpu"] = fmt.Sprintf("%d", gpus)
	}
	if gpuType, ok := flags["gpu-type"].(string); ok && gpuType != "" {
		if t.Scheduling.NodeSelector == nil {
			t.Scheduling.NodeSelector = make(map[string]string)
		}
		t.Scheduling.NodeSelector["nvidia.com/gpu.product"] = gpuType
	}

	if v, ok := flags["cpus"].(string); ok && v != "" {
		ensureWorker()
		if t.Worker.Resources == nil {
			t.Worker.Resources = Resources{}
		}
		t.Worker.Resources["cpu"] = v
	}
	if v, ok := flags["mem"].(string); ok && v != "" {
		ensureWorker()
		if t.Worker.Resources == nil {
			t.Worker.Resources = Resources{}
		}
		t.Worker.Resources["memory"] = v
	}

	if shm, ok := flags["shm"].(string); ok && shm != "" {
		t.Storages = append(t.Storages, Storage{
			Name:      "shm",
			SHM:       shm,
			MountPath: constants.DefaultSHMMountPath,
		})
	}

	if devices, ok := flags["device"].([]string); ok {
		ensureWorker()
		if t.Worker.Resources == nil {
			t.Worker.Resources = Resources{}
		}
		for _, d := range devices {
			k, v := splitKV(d)
			if k != "" {
				t.Worker.Resources[k] = v
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

	// Scheduling
	if err := setInt("priority", &t.Scheduling.Priority); err != nil {
		return err
	}
	if err := setStr("priority-class-name", &t.Scheduling.PriorityClassName); err != nil {
		return err
	}
	if err := setBool("gang", &t.Scheduling.Gang.Enabled); err != nil {
		return err
	}
	if err := setStr("scheduler-name", &t.Scheduling.SchedulerName); err != nil {
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
			parsed := parseTolerationFlag(tol)
			if parsed != nil {
				t.Scheduling.Tolerations = append(t.Scheduling.Tolerations, *parsed)
			}
		}
	}

	// Queue
	if err := setStr("queue", &t.Scheduling.Queue); err != nil {
		return err
	}

	// Lifecycle
	if err := setStr("clean-pod-policy", &t.Lifecycle.CleanPodPolicy); err != nil {
		return err
	}
	if err := setStr("active-deadline", &t.Lifecycle.ActiveDeadline); err != nil {
		return err
	}
	if err := setStr("ttl-after-finished", &t.Lifecycle.TTLAfterFinished); err != nil {
		return err
	}
	if err := setStr("success-policy", &t.Lifecycle.SuccessPolicy); err != nil {
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

	// Runtime
	if err := setStr("image-pull-policy", &t.ImagePullPolicy); err != nil {
		return err
	}
	if err := setStr("service-account", &t.ServiceAccount); err != nil {
		return err
	}
	if err := setStr("restart", &t.Restart); err != nil {
		return err
	}
	if err := setBool("host-network", &t.HostNetwork); err != nil {
		return err
	}
	if err := setBool("host-ipc", &t.HostIPC); err != nil {
		return err
	}
	if err := setBool("host-pid", &t.HostPID); err != nil {
		return err
	}

	if secrets, ok := flags["image-pull-secret"].([]string); ok {
		t.ImagePullSecrets = append(t.ImagePullSecrets, secrets...)
	}

	// Logging — direct assignment for tensorboard fields
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

	// Framework-specific options
	if err := setStr("nproc-per-node", &t.Framework.Options.NprocPerNode); err != nil {
		return err
	}
	if err := setInt("slots-per-worker", &t.Framework.Options.SlotsPerWorker); err != nil {
		return err
	}
	if err := setBool("gpu-topology", &t.Framework.Options.GPUTopology); err != nil {
		return err
	}
	if err := setBool("mounts-on-launcher", &t.Framework.Options.MountsOnLauncher); err != nil {
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
// Returns nil for invalid/unparseable input.
func parseTolerationFlag(s string) *Toleration {
	if s == "" {
		return nil
	}

	tol := &Toleration{}

	// Split by : to separate key=value from effect
	colonIdx := strings.LastIndex(s, ":")
	keyValue := s
	if colonIdx >= 0 {
		keyValue = s[:colonIdx]
		tol.Effect = s[colonIdx+1:]
	}

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
		return nil
	}

	return tol
}
