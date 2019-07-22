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
	"fmt"
	"regexp"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const dns1123SubdomainFmt string = dns1123LabelFmt + "(\\." + dns1123LabelFmt + ")*"
const dns1123SubdomainErrorMsg string = "a DNS-1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character"
const DNS1123SubdomainMaxLength int = 253

const dns1123LabelFmt string = "[a-z0-9]([-a-z0-9]*[a-z0-9])?"
const dns1123LabelErrMsg string = "a DNS-1123 label must consist of lower case alphanumeric characters or '-', and must start and end with an alphanumeric character"
const DNS1123LabelMaxLength int = 63

// Job Max lenth should be 49
const JobMaxLength int = 49

var dns1123LabelRegexp = regexp.MustCompile("^" + dns1123LabelFmt + "$")

var dns1123SubdomainRegexp = regexp.MustCompile("^" + dns1123SubdomainFmt + "$")

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
	_, err := client.Scheduling().PriorityClasses().Get(name, metav1.GetOptions{})
	if err != nil && errors.IsNotFound(err) {
		err = fmt.Errorf("The priority %s doesn't exist. Please check with `kubectl get pc` to get a valid priority.", name)
	}

	return err
}
