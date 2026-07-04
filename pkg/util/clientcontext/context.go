// Copyright 2026 The Kubeflow Authors
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

package clientcontext

import (
	"bytes"
	"runtime"
	"strconv"
	"sync"
)

var (
	activeConfigers = make(map[uint64]interface{})
	activeAccessers = make(map[uint64]interface{})
	contextMu       sync.RWMutex

	// Registries mapping configer interface{} to resources
	configerAccessers  = make(map[interface{}]interface{})
	configerTrainers   = make(map[interface{}]interface{})
	configerProcessers = make(map[interface{}]interface{})
	configerPrometheus = make(map[interface{}]interface{})
	registryMu         sync.RWMutex
)

func getGID() uint64 {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	b := buf[:n]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	i := bytes.IndexByte(b, ' ')
	if i < 0 {
		return 0
	}
	id, _ := strconv.ParseUint(string(b[:i]), 10, 64)
	return id
}

// AssociateConfigerAndAccesser links a configer to its accesser
func AssociateConfigerAndAccesser(configer, accesser interface{}) {
	registryMu.Lock()
	defer registryMu.Unlock()
	configerAccessers[configer] = accesser
}

// GetAssociatedAccesser gets the accesser linked to a configer
func GetAssociatedAccesser(configer interface{}) interface{} {
	registryMu.RLock()
	defer registryMu.RUnlock()
	return configerAccessers[configer]
}

// SetConfigerTrainers caches trainers for a configer
func SetConfigerTrainers(configer, trainers interface{}) {
	registryMu.Lock()
	defer registryMu.Unlock()
	configerTrainers[configer] = trainers
}

// GetConfigerTrainers retrieves trainers cached for a configer
func GetConfigerTrainers(configer interface{}) interface{} {
	registryMu.RLock()
	defer registryMu.RUnlock()
	return configerTrainers[configer]
}

// SetConfigerProcessers caches serving processers for a configer
func SetConfigerProcessers(configer, processers interface{}) {
	registryMu.Lock()
	defer registryMu.Unlock()
	configerProcessers[configer] = processers
}

// GetConfigerProcessers retrieves processers cached for a configer
func GetConfigerProcessers(configer interface{}) interface{} {
	registryMu.RLock()
	defer registryMu.RUnlock()
	return configerProcessers[configer]
}

// SetConfigerPrometheus caches prometheus client for a configer
func SetConfigerPrometheus(configer, promClient interface{}) {
	registryMu.Lock()
	defer registryMu.Unlock()
	configerPrometheus[configer] = promClient
}

// GetConfigerPrometheus retrieves prometheus client cached for a configer
func GetConfigerPrometheus(configer interface{}) interface{} {
	registryMu.RLock()
	defer registryMu.RUnlock()
	return configerPrometheus[configer]
}

// SetCurrentContext sets the active configer and accesser for the current goroutine
func SetCurrentContext(configer, accesser interface{}) {
	gid := getGID()
	if gid == 0 {
		return
	}
	contextMu.Lock()
	defer contextMu.Unlock()
	activeConfigers[gid] = configer
	activeAccessers[gid] = accesser
}

// ClearCurrentContext clears context for the current goroutine
func ClearCurrentContext() {
	gid := getGID()
	if gid == 0 {
		return
	}
	contextMu.Lock()
	defer contextMu.Unlock()
	delete(activeConfigers, gid)
	delete(activeAccessers, gid)
}

// GetActiveConfiger gets the active configer for the current goroutine
func GetActiveConfiger() interface{} {
	gid := getGID()
	if gid == 0 {
		return nil
	}
	contextMu.RLock()
	defer contextMu.RUnlock()
	return activeConfigers[gid]
}

// GetActiveAccesser gets the active accesser for the current goroutine
func GetActiveAccesser() interface{} {
	gid := getGID()
	if gid == 0 {
		return nil
	}
	contextMu.RLock()
	defer contextMu.RUnlock()
	return activeAccessers[gid]
}
