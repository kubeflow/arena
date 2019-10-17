package types

import (
	"errors"
	"fmt"
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
	ErrTooManyPods = errors.New(`too many pods have been found`)
	ErrNotFoundTargetPod = errors.New(`not found target pod`)
	ErrInvalidUsage = errors.New(`invalid usage for arena exec`)
)

var SERVING_CHARTS = map[string]string{
	"tensorflow-serving-0.2.0":        "Tensorflow",
	"tensorrt-inference-server-0.0.1": "TensorRT",
}

func KeyMapServingType(servingKey string) ServingType {
	switch servingKey {
	case "tf", "tf-serving", "tensorflow-serving":
		return ServingTF
	case "trt", "trt-serving", "tensorrt-serving":
		return ServingTRT
	case "custom", "custom-serving":
		return ServingCustom
	default:
		return ServingType("")
	}
}
func CheckServingTypeIsOk(stype string) error {
	if stype == "" {
		return nil
	}
	if KeyMapServingType(stype) == ServingType("") {
		return fmt.Errorf("unknow serving type: %s", stype)
	}
	return nil
}
