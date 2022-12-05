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
	"time"
)

const (
	timeLayout = "2006-01-02 15:04:05"
)

// ShortHumanDuration returns a succint representation of the provided duration
// with limited precision for consumption by humans.
func ShortHumanDuration(d time.Duration) string {
	// Allow deviation no more than 2 seconds(excluded) to tolerate machine time
	// inconsistence, it can be considered as almost now.
	if seconds := int(d.Seconds()); seconds < -1 {
		return fmt.Sprintf("<invalid>")
	} else if seconds <= 0 {
		return fmt.Sprintf("0s")
	} else if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	} else if minutes := int(d.Minutes()); minutes < 60 {
		return fmt.Sprintf("%dm%ds", minutes, seconds%60)
	} else if hours := int(d.Hours()); hours < 24 {
		minutes = (seconds - hours*3600) / 60
		return fmt.Sprintf("%dh%dm%ds", hours, minutes, seconds%60)
	} else if hours < 24*365 {
		minutes = (seconds - hours*3600) / 60
		return fmt.Sprintf("%dd%dh%dm%ds", hours/24, hours%24, minutes, seconds%60)
	}
	return fmt.Sprintf("%dy", int(d.Hours()/24/365))
}

func GetFormatTime(timestamp int64) string {
	return time.Unix(timestamp, 0).Format(timeLayout)
}
