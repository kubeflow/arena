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
)

const dns1123SubdomainFmt string = dns1123LabelFmt + "(\\." + dns1123LabelFmt + ")*"
const dns1123SubdomainErrorMsg string = "a DNS-1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character"
const DNS1123SubdomainMaxLength int = 253

const dns1123LabelFmt string = "[a-z0-9]([-a-z0-9]*[a-z0-9])?"
const dns1123LabelErrMsg string = "a DNS-1123 label must consist of lower case alphanumeric characters or '-', and must start and end with an alphanumeric character"
const DNS1123LabelMaxLength int = 63

var dns1123LabelRegexp = regexp.MustCompile("^" + dns1123LabelFmt + "$")

var dns1123SubdomainRegexp = regexp.MustCompile("^" + dns1123SubdomainFmt + "$")

func ValidateJobName(value string) error {
	if len(value) > DNS1123LabelMaxLength {
		return fmt.Errorf("The len of name %d is too long, it should be less than %d",
			len(value),
			DNS1123SubdomainMaxLength)
	}
	if !dns1123LabelRegexp.MatchString(value) {
		return fmt.Errorf("The job name must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character.")
	}
	return nil
}
