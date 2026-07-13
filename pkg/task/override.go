package task

import (
	"fmt"
	"strings"

	"github.com/kubeflow/arena/pkg/constants"
)

// ApplyOverrides applies flag-based overrides to a Task struct.
// Flag keys match the CLI flag names (without the leading --).
func ApplyOverrides(t *Task, flags map[string]interface{}) {
	setStr := func(flag string, target *string) {
		if v, ok := flags[flag].(string); ok && v != "" {
			*target = v
		}
	}
	setInt := func(flag string, target *int) {
		if v, ok := flags[flag].(int); ok && v > 0 {
			*target = v
		}
	}
	setBool := func(flag string, target *bool) {
		if v, ok := flags[flag].(bool); ok {
			*target = v
		}
	}

	// Identity
	setStr("name", &t.Name)
	setStr("namespace", &t.Namespace)

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
	setStr("framework", &t.Framework.Name)
	setStr("image", &t.Image)
	setStr("working-dir", &t.WorkingDir)
	setStr("shell", &t.Shell)

	if run, ok := flags["run"].(string); ok && run != "" {
		t.Run = run
	}

	// Scale + Resources — auto-create Worker only when overrides target it
	ensureWorker := func() {
		if t.Worker == nil {
			t.Worker = &Worker{}
		}
	}

	// Scale
	if v, ok := flags["workers"].(int); ok && v > 0 {
		ensureWorker()
		t.Worker.Replicas = v
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
			parts := splitN(d, ":", 3)
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
			parts := splitN(d, ":", 3)
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
			parts := splitN(c, ":", 3)
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
	setInt("priority", &t.Scheduling.Priority)
	setStr("priority-class-name", &t.Scheduling.PriorityClassName)
	setBool("gang", &t.Scheduling.Gang.Enabled)
	setStr("scheduler-name", &t.Scheduling.SchedulerName)

	// Direct assignment for affinity-policy and affinity-constraint
	if v, ok := flags["affinity-policy"].(string); ok && v != "" {
		if t.Scheduling.Affinity == nil {
			t.Scheduling.Affinity = &Affinity{}
		}
		t.Scheduling.Affinity.Policy = v
	}
	if v, ok := flags["affinity-constraint"].(string); ok && v != "" {
		if t.Scheduling.Affinity == nil {
			t.Scheduling.Affinity = &Affinity{}
		}
		t.Scheduling.Affinity.Constraint = v
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
	setStr("queue", &t.Scheduling.Queue)

	// Lifecycle
	setStr("clean-pod-policy", &t.Lifecycle.CleanPodPolicy)
	setStr("active-deadline", &t.Lifecycle.ActiveDeadline)
	setStr("ttl-after-finished", &t.Lifecycle.TTLAfterFinished)
	setStr("success-policy", &t.Lifecycle.SuccessPolicy)

	if bl, ok := flags["backoff-limit"].(int); ok && bl >= 0 {
		t.Lifecycle.BackoffLimit = &bl
	}

	// Runtime
	setStr("image-pull-policy", &t.ImagePullPolicy)
	setStr("service-account", &t.ServiceAccount)
	setStr("restart", &t.Restart)
	setBool("host-network", &t.HostNetwork)
	setBool("host-ipc", &t.HostIPC)
	setBool("host-pid", &t.HostPID)

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
	setStr("nproc-per-node", &t.Framework.Options.NprocPerNode)
	setInt("slots-per-worker", &t.Framework.Options.SlotsPerWorker)
	setBool("gpu-topology", &t.Framework.Options.GPUTopology)
	setBool("mounts-on-launcher", &t.Framework.Options.MountsOnLauncher)

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
}

// splitKV splits "key=value" into (key, value).
func splitKV(s string) (string, string) {
	parts := splitN(s, "=", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return s, ""
}

func splitN(s, sep string, n int) []string {
	result := make([]string, 0, n)
	for i := 0; i < n-1; i++ {
		idx := indexOf(s, sep)
		if idx < 0 {
			break
		}
		result = append(result, s[:idx])
		s = s[idx+len(sep):]
	}
	result = append(result, s)
	return result
}

func indexOf(s, sub string) int {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
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
