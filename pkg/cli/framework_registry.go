package cli

import (
	"strings"

	"github.com/kubeflow/arena/pkg/constants"
)

// frameworkDef describes one arena submit framework type: its v2 canonical
// name, the label value preserving the user's original choice, the CRD kind
// it renders, and every alias arena v1 accepted on the command line.
type frameworkDef struct {
	canonical string // v2 canonical framework name ("" when v2 does not support it)
	original  string // label value for the user's original choice
	kind      string // CRD kind ("" when none)
	cmdName   string // v1 primary submit subcommand name ("" when v2 does not support it)
	aliases   []string
}

// frameworkRegistry is the single source of truth for submit framework
// types. Alias lists mirror `arena v0.15.4 submit <type> -h`.
var frameworkRegistry = []frameworkDef{
	{canonical: constants.FrameworkPyTorch, original: constants.FrameworkPyTorch, kind: constants.KindPyTorchJob,
		cmdName: "pytorchjob",
		aliases: []string{constants.FrameworkPyTorch, "pytorchjob"}},
	{canonical: constants.FrameworkTensorFlow, original: constants.FrameworkTensorFlow, kind: constants.KindTFJob,
		cmdName: "tfjob",
		// Alias order is user-visible: acceptedFrameworkTypes() renders it in
		// the unsupported-type error, matching v1's tf/tfjob/tensorflow list.
		aliases: []string{"tf", "tfjob", constants.FrameworkTensorFlow}},
	{canonical: constants.FrameworkMPI, original: constants.FrameworkMPI, kind: constants.KindMPIJob,
		cmdName: "mpijob",
		aliases: []string{constants.FrameworkMPI, "mpijob", "mj"}},
	{canonical: constants.FrameworkHorovod, original: constants.FrameworkHorovod, kind: constants.KindMPIJob,
		cmdName: "horovodjob",
		aliases: []string{constants.FrameworkHorovod, "horovodjob", "hj"}},
	{canonical: constants.FrameworkDeepSpeed, original: constants.FrameworkDeepSpeed, kind: constants.KindMPIJob,
		cmdName: "deepspeedjob",
		aliases: []string{constants.FrameworkDeepSpeed, "deepspeedjob", "dp"}},
	// arena v1 types with no v2 provider yet. They are recognized so submit
	// can fail fast with a dedicated message instead of the generic
	// unknown-type error.
	{original: "ray", aliases: []string{"ray", "rayjob", "rj"}},
	{original: "spark", aliases: []string{"spark", "sparkjob"}},
	{original: "volcano", aliases: []string{"volcano", "volcanojob", "vj"}},
	{original: "et", aliases: []string{"et", "etjob"}},
}

// lookupFramework resolves a user-supplied type string. It returns the v2
// canonical framework name, and whether the string is a known arena v1 type
// that v2 does not support yet.
func lookupFramework(s string) (canonical string, v1TypeUnsupported bool) {
	key := strings.ToLower(s)
	for _, def := range frameworkRegistry {
		for _, a := range def.aliases {
			if a == key {
				return def.canonical, def.canonical == ""
			}
		}
	}
	return "", false
}

// normalizeFramework maps framework aliases to canonical names.
func normalizeFramework(s string) string {
	canonical, _ := lookupFramework(s)
	return canonical
}

// originalFramework preserves the user's original framework choice for labeling.
func originalFramework(s string) string {
	key := strings.ToLower(s)
	for _, def := range frameworkRegistry {
		for _, a := range def.aliases {
			if a == key {
				return def.original
			}
		}
	}
	return ""
}

// frameworkToKind maps a canonical framework name to its CRD kind.
func frameworkToKind(framework string) string {
	for _, def := range frameworkRegistry {
		if def.canonical == framework {
			return def.kind
		}
	}
	return ""
}

// kindToFramework maps a CRD kind to its canonical framework name.
// For unrecognized kinds, it returns the lowercased kind as a fallback.
func kindToFramework(kind string) string {
	switch kind {
	case constants.KindPyTorchJob:
		return constants.FrameworkPyTorch
	case constants.KindTFJob:
		return constants.FrameworkTensorFlow
	case constants.KindMPIJob:
		return constants.FrameworkMPI
	default:
		return strings.ToLower(kind)
	}
}

// isMPIFamily returns true if the framework uses the MPIJob CRD.
func isMPIFamily(framework string) bool {
	return framework == constants.FrameworkMPI ||
		framework == constants.FrameworkHorovod ||
		framework == constants.FrameworkDeepSpeed
}

// v2FrameworkNames returns the canonical framework names arena-v2 supports,
// in registry order.
func v2FrameworkNames() []string {
	names := make([]string, 0, len(frameworkRegistry))
	for _, def := range frameworkRegistry {
		if def.canonical != "" {
			names = append(names, def.canonical)
		}
	}
	return names
}

// acceptedFrameworkTypes returns every framework type string the submit
// command line accepts, across all v2-supported aliases.
func acceptedFrameworkTypes() []string {
	var types []string
	for _, def := range frameworkRegistry {
		if def.canonical != "" {
			types = append(types, def.aliases...)
		}
	}
	return types
}
