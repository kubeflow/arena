# List Serving Jobs

API of Listing Serving Jobs can list all serving jobs or list the target serving jobs under the serving job type(or job version) has been specified.

## Path

pkg/apis/arenaclient.ServingJobClient

## Function

	func (t *ServingJobClient) List(allNamespaces bool, servingType types.ServingJobType) ([]*types.ServingJobInfo, error) 

## Parameters

* allNamespaces(type: bool) => if allNamespaces is true,api will return all serving jobs of all namespaces
* servingType(type: pkg/apis/types.ServingJobType) = > specify the serving job type
  
## Example

### List all serving jobs of all namespaces

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
		// list all serving jobs of all namespaces
		servingJobs,err := client.Serving().List(true,types.AllServingJob)
        if err != nil {
            fmt.Printf("failed to list serving jobs,reason: %v",err)
            return 
        }
        for _,j := range servingJobs {
            fmt.Printf("job Name: %v\n",j.Name)
        }
	}

### List all serving jobs of target serving job type

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
		// list all custom serving jobs of test1 namespace
		servingJobs,err := client.Serving().Namespace("test1").List(true,types.CustomServingJob)
        if err != nil {
            fmt.Printf("failed to list serving jobs,reason: %v",err)
            return 
        }
        for _,j := range servingJobs {
            fmt.Printf("job Name: %v\n",j.Name)
        }
	}