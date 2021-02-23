# Build TensorRT Serving Job

This section introduces how to customly build a TensorRT serving job.

## Path

    pkg/apis/serving.TRTServingJobBuilder

## Function

    func NewTRTServingJobBuilder() *TRTServingJobBuilder 

## Parameters

TRTServingJobBuilder has following functions to custom your TensorRT serving job.

| function  |  description  | matches cli option |
|:---|:--:|:---|
|Name(name string) *TRTServingJobBuilder|specify the job name|--name|
| Namespace(namespace string) *TRTServingJobBuilder|specify the namespace|--namespace/-n|
|Command(args []string) *TRTServingJobBuilder|specify the command|-|
| GPUCount(count int) *TRTServingJobBuilder|specify the gpu count|--gpus|
| GPUMemory(memory int) *TRTServingJobBuilder |specify the gpu memory(gpushare)| --gpumemory|
| Image(image string) *TRTServingJobBuilder|specify the image|--image|
| ImagePullPolicy(policy string) *TRTServingJobBuilder|specify the image pull policy|--image-pull-policy|
| CPU(cpu string) *TRTServingJobBuilder | specify the cpu limitation|--cpu|
|Memory(memory string) *TRTServingJobBuilder |specify the memory limitation|--memory|
|Envs(envs map[string]string) *TRTServingJobBuilder | specify the envs of containers| --env |
| Replicas(count int) *TRTServingJobBuilder|specify the replicas| --replicas|
| EnableIstio() *TRTServingJobBuilder|enable istio|--enable-istio|
| ExposeService() *TRTServingJobBuilder|expose service|--expose-service|
| Version(version string) *TRTServingJobBuilder| specify the version|--version|
| Tolerations(tolerations []string) *TRTServingJobBuilder|specify the node taint tolerations| --toleration|
| NodeSelectors(selectors map[string]string) *TRTServingJobBuilder|specify the node selectors|--selector|
|Annotations(annotations map[string]string) *TRTServingJobBuilder |specify the annotation|--annotation|
|Datas(volumes map[string]string) *TRTServingJobBuilder|specify the pvc which stores dataset|--data|
| DataDirs(volumes map[string]string) *TRTServingJobBuilder|specify the host path which stores dataset|--data-dir|
| HttpPort(port int) *TRTServingJobBuilder|specify the http service port|--http-port|
|GrpcPort(port int) *TRTServingJobBuilder|specify the grpc service port|--grpc-port|
| MetricsPort(port int) *TRTServingJobBuilder|specify the metric port|--metric-port|
| ModelStore(store string) *TRTServingJobBuilder |specify the path of storing model| --model-store|
|AllowMetrics() *TRTServingJobBuilder |enable metrics| --allow-metrics|
|Build() (*Job, error) |build the job|-|