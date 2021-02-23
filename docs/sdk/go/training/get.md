# Get The Job Details

API of Getinng Training Job Details can get the training job details.

## Path

pkg/apis/arenaclient.TrainingJobClient

## Function

	func (t *TrainingJobClient) Get(jobName string, jobType types.TrainingJobType) (*types.TrainingJobInfo, error)

## Parameters

* jobName(type: string) => the name of training job
* jobType(type: pkg/apis/types.TrainingJobType) = > specify the training job type
  
## Example

### Get the training job details(training job type is unknown)

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
		// get the training job which name is test and its' type is unknown
		trainingJob,err := client.Training().Get("test",types.AllTrainingJob)
        if err != nil {
            fmt.Printf("failed to get training job,reason: %v",err)
            return 
        }
        fmt.Printf("training job status: %v\n",trainingJob.Status)
	}

### Get the training job details(training job type is known)

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
		// get the training job which name is test and its' type is tensorflow job
		trainingJob,err := client.Training().Get("test",types.TFTrainingJob))
        if err != nil {
            fmt.Printf("failed to get training job,reason: %v",err)
            return 
        }
        fmt.Printf("training job status: %v\n",trainingJob.Status)
	}