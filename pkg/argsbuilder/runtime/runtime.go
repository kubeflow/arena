package runtime

import (
	"github.com/kubeflow/arena/pkg/apis/types"
)

var tfRuntimes = []types.TFRuntime{}

func init() {
	tfRuntimes = append(tfRuntimes, NewDefaultTFJobRuntime())
}

func GetTFRuntime(name string) types.TFRuntime {
	var d types.TFRuntime
	for _, tfRuntime := range tfRuntimes {
		if tfRuntime.IsDefault() {
			d = tfRuntime
		}
		if name == tfRuntime.GetChartName() {
			return tfRuntime
		}
	}
	return d
}
