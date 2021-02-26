# Delete The Serving Job

This API is used to delete the serving job.

## Path

pkg/apis/arenaclient.ServingJobClient

## Function

	func (t *ServingJobClient) Delete(jobType types.ServingJobType, version string, jobNames ...string) error 

## Parameters

* jobName(type: ...string) => the names of serving jobs
* jobVersion(type: string) => specify the the version of jobs
* jobType(type: pkg/apis/types.ServingJobType) = > specify the serving job type
  
## Example

### delete the serving job 

	package main
	import(
		"fmt"
		"github.com/kubeflow/arena/pkg/apis/arenaclient"
		"github.com/kubeflow/arena/pkg/apis/types"
	)

	func main() {
		// create the arena client
		client, err := arenaclient.NewArenaClient(types.ArenaClientArgs{
			Kubeconfig:     "",
			LogLevel:      	"debug",
			Namespace:      "",
			ArenaNamespace: "",
			IsDaemonMode:   false,
		})
		if err != nil {
			fmt.Printf("failed to build arena client.,reason: %v",err)
			return
		}
		// delete the serving job test
		err = client.Serving().Delete(types.AllServingJob,"alpha","test")
        if err != nil {
            fmt.Printf("failed to delete serving job,reason: %v",err)
        }
	}

### delete serving jobs 

	package main
	import(
		"fmt"
		"github.com/kubeflow/arena/pkg/apis/arenaclient"
		"github.com/kubeflow/arena/pkg/apis/types"
	)

	func main() {
		// create the arena client
		client, err := arenaclient.NewArenaClient(types.ArenaClientArgs{
			Kubeconfig:     "",
			LogLevel:      	"debug",
			Namespace:      "",
			ArenaNamespace: "",
			IsDaemonMode:   false,
		})
		if err != nil {
			fmt.Printf("failed to build arena client.,reason: %v",err)
			return
		}
		// delete the serving jobs: test1,test2,test3
		err = client.Serving().Delete(types.AllServingJob,"test1","test2","test3")
        if err != nil {
            fmt.Printf("failed to delete serving jobs,reason: %v",err)
        }
	}