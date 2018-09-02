package commands

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/kubeflow/arena/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/kubeflow/arena/util/helm"
)

var serving_charts = map[string]bool{
	"tensorflow-serving-0.2.0": true,
}

func NewServingListCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "list",
		Short: "list all the serving jobs",
		Run: func(cmd *cobra.Command, args []string) {
			util.SetLogLevel(logLevel)

			setupKubeconfig()
			_, err := initKubeClient()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			releaseMap, err := helm.ListAllReleasesWithDetail()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

			fmt.Fprintf(w, "NAME\tSTATUS\tVERSION\tCHART\tNAMESPACE\n")

			for name, cols := range releaseMap {
				log.Debugf("name: %s, cols: %s", name, cols)
				namespace := cols[len(cols)-1]
				chart := cols[len(cols)-2]
				status := cols[len(cols)-3]
				log.Debugf("namespace: %s, chart: %s, status:%s", namespace, chart, status)
				if serving_charts[chart] {
					index := strings.Index(name, "-")
					//serviceName := name[0:index]
					serviceVersion := ""
					if index > -1 {
						serviceVersion = name[index+1:]
					}
					nameAndVersion := strings.Split(name, "-")
					log.Debugf("nameAndVersion: %s, len(nameAndVersion): %s", nameAndVersion, len(nameAndVersion))
					fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", name,
						status,
						serviceVersion,
						chart, namespace)

				}
			}

			_ = w.Flush()
		},
	}

	return command
}
