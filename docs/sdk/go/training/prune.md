# Clean All Finished Training Jobs

This API is used to clean job that live longer than relative duration like 5s, 2m, or 3h.

## Path

pkg/apis/arenaclient.TrainingJobClient

## Function

	func (t *TrainingJobClient) Prune(allNamespaces bool, since time.Duration) error

## Parameters

* allNamespaces(type: bool) => if allNamespaces is true,api will return all training jobs of all namespace
* since(type: time.Duration) = > clean job that live longer than relative duration like 5s, 2m, or 3h
  
## Example

	package main
	import(
		"fmt"
        "time"
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
		// Clean up jobs completed 100 seconds ago
		err = client.Training().Prune(true,time.Duration(100 * time.Second))
        if err != nil {
            fmt.Printf("failed to clean training jobs,reason: %v",err)
        }
	}
