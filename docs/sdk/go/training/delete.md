# Delete The Training Job

This API is used to delete the training job.

## Path

pkg/apis/arenaclient.TrainingJobClient

## Function

	func (t *TrainingJobClient) Delete(jobType types.TrainingJobType, jobNames ...string) error

## Parameters

* jobName(type: ...string) => the names of training jobs
* jobType(type: pkg/apis/types.TrainingJobType) = > specify the training job type
  
## Example

### delete the training job 

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
		// delete the training job test
		err = client.Training().Delete(types.AllTrainingJob,"test")
        if err != nil {
            fmt.Printf("failed to delete training job,reason: %v",err)
        }
	}

### delete training jobs 

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
		// delete the training jobs: test1,test2,test3
		err = client.Training().Delete(types.AllTrainingJob,"test1","test2","test3")
        if err != nil {
            fmt.Printf("failed to delete training jobs,reason: %v",err)
        }
	}