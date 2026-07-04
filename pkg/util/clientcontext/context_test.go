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
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

type dummyConfiger struct {
	name string
}

type dummyAccessor struct {
	id string
}

func TestGoroutineLocalContext(t *testing.T) {
	configer1 := &dummyConfiger{name: "configer-1"}
	accessor1 := &dummyAccessor{id: "accessor-1"}

	configer2 := &dummyConfiger{name: "configer-2"}
	accessor2 := &dummyAccessor{id: "accessor-2"}

	var wg sync.WaitGroup
	wg.Add(2)

	// Goroutine 1
	go func() {
		defer wg.Done()
		SetCurrentContext(configer1, accessor1)
		defer ClearCurrentContext()

		// Verify this goroutine gets configer1 and accessor1
		assert.Equal(t, configer1, GetActiveConfiger())
		assert.Equal(t, accessor1, GetActiveAccesser())
	}()

	// Goroutine 2
	go func() {
		defer wg.Done()
		SetCurrentContext(configer2, accessor2)
		defer ClearCurrentContext()

		// Verify this goroutine gets configer2 and accessor2
		assert.Equal(t, configer2, GetActiveConfiger())
		assert.Equal(t, accessor2, GetActiveAccesser())
	}()

	wg.Wait()

	// Verify main goroutine has no active context
	assert.Nil(t, GetActiveConfiger())
	assert.Nil(t, GetActiveAccesser())
}

func TestConfigerRegistryAssociationAndCaching(t *testing.T) {
	configer1 := &dummyConfiger{name: "configer-1"}
	accessor1 := &dummyAccessor{id: "accessor-1"}
	trainers1 := map[string]string{"tf": "trainer-1"}
	processers1 := map[string]string{"serving": "processer-1"}
	prometheus1 := "prom-1"

	configer2 := &dummyConfiger{name: "configer-2"}
	accessor2 := &dummyAccessor{id: "accessor-2"}
	trainers2 := map[string]string{"tf": "trainer-2"}
	processers2 := map[string]string{"serving": "processer-2"}
	prometheus2 := "prom-2"

	// Associate
	AssociateConfigerAndAccesser(configer1, accessor1)
	AssociateConfigerAndAccesser(configer2, accessor2)

	assert.Equal(t, accessor1, GetAssociatedAccesser(configer1))
	assert.Equal(t, accessor2, GetAssociatedAccesser(configer2))

	// Cache trainers
	SetConfigerTrainers(configer1, trainers1)
	SetConfigerTrainers(configer2, trainers2)

	assert.Equal(t, trainers1, GetConfigerTrainers(configer1))
	assert.Equal(t, trainers2, GetConfigerTrainers(configer2))

	// Cache processers
	SetConfigerProcessers(configer1, processers1)
	SetConfigerProcessers(configer2, processers2)

	assert.Equal(t, processers1, GetConfigerProcessers(configer1))
	assert.Equal(t, processers2, GetConfigerProcessers(configer2))

	// Cache prometheus
	SetConfigerPrometheus(configer1, prometheus1)
	SetConfigerPrometheus(configer2, prometheus2)

	assert.Equal(t, prometheus1, GetConfigerPrometheus(configer1))
	assert.Equal(t, prometheus2, GetConfigerPrometheus(configer2))
}
