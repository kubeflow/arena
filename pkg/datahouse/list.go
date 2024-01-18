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

package datahouse

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/kubeflow/arena/pkg/apis/config"
	"github.com/kubeflow/arena/pkg/util"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	dataDescritpion = "description"
	dataOwner       = "owner"
)

func DisplayDataVolumes(namespace string, allNamespaces bool) error {
	client := config.GetArenaConfiger().GetClientSet()
	var pvcList *v1.PersistentVolumeClaimList
	if allNamespaces {
		namespace = metav1.NamespaceAll
	}
	pvcList, err := client.CoreV1().PersistentVolumeClaims(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("Failed to list data volume due to %v", err)
	}
	displayDataVolume(pvcList, allNamespaces)
	return nil
}

// Display the data volume
func displayDataVolume(pvcList *v1.PersistentVolumeClaimList, allNamespaces bool) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	if allNamespaces {
		fmt.Fprintf(w, "NAME\tNAMESPACE\tACCESSMODE\tDESCRIPTION\tOWNER\tAGE\n")
	} else {
		fmt.Fprintf(w, "NAME\tACCESSMODE\tDESCRIPTION\tOWNER\tAGE\n")
	}
	if pvcList == nil {
		return
	}
	for _, item := range pvcList.Items {
		if item.ObjectMeta.DeletionTimestamp != nil || item.Status.Phase != v1.ClaimBound {
			log.Debugf("skip the data is not ready: %s", item.Name)
			continue
		}

		// storage := item.Status.Capacity[v1.ResourceStorage]
		// capacity := storage.String()
		description := item.Annotations[dataDescritpion]
		owner := item.Annotations[dataOwner]

		if allNamespaces {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
				item.Name,
				item.Namespace,
				getAccessModesAsString(item.Spec.AccessModes),
				description,
				owner,
				translateTimestamp(item.CreationTimestamp))
		} else {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				item.Name,
				getAccessModesAsString(item.Spec.AccessModes),
				description,
				owner,
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
	return util.ShortHumanDuration(time.Since(timestamp.Time))
}
