package training

import (
	"github.com/kubeflow/arena/pkg/apis/types"
)

var DefaultCommonSubmitArgs = types.CommonSubmitArgs{
	WorkingDir:  "/root",
	WorkerCount: 1,
}

var DefaultSubmitTensorboardArgs = types.SubmitTensorboardArgs{
	TensorboardImage: "registry.cn-zhangjiakou.aliyuncs.com/tensorflow-samples/tensorflow:1.12.0-devel",
	TrainingLogdir:   "/training_logs",
}
