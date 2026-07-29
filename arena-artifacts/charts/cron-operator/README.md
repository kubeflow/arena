# cron-operator

![Version: 0.2.1](https://img.shields.io/badge/Version-0.2.1-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 0.2.1](https://img.shields.io/badge/AppVersion-0.2.1-informational?style=flat-square)

A Kubernetes operator that enables cron-based scheduling for machine learning training workloads using standard cron expressions.

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| nameOverride | string | `""` | String to partially override release name. |
| fullnameOverride | string | `""` | String to fully override release name. |
| image.registry | string | `"registry-cn-beijing.ack.aliyuncs.com"` | Image registry. |
| image.repository | string | `"acs/cron-operator"` | Image repository. |
| image.tag | string | `""` | Image tag. |
| image.pullPolicy | string | `"IfNotPresent"` | Image pull policy. |
| image.pullSecrets | list | `[]` | Image pull secrets. |
| replicas | int | `1` | Number of replicas. |
| logEncoder | string | `"console"` | Configure the encoder of logging, can be one of `console` or `json`. |
| logLevel | string | `"info"` | Configure the verbosity of logging, can be one of `debug`, `info`, `error`. |
| leaderElection.enable | bool | `true` | Whether to enable leader election. |
| maxConcurrentReconciles | int | `10` | Maximum number of concurrent reconciles. |
| qps | int | `30` | Maximum QPS to the Kubernetes API server from this client. |
| burst | int | `50` | Maximum burst for throttle. |
| useHostTimezone | bool | `false` | Whether to use host timezone in the container. |
| resources | object | `{"limits":{"cpu":"400m","memory":"512Mi"},"requests":{"cpu":"100m","memory":"128Mi"}}` | Container resources. |
| securityContext | object | `{}` | Container security context. |
| nodeSelector | object | `{}` | Pod node selector. |
| affinity | object | `{}` | Pod affinity. |
| tolerations | list | `[]` | Pod tolerations. |
| podSecurityContext | object | `{}` | Pod security context. |
| service.type | string | `"ClusterIP"` | Service type. |

