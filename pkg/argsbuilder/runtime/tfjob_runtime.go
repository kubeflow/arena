package runtime

import (
	"github.com/kubeflow/arena/pkg/apis/types"
)

type defaultTFRuntime struct {
	name string
}

func NewDefaultTFJobRuntime() types.TFRuntime {
	return &defaultTFRuntime{
		name: "tfjob",
	}
}

func (d *defaultTFRuntime) Check(tf *types.SubmitTFJobArgs) error {
	return nil
}

func (d *defaultTFRuntime) Transform(tf *types.SubmitTFJobArgs) error {
	return nil
}

func (d *defaultTFRuntime) GetChartName() string {
	return d.name
}

func (d *defaultTFRuntime) IsDefault() bool {
	return true
}
