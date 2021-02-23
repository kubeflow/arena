# ServingJobClient

ServingJobClient includes some apis for managing training jobs.

## Path

pkg/apis/arenaclient.ServingJobClient

## Function

	func (a *ArenaClient) Serving() *ServingJobClient

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
		fmt.Println(client)
		// serving client should be created by arena client
		servingClient := client.Serving()
		fmt.Println(servingClient)
	}
