# Query Cluster Nodes' Details

This API is used to query the cluster nodes' details.

## Path

pkg/apis/arenaclient.NodeClient

## Function

	func (t *NodeClient) Details(nodeNames []string, nodeType types.NodeType, showMetric bool) (types.AllNodeInfo, error) 

## Parameters

* nodeNames(type: []string) => the node names that you want to query, if the array length is 0, the api will list all nodes.
* nodeType(type: pkg/apis/types.NodeType) => specify the node type.
* showMetric(type: bool) => if true,the arena will get gpu metrics from prometheus.
  
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
		nodeInfos,err := client.Node().Details([]string{},types.AllKnownNode,false)
        if err != nil {
            fmt.Printf("failed to get node details,reason: %v",err)
            return 
        }
        for key, objs := range nodeInfos {
            for _, obj := range objs {
                switch key {
                case "gpuExclusiveNodes":
                    node := obj.(types.GPUExclusiveNodeInfo)
                    fmt.Printf("node name: %v\n",node.Name)
                }
            }
        }
	}

