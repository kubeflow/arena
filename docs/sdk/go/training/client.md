# TrainingJobClient

TrainingJobClient includes some apis for managing training jobs.

## Path

pkg/apis/arenaclient.TrainingJobClient

## Function

	func (a *ArenaClient) Training() *TrainingJobClient

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
		// training client should be created by arena client
		trainingClient := client.Training()
		fmt.Println(trainingClient)
	}
