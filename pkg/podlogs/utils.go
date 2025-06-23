// Copyright 2024 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package podlogs

import (
	"errors"
	"strconv"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	errInvalidSinceSecond = errors.New("failed to parse since seconds,invalid format,should like: ['1h','1m','1s','1h1m1s'...]")
)

type OuterRequestArgs struct {
	PodName      string
	Namespace    string
	Follow       bool
	Tail         int
	RetryCount   int
	RetryTimeout time.Duration
	SinceSeconds string
	SinceTime    string
	Timestamps   bool
	KubeClient   kubernetes.Interface
}

func ParseSinceTime(sinceTime string) (*metav1.Time, error) {
	if sinceTime == "" {
		return nil, nil
	}
	parsedTime, err := time.Parse(time.RFC3339, sinceTime)
	if err != nil {
		return nil, err
	}
	meta1Time := metav1.NewTime(parsedTime)
	return &meta1Time, nil

}

func ParseSinceSeconds(since string) (*int64, error) {
	if since == "" {
		return nil, nil
	}
	totalSeconds := int64(0)
	items := []string{}
	for i := 0; i < len(since); i++ {
		switch string(since[i]) {
		case "h":
			hour, err := strconv.ParseInt(strings.Join(items, ""), 10, 64)
			if err != nil {
				return nil, errInvalidSinceSecond
			}
			totalSeconds = totalSeconds + hour*3600
			items = []string{}
		case "m":
			m, err := strconv.ParseInt(strings.Join(items, ""), 10, 64)
			if err != nil {
				return nil, errInvalidSinceSecond
			}
			totalSeconds = totalSeconds + m*60
			items = []string{}
		case "s":
			second, err := strconv.ParseInt(strings.Join(items, ""), 10, 64)
			if err != nil {
				return nil, errInvalidSinceSecond
			}
			totalSeconds = totalSeconds + second
			items = []string{}
		default:
			items = append(items, string(since[i]))
		}
	}
	if len(items) != 0 {
		return nil, errInvalidSinceSecond
	}
	return &totalSeconds, nil
}
