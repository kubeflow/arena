{{- /*
Copyright 2025 The Kubeflow authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/ -}}

{{- /* Expand the name of the chart. */ -}}
{{- define "tf-operator.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- /*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/ -}}
{{- define "tf-operator.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{- /* Create chart name and version as used by the chart label. */ -}}
{{- define "tf-operator.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- /* Common labels */ -}}
{{- define "tf-operator.labels" -}}
helm.sh/chart: {{ include "tf-operator.chart" . }}
{{ include "tf-operator.selectorLabels" . }}
{{- with .Chart.AppVersion }}
app.kubernetes.io/version: {{ . | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- with .Values.labels }}
{{- . | toYaml | nindent 0 }}
{{- end }}
{{- end }}

{{- /* Selector labels. */ -}}
{{- define "tf-operator.selectorLabels" -}}
app.kubernetes.io/name: {{ include "tf-operator.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{- /* Common annotations. */ -}}
{{- define "tf-operator.annotations" -}}
{{- with .Values.annotations }}
{{- . | toYaml | nindent 0 }}
{{- end }}
{{- end -}}

{{- /* Create the name of the service account to use. */ -}}
{{- define "tf-operator.serviceAccount.name" -}}
{{- include "tf-operator.fullname" . }}
{{- end -}}

{{- /* Create the name of the cluster role to use. */ -}}
{{- define "tf-operator.clusterRole.name" -}}
{{- include "tf-operator.fullname" . }}
{{- end -}}

{{- /* Create the name of the cluster role binding to use. */ -}}
{{- define "tf-operator.clusterRoleBinding.name" -}}
{{- include "tf-operator.fullname" . }}
{{- end -}}

{{- /* Create the name of the configmap to use. */ -}}
{{- define "tf-operator.configMap.name" -}}
{{- include "tf-operator.fullname" . }}
{{- end -}}

{{- /* Create the name of the deployment to use. */ -}}
{{- define "tf-operator.deployment.name" -}}
{{- include "tf-operator.fullname" . }}
{{- end -}}

{{- /* Create the name of the service to use. */ -}}
{{- define "tf-operator.service.name" -}}
{{- include "tf-operator.fullname" . }}
{{- end -}}

{{- /* Create the name of the image to use. */ -}}
{{- define "tf-operator.image" -}}
{{- $imageRegistry := .Values.image.registry | default .Values.global.imagePrefix | default .Values.global.image.registry }}
{{- $imageRepository := .Values.image.repository }}
{{- $imageTag := .Values.image.tag | default .Chart.AppVersion }}
{{- printf "%s/%s:%s" $imageRegistry $imageRepository $imageTag }}
{{- end -}}

{{- /* Create the nodeSelector of tf-operator pods to use. */ -}}
{{- define "tf-operator.nodeSelector" -}}
{{- with .Values.global.nodeSelector }}
{{- . | toYaml | nindent 0 }}
{{- end }}
{{- with .Values.nodeSelector }}
{{- . | toYaml | nindent 0 }}
{{- end }}
{{- if eq .Values.global.clusterProfile "Edge" }}
alibabacloud.com/is-edge-worker: "false"
{{- end }}
{{- end -}}

{{- /* Create the affinity of tf-operator pods to use. */ -}}
{{- define "tf-operator.affinity" -}}
podAntiAffinity:
  preferredDuringSchedulingIgnoredDuringExecution:
  - weight: 100
    podAffinityTerm:
      labelSelector:
        matchLabels:
          {{- include "tf-operator.selectorLabels" . | nindent 10 }}
      topologyKey: kubernetes.io/hostname
{{- end -}}

{{- /* Create the tolerations of tf-operator pods to use. */ -}}
{{- define "tf-operator.tolerations" -}}
{{- with .Values.global.tolerations }}
{{- . | toYaml | nindent 6 }}
{{- end }}
{{- with .Values.tolerations }}
{{- . | toYaml | nindent 6 }}
{{- end }}
{{- if eq .Values.global.clusterProfile "Edge" }}
- key: node-role.alibabacloud.com/addon
  operator: Exists
  effect: NoSchedule
{{- end -}}
{{- end -}}
