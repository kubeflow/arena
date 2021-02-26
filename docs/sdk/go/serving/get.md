# Get The Job Details

API of Getinng Serving Job Details can get the serving job details.

## Path

pkg/apis/arenaclient.ServingJobClient

## Function

	func (t *ServingJobClient) Get(jobName, version string, jobType types.ServingJobType) (*types.ServingJobInfo, error)

## Parameters

* jobName(type: string) => the name of serving job
* version(type string) => the version of serving job,if the version is null(""),match the all versions of the job.
* jobType(type: pkg/apis/types.ServingJobType) = > specify the serving job type
  
## Example

### serving job type is unknown

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
		// get the serving job which name is test and its' type is unknown
		servingJob,err := client.Serving().Get("test","",types.AllServingJob)
        if err != nil {
            fmt.Printf("failed to get serving job,reason: %v",err)
            return 
        }
        fmt.Printf("serving job instances: %v\n",servingJob.Instances)
	}

### serving job type is known

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
		// get the serving job which name is test and its' type is custom serving job
		servingJob,err := client.Serving().Get("test","",types.CustomServingJob))
        if err != nil {
            fmt.Printf("failed to get serving job,reason: %v",err)
            return 
        }
        fmt.Printf("serving job Instances: %v\n",servingJob.Instances)
	}