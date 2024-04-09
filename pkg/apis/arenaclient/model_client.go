package arenaclient

import (
	"errors"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kubeflow/arena/pkg/apis/config"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/k8saccesser"
	"github.com/kubeflow/arena/pkg/model"
)

type ModelClient struct {
	namespace string
	configr   *config.ArenaConfiger

	model.MlflowClient
}

func NewModelClient(namespace string, configer *config.ArenaConfiger) (*ModelClient, error) {
	trackingUri := os.Getenv("MLFLOW_TRACKING_URI")
	username := os.Getenv("MLFLOW_TRACKING_USERNAME")
	password := os.Getenv("MLFLOW_TRACKING_PASSWORD")

	var mlflowClient *model.MlflowClient
	if trackingUri != "" {
		// Construct a non-proxied MLflow client if `MLFLOW_TRACKING_URI` is specified
		mlflowClient = model.NewMlflowClient(trackingUri, username, password)
	} else {
		// Construct a MLflow client proxied by api server
		mlflowServices, err := listMlflowServices()
		if err != nil {
			return nil, fmt.Errorf("failed to create proxied model client: %v", err)
		}
		if len(mlflowServices) == 0 {
			return nil, fmt.Errorf("failed to create proxied model client: no mlflow service in any namespace found")
		}
		mlflowService := mlflowServices[0].DeepCopy()
		if len(mlflowServices) > 1 {
			log.Warnf("there are multiple mlflow services found, use %s/%s", mlflowService.ObjectMeta.Namespace, mlflowService.ObjectMeta.Name)
		}
		mlflowClient = model.NewProxiedMlflowClient(configer, mlflowService, username, password)
	}

	modelClient := &ModelClient{
		namespace:    namespace,
		configr:      configer,
		MlflowClient: *mlflowClient,
	}

	health, err := modelClient.CheckHealth()
	if err != nil {
		return nil, fmt.Errorf("failed to create model client: %v", err)
	}
	if !health {
		return nil, errors.New("failed to create model client: mlflow tracking server is not healthy")
	}
	return modelClient, nil
}

func searchModelVersionByJobLabels(namespace string, configer *config.ArenaConfiger, labels map[string]string) *types.ModelVersion {
	var mv *types.ModelVersion
	modelName := labels["modelName"]
	modelVersion := labels["modelVersion"]
	if modelName != "" && modelVersion != "" {
		modelClient, err := NewModelClient(namespace, configer)
		if err != nil {
			log.Warnf("failed to search model version by job labels: %v", err)
		}
		mv, err = modelClient.GetModelVersion(modelName, modelVersion)
		if err != nil {
			log.Warnf("%v", err)
		}
	}
	if mv == nil {
		mv = &types.ModelVersion{
			Name:    modelName,
			Version: modelVersion,
		}
	}
	return mv
}

func listMlflowServices() ([]*corev1.Service, error) {
	services, err := k8saccesser.GetK8sResourceAccesser().ListServices(metav1.NamespaceAll, "app.kubernetes.io/name in (ack-mlflow, mlflow)")
	if err != nil {
		return services, fmt.Errorf("failed to list mlflow service: %v", err)
	}
	return services, nil
}
