package commands

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/kubeflow/arena/util/helm"
)

func NewTopJobCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "job",
		Short: "Display Resource (GPU) usage of jobs.",
		Run: func(cmd *cobra.Command, args []string) {
			setupKubeconfig()
			client, err := initKubeClient()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			releaseMap, err := helm.ListReleaseMap()
			// log.Printf("releaseMap %v", releaseMap)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			// allPods, err := acquireAllActivePods(client)
			// if err != nil {
			// 	fmt.Println(err)
			// 	os.Exit(1)
			// }

			allPods, err = acquireAllPods(client)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			allJobs, err = acquireAllJobs(client)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			trainers := NewTrainers(client)
			jobs := []TrainingJob{}

			for name, ns := range releaseMap {
				supportedChart := false
				for _, trainer := range trainers {
					if trainer.IsSupported(name, ns) {
						job, err := trainer.GetTrainingJob(name, ns)
						if err != nil {
							fmt.Println(err)
							os.Exit(1)
						}
						jobs = append(jobs, job)
						supportedChart = true
						break
					}
				}

				if !supportedChart {
					log.Debugf("Unkown chart %s\n", name)
				}

			}

			jobs = makeTrainingJobOrderdByGPUCount(jobs)

			// TODO(cheyang): Support different job describer, such as MPI job/tf job describer
			displayTrainingJobList(jobs, true)

		},
	}

	// command.Flags().BoolVarP(&showDetails, "details", "d", false, "Display details")
	return command
}
