package commands

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func newMockSubmitSparkJobArgs() *submitSparkJobArgs {
	return &submitSparkJobArgs{
		Image:     "spark-demo:latest",
		MainClass: "com.alibaba.www.main",
		Jar:       "local://spark-demo.jar",
		Executor: &Executor{
			Replicas:              0,
			ExecutorCPURequest:    "1",
			ExecutorMemoryRequest: "200Mi",
		},
		Driver: &Driver{
			DriverCPURequest:    "1",
			DriverMemoryRequest: "200Mi",
		},
	}
}

func TestSubmitSparkJobArgsCheck(t *testing.T) {
	args := newMockSubmitSparkJobArgs()
	assert.EqualError(t, args.isValid(), "WorkersMustMoreThanOne", "Workers should be more than one")
}
