# ArenaClient

ArenaClient is the entry point of all APIs and it can create other sub client such as TrainingClient or ServingClient.

## Path

pkg/apis/arenaclient.ArenaClient

## Function

    func NewArenaClient(args types.ArenaClientArgs) (*ArenaClient, error)

## Parameters

The types.ArenaClientArgs is defined at ``pkg/apis/types.ArenaClientArgs`` and is described as below:

	type ArenaClientArgs struct {
		// Kubeconfig is used to specify the kubeconfig file,for example: "~/.kubeconfig",
		// if you use service account,please set the kubeconfig as "".
		Kubeconfig     string
		// Namespace is used to set the namespace,if the namespace is "",the default namespace will be found from
		// file ~/.arena/config (set like: namespace = YOUR_DEFAULT_NAMESPACE) or kubeconfig file.
		Namespace      string
		// ArenaNamespace is used to set the arena namespace, the operators of arena depends
		// are installed in this namespace.
		ArenaNamespace string
		// if IsDaemonMode is false, the arena will directly access k8s resources from k8s api server
		// i IsDaemonMode is true,the arena will access k8s resources from cache, the cache is based on
		// client-go event mechanism
		IsDaemonMode   bool
		// LogLevel is used to set log level.(debug,info)
		LogLevel       string
	}

## Example

	package main
	import(
		"fmt"
		"github.com/kubeflow/arena/pkg/apis/arenaclient"
		"github.com/kubeflow/arena/pkg/apis/types"
	)

	func main() {
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
	}