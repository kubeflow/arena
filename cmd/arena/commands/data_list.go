// Copyright 2018 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package commands

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/kubeflow/arena/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	allNamespaces bool = false
)

// List Data Command
func NewDataListCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "list",
		Short: "list all the data volume.",

		Run: func(cmd *cobra.Command, args []string) {
			util.SetLogLevel(logLevel)

			setupKubeconfig()
			_, err := initKubeClient()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			var pvcList *v1.PersistentVolumeClaimList
			if allNamespaces {
				namespace = metav1.NamespaceAll
			}
			pvcList, err = clientset.CoreV1().PersistentVolumeClaims(namespace).List(metav1.ListOptions{})
			if err != nil {
				log.Debugf("Failed to list data volume due to %v", err)
			}

			displayDataVolume(pvcList)
		},
	}

	command.Flags().BoolVar(&allNamespaces, "allNamespaces", false, "show all the namespaces")

	return command
}

func displayDataVolume(pvcList *v1.PersistentVolumeClaimList) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	if allNamespaces {
		fmt.Fprintf(w, "NAMESPACE\tNAME\tSTATUS\tCAPACITY\tACCESS MODES\tAGE\n")
	} else {
		fmt.Fprintf(w, "NAME\tSTATUS\tCAPACITY\tACCESS MODES\tAGE\n")
	}

	if pvcList == nil {
		return
	}

	for _, item := range pvcList.Items {
		storage := item.Status.Capacity[v1.ResourceStorage]
		capacity := storage.String()
		phase := item.Status.Phase
		if item.ObjectMeta.DeletionTimestamp != nil {
			phase = "Terminating"
		}

		if allNamespaces {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
				item.Namespace,
				item.Name,
				string(phase),
				// item.Spec.VolumeName,
				capacity,
				getAccessModesAsString(item.Spec.AccessModes),
				translateTimestamp(item.CreationTimestamp))
		} else {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				item.Name,
				string(phase),
				// item.Spec.VolumeName,
				capacity,
				getAccessModesAsString(item.Spec.AccessModes),
				translateTimestamp(item.CreationTimestamp))
		}
	}

	_ = w.Flush()

}

func getAccessModesAsString(modes []v1.PersistentVolumeAccessMode) string {
	modes = removeDuplicateAccessModes(modes)
	modesStr := []string{}
	if containsAccessMode(modes, v1.ReadWriteOnce) {
		modesStr = append(modesStr, "ReadWriteOnce")
	}
	if containsAccessMode(modes, v1.ReadOnlyMany) {
		modesStr = append(modesStr, "ReadOnlyMany")
	}
	if containsAccessMode(modes, v1.ReadWriteMany) {
		modesStr = append(modesStr, "ReadWriteMany")
	}
	return strings.Join(modesStr, ",")
}

func removeDuplicateAccessModes(modes []v1.PersistentVolumeAccessMode) []v1.PersistentVolumeAccessMode {
	accessModes := []v1.PersistentVolumeAccessMode{}
	for _, m := range modes {
		if !containsAccessMode(accessModes, m) {
			accessModes = append(accessModes, m)
		}
	}
	return accessModes
}

func containsAccessMode(modes []v1.PersistentVolumeAccessMode, mode v1.PersistentVolumeAccessMode) bool {
	for _, m := range modes {
		if m == mode {
			return true
		}
	}
	return false
}

// translateTimestamp returns the elapsed time since timestamp in
// human-readable approximation.
func translateTimestamp(timestamp metav1.Time) string {
	if timestamp.IsZero() {
		return "<unknown>"
	}
	return util.ShortHumanDuration(time.Now().Sub(timestamp.Time))
}
