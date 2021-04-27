package cron

import (
	"context"
	"encoding/json"
	"github.com/kubeflow/arena/pkg/apis/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types2 "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
)

func SuspendCron(name, namespace string, suspend bool) error {
	config := config.GetArenaConfiger().GetRestConfig()

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return err
	}

	patchData := make(map[string]interface{})
	patchData["suspend"] = suspend

	specData := make(map[string]interface{})
	specData["spec"] = patchData

	payloadBytes, _ := json.Marshal(specData)

	_, err = dynamicClient.Resource(gvr).Namespace(namespace).
		Patch(context.TODO(), name, types2.MergePatchType, payloadBytes, metav1.PatchOptions{})

	return err
}
