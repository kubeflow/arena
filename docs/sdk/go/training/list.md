# List Training Jobs

API of Listing Training Jobs can list all training jobs or list the target training jobs under the training job type has been specified.

## Path

pkg/apis/arenaclient.TrainingJobClient

## Function

	func (t *TrainingJobClient) List(allNamespaces bool, trainingType types.TrainingJobType) ([]*types.TrainingJobInfo, error)

## Parameters

* allNamespaces(type: bool) => if allNamespaces is true,api will return all training jobs of all namespaces
* trainingType(type: pkg/apis/types.TrainingJobType) = > specify the training job type
  
## Example

### List all training jobs of all namespaces

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
		// list all training jobs of all namespaces
		trainingJobs,err := client.Training().List(true,types.AllTrainingJob)
        if err != nil {
            fmt.Printf("failed to list training jobs,reason: %v",err)
            return 
        }
        for _,j := range trainingJobs {
            fmt.Printf("job Name: %v\n",j.Name)
        }
	}

### List all training jobs of target training job type

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
		// list all tensorflow training jobs of tfjob-namespace
		trainingJobs,err := client.Training().Namespace("tfjob-namespace").List(true,types.TFTrainingJob)
        if err != nil {
            fmt.Printf("failed to list training jobs,reason: %v",err)
            return 
        }
        for _,j := range trainingJobs {
            fmt.Printf("job Name: %v\n",j.Name)
        }
	}