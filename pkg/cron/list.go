package cron

import (
	"encoding/json"
	"fmt"
	"github.com/kubeflow/arena/pkg/apis/config"
	"github.com/kubeflow/arena/pkg/apis/types"
	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"os"
	"strconv"
	"text/tabwriter"
)

func ListCrons(namespace string, allNamespaces bool) ([]*types.CronInfo, error) {
	if allNamespaces {
		namespace = metav1.NamespaceAll
	}

	config := config.GetArenaConfiger().GetRestConfig()

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	list, err := dynamicClient.Resource(gvr).Namespace(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var cronInfos []*types.CronInfo

	for _, item := range list.Items {
		b, err := item.MarshalJSON()
		if err != nil {
			continue
		}

		c, err := buildCronInfo(b)
		if err != nil {
			continue
		}

		cronInfos = append(cronInfos, c)
	}
	return cronInfos, nil
}

func DisplayAllCrons(crons []*types.CronInfo, allNamespaces bool, format types.FormatStyle) {
	switch format {
	case "json":
		for _, cron := range crons {
			data, _ := json.MarshalIndent(cron, "", "    ")
			fmt.Printf("%v", string(data))
		}
		return
	case "yaml":
		for _, cron := range crons {
			data, _ := yaml.Marshal(cron)
			fmt.Printf("%v", string(data))
		}
		return
	case "", "wide":
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		var header []string
		if allNamespaces {
			header = append(header, "NAMESPACE")
		}
		header = append(header, []string{"NAME", "TYPE", "SCHEDULE", "SUSPEND", "DEADLINE", "CONCURRENCYPOLICY"}...)
		printLine(w, header...)

		for _, cron := range crons {
			var items []string
			if allNamespaces {
				items = append(items, cron.Namespace)
			}

			items = append(items, []string{
				cron.Name,
				cron.Type,
				cron.Schedule,
				strconv.FormatBool(cron.Suspend),
				cron.Deadline,
				cron.ConcurrencyPolicy,
			}...)
			printLine(w, items...)
		}
		_ = w.Flush()
		return
	}
}
