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
	"path/filepath"
	"regexp"
	"strings"
)

// RestrictedNameChars collects the characters allowed to represent a name, normally used to validate container and volume names.
const restrictedNameChars = `[a-zA-Z0-9][a-zA-Z0-9_.-]`

// RestrictedNamePattern is a regular expression to validate names against the collection of restricted characters.
var restrictedNamePattern = regexp.MustCompile(`^` + restrictedNameChars + `+$`)

func ValidateDatasets(dataset []string) (err error) {
	tempPVCList := make(map[string]string)
	for _, dataset := range dataset {
		parts := strings.Split(dataset, ":")
		if len(parts) != 2 {
			return fmt.Errorf("dataset %s has incorrect format, should like data_name:/data0", dataset)
		}
		err = validateVPCName(parts[0])
		if err != nil {
			return err
		}
		if _, ok := tempPVCList[parts[0]]; ok {
			return fmt.Errorf("Your PVC name: %s is repeated, please check it.", parts[0])
		} else {
			tempPVCList[parts[0]] = ""
		}

		// validate mount destination
		err = validateMountDestination(parts[1])
		if err != nil {
			return err
		}
	}

	return nil
}

func validateVPCName(name string) error {
	if len(name) == 1 {
		return fmt.Errorf("volume name is too short, names should be at least two alphanumeric characters")
	}
	if !restrictedNamePattern.MatchString(name) {
		return fmt.Errorf("%q includes invalid characters for a local volume name, only %q are allowed. If you intended to pass a host directory, use absolute path", name, restrictedNameChars)
	}
	return nil
}

// validate Absolute path
func validateAbsolute(p string) error {
	p = convertSlash(p)
	if filepath.IsAbs(p) {
		return nil
	}
	return fmt.Errorf("invalid dataDir: '%s' must be absolute", p)
}

// ValidateMountDestination validates the destination.
// Currently, we have only two obvious rule for validation:
//  - path must not be "/"
//  - path must be absolute
func validateMountDestination(dest string) error {
	if err := validateNotRoot(dest); err != nil {
		return err
	}
	return validateAbsolute(dest)
}

func validateHostPath(path string) error {
	if path == "" {
		return errInvalidSpec(path)
	}
	return validateMountDestination(path)
}

func convertSlash(p string) string {
	return filepath.ToSlash(p)
}

func validateNotRoot(p string) error {
	p = filepath.Clean(convertSlash(p))
	if p == "/" {
		return fmt.Errorf("invalid specification: dataDir can't be '/'")
	}
	return nil
}

// ParseDataDirRaw parse DataDir into hostPath, containerPath
func ParseDataDirRaw(raw string) (hostPath, containerPath string, err error) {
	arr, err := splitRaw(convertSlash(raw))
	if err != nil {
		return "", "", err
	}

	switch len(arr) {
	case 1:
		hostPath = arr[0]
		containerPath = arr[0]
	case 2:
		hostPath = arr[0]
		containerPath = arr[1]
	default:
		return "", "", errInvalidSpec(raw)
	}

	err = validateHostPath(hostPath)
	if err != nil {
		return "", "", err
	}

	err = validateHostPath(containerPath)
	if err != nil {
		return "", "", err
	}

	return hostPath, containerPath, nil
}

func splitRaw(raw string) ([]string, error) {
	if strings.Count(raw, ":") > 1 {
		return nil, errInvalidSpec(raw)
	}

	arr := strings.Split(raw, ":")
	if arr[0] == "" {
		return nil, errInvalidSpec(raw)
	}
	return arr, nil
}

func errInvalidSpec(spec string) error {
	return fmt.Errorf("invalid DataDir: '%s'", spec)
}
