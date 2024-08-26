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

package util

import (
	"context"
	"fmt"
	"github.com/kubeflow/arena/pkg/apis/types"
	"k8s.io/apimachinery/pkg/util/validation"
	"regexp"
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const dns1123LabelFmt string = "[a-z0-9]([-a-z0-9]*[a-z0-9])?"

// Job Max lenth should be 49
const JobMaxLength int = 49

var dns1123LabelRegexp = regexp.MustCompile("^" + dns1123LabelFmt + "$")

// ValidateJobName validates the job name, its length should less than 63, and match dns1123LabelFmt
func ValidateJobName(value string) error {
	if len(value) > JobMaxLength {
		return fmt.Errorf("The len %d of name %s is too long, it should be less than %d",
			len(value),
			value,
			JobMaxLength)
	}
	if !dns1123LabelRegexp.MatchString(value) {
		return fmt.Errorf("The job name must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character.")
	}
	return nil
}

// Check if PriorityClassName exists
func ValidatePriorityClassName(client *kubernetes.Clientset, name string) error {
	// client.SchedulingV1alpha1()
	_, err := client.SchedulingV1().PriorityClasses().Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil && errors.IsNotFound(err) {
		err = fmt.Errorf("The priority %s doesn't exist. Please check with `kubectl get pc` to get a valid priority.", name)
	}

	return err
}

func ValidateDevices(devices []string) error {
	for _, member := range devices {
		splits := strings.SplitN(member, "=", 2)
		if len(splits) != 2 || len(splits[0]) == 0 || len(splits[1]) == 0 {
			err := fmt.Errorf("Invalid device member %s, refer to amd.com/gpu=1.", member)
			return err
		}
		if errs := validation.IsQualifiedName(splits[0]); len(errs) != 0 {
			err := fmt.Errorf("Invalid device name %s is not Qualified", splits[0])
			return err
		}
		_, err := strconv.Atoi(splits[1])
		if err != nil {
			err = fmt.Errorf("Invalid device value %s should be a number, refer to amd.com/gpu=1.", member)
			return err
		}
		if strings.EqualFold(splits[0], types.NvidiaGPUResourceName) {
			err = fmt.Errorf("Invalid device member %s, use --gpus to set %s count.", types.NvidiaGPUResourceName, types.NvidiaGPUResourceName)
			return err
		}
	}
	return nil
}
