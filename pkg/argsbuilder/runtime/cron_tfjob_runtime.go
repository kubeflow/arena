package runtime

import (
	"github.com/kubeflow/arena/pkg/apis/types"
)

type defaultCronTFRuntime struct {
	name string
}

func NewDefaultCronTFJobRuntime() types.TFRuntime {
	return &defaultCronTFRuntime{
		name: "cron",
	}
}

func (d *defaultCronTFRuntime) Check(tf *types.SubmitTFJobArgs) error {
	return nil
}

func (d *defaultCronTFRuntime) Transform(tf *types.SubmitTFJobArgs) error {
	return nil
}

func (d *defaultCronTFRuntime) GetChartName() string {
	return d.name
}

func (d *defaultCronTFRuntime) IsDefault() bool {
	return true
}
