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
	"strconv"
	"strings"
	"time"
)

const (
	timeLayout = "2006-01-02 15:04:05"
)

var unitFactor = map[string]int64{
	"s": 1,
	"m": 60,
	"h": 3600,
	"d": 24 * 3600,
	"y": 3600 * 24 * 365,
}

// ShortHumanDuration returns a succinct representation of the provided duration
// with limited precision for consumption by humans.
func ShortHumanDuration(d time.Duration) string {
	// Allow deviation no more than 2 seconds(excluded) to tolerate machine time
	// inconsistence, it can be considered as almost now.
	if seconds := int(d.Seconds()); seconds < -1 {
		return "<invalid>"
	} else if seconds < 0 {
		return "0s"
	} else if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	} else if minutes := int(d.Minutes()); minutes < 60 {
		return fmt.Sprintf("%dm", minutes)
	} else if hours := int(d.Hours()); hours < 24 {
		return fmt.Sprintf("%dh", hours)
	} else if hours < 24*365 {
		return fmt.Sprintf("%dd", hours/24)
	}
	return fmt.Sprintf("%dy", int(d.Hours()/24/365))
}

func GetFormatTime(timestamp int64) string {
	return time.Unix(timestamp, 0).Format(timeLayout)
}

func TransTimeStrToSeconds(duration string) (int64, error) {
	var unit string
	for k := range unitFactor {
		if strings.HasSuffix(duration, k) {
			unit = k
			break
		}
	}

	if unit == "" {
		return 0, fmt.Errorf("invalid duration unit in %s", duration)
	}

	durationValue := strings.TrimSuffix(duration, unit)
	durationSeconds, err := strconv.ParseInt(durationValue, 10, 64)
	if err != nil {
		return 0, err
	}

	return durationSeconds * unitFactor[unit], nil
}
