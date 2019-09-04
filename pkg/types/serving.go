package types

import (
	"errors"
)

// this file is used to define serving type

type ServingType string

// three serving types.
const (
	// tensorflow
	ServingTF ServingType = "TENSORFLOW"
	// tensorrt
	ServingTRT ServingType = "TENSORRT"
	// custom
	ServingCustom ServingType = "CUSTOM"
)

var (
	ErrNotFoundJobs = errors.New(`not found jobs under the assigned conditions.`)
	ErrTooManyJobs  = errors.New(`found jobs more than one,please use --version or --type to filter.`)
)

var SERVING_CHARTS = map[string]string{
	"tensorflow-serving-0.2.0":        "Tensorflow",
	"tensorrt-inference-server-0.0.1": "TensorRT",
}
