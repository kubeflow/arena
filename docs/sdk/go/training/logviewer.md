# Get The Job LogViewer

This API is used to get the training job logviewer.

## Path

pkg/apis/arenaclient.TrainingJobClient

## Function

	func (t *TrainingJobClient) LogViewer(jobName string, jobType types.TrainingJobType) ([]string, error)

## Parameters

* jobName(type: string) => the name of training job
* jobType(type: pkg/apis/types.TrainingJobType) = > specify the training job type
  
## Example

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
		urls,err := client.Training().LogViewer("test",types.AllTrainingJob)
        if err != nil {
            fmt.Printf("failed to get training job logviewer,reason: %v",err)
            return 
        }
        fmt.Printf("urls: %v\n",urls)
	}