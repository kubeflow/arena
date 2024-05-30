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

package cron

import (
	"fmt"
	"io"
	"strings"
	"time"
)

const (
	formatLayout = "2006-01-02T15:04:05Z"
)

func printLine(w io.Writer, fields ...string) {
	buffer := strings.Join(fields, "\t")
	fmt.Fprintln(w, buffer)
}

func formatTime(t time.Time) string {
	return t.Format(formatLayout)
}
