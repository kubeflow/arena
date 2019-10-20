// Copyright 2018 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	ErrNotFoundJobs      = errors.New(`not found jobs under the assigned conditions.`)
	ErrTooManyJobs       = errors.New(`found jobs more than one,please use --version or --type to filter.`)
	ErrTooManyPods       = errors.New(`too many pods have been found`)
	ErrNotFoundTargetPod = errors.New(`not found target pod`)
	ErrInvalidUsage      = errors.New(`invalid usage for arena exec`)
)

var SERVING_CHARTS = map[string]string{
	"tensorflow-serving-0.2.0":        "Tensorflow",
	"tensorrt-inference-server-0.0.1": "TensorRT",
}

// ServingJobInfo is use to print
type ServingJobPrinterInfo struct {
	JobPrinterInfo  `yaml:",inline" json:",inline"`
	Version         string `yaml:"version" json:"version"`
	Desired         int32  `yaml:"desired" json:"desired"`
	Available       int32  `yaml:"available" json:"available"`
	EndpointAddress string `yaml:"endpoint_address" json:"endpoint_address"`
	EndpointPorts   string `yaml:"endpoint_ports" json:"endpoint_ports"`
}

type ServingInstance struct {
	Instance     `yaml:",inline" json:",inline"`
	Ready        string `yaml:"ready" json:"ready"`
	RestartCount string `yaml:"restart_count" json:"restart_count"`
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
