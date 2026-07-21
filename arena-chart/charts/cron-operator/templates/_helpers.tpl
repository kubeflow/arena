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
{{- define "cron-operator.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- /*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/ -}}
{{- define "cron-operator.fullname" -}}
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
{{- define "cron-operator.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- /* Common labels */ -}}
{{- define "cron-operator.labels" -}}
helm.sh/chart: {{ include "cron-operator.chart" . }}
{{ include "cron-operator.selectorLabels" . }}
{{- with .Chart.AppVersion }}
app.kubernetes.io/version: {{ . | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- with .Values.labels }}
{{- . | toYaml | nindent 0 }}
{{- end }}
{{- end }}

{{- /* Selector labels. */ -}}
{{- define "cron-operator.selectorLabels" -}}
app.kubernetes.io/name: {{ include "cron-operator.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{- /* Common annotations. */ -}}
{{- define "cron-operator.annotations" -}}
{{- with .Values.annotations }}
{{- . | toYaml | nindent 0 }}
{{- end }}
{{- end -}}

{{- /* Create the name of the service account to use. */ -}}
{{- define "cron-operator.serviceAccount.name" -}}
{{- include "cron-operator.fullname" . }}
{{- end -}}

{{- /* Create the name of the cluster role to use. */ -}}
{{- define "cron-operator.clusterRole.name" -}}
{{- include "cron-operator.fullname" . }}
{{- end -}}

{{- /* Create the name of the cluster role binding to use. */ -}}
{{- define "cron-operator.clusterRoleBinding.name" -}}
{{- include "cron-operator.fullname" . }}
{{- end -}}

{{- /* Create the name of the deployment to use. */ -}}
{{- define "cron-operator.deployment.name" -}}
{{- include "cron-operator.fullname" . }}
{{- end -}}

{{- /* Create the name of the service to use. */ -}}
{{- define "cron-operator.service.name" -}}
{{- include "cron-operator.fullname" . }}
{{- end -}}

{{- /* Create the name of the image to use. */ -}}
{{- define "cron-operator.image" -}}
{{- $imageRegistry := .Values.image.registry | default .Values.global.imagePrefix | default .Values.global.image.registry }}
{{- $imageRepository := .Values.image.repository }}
{{- $imageTag := .Values.image.tag | default .Chart.AppVersion }}
{{- printf "%s/%s:%s" $imageRegistry $imageRepository $imageTag }}
{{- end -}}

{{- /* Create the nodeSelector of cron-operator pods to use. */ -}}
{{- define "cron-operator.nodeSelector" -}}
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

{{- /* Create the affinity of cron-operator pods to use. */ -}}
{{- define "cron-operator.affinity" -}}
podAntiAffinity:
  preferredDuringSchedulingIgnoredDuringExecution:
  - weight: 100
    podAffinityTerm:
      labelSelector:
        matchLabels:
          {{- include "cron-operator.selectorLabels" . | nindent 10 }}
      topologyKey: kubernetes.io/hostname
{{- end -}}

{{- /* Create the tolerations of cron-operator pods to use. */ -}}
{{- define "cron-operator.tolerations" -}}
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
{{- end }}
{{- end -}}
