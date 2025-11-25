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
{{- define "tf-dashboard.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- /*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/ -}}
{{- define "tf-dashboard.fullname" -}}
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
{{- define "tf-dashboard.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- /* Common labels */ -}}
{{- define "tf-dashboard.labels" -}}
helm.sh/chart: {{ include "tf-dashboard.chart" . }}
{{ include "tf-dashboard.selectorLabels" . }}
{{- with .Chart.AppVersion }}
app.kubernetes.io/version: {{ . | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- with .Values.labels }}
{{- . | toYaml | nindent 0 }}
{{- end }}
{{- end }}

{{- /* Selector labels. */ -}}
{{- define "tf-dashboard.selectorLabels" -}}
app.kubernetes.io/name: {{ include "tf-dashboard.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{- /* Common annotations. */ -}}
{{- define "tf-dashboard.annotations" -}}
{{- with .Values.annotations }}
{{- . | toYaml | nindent 0 }}
{{- end }}
{{- end -}}

{{- /* Create the name of the service account to use. */ -}}
{{- define "tf-dashboard.serviceAccount.name" -}}
{{- include "tf-dashboard.fullname" . }}
{{- end -}}

{{- /* Create the name of the cluster role to use. */ -}}
{{- define "tf-dashboard.clusterRole.name" -}}
{{- include "tf-dashboard.fullname" . }}
{{- end -}}

{{- /* Create the name of the cluster role binding to use. */ -}}
{{- define "tf-dashboard.clusterRoleBinding.name" -}}
{{- include "tf-dashboard.fullname" . }}
{{- end -}}

{{- /* Create the name of the deployment to use. */ -}}
{{- define "tf-dashboard.deployment.name" -}}
{{- include "tf-dashboard.fullname" . }}
{{- end -}}

{{- /* Create the name of the service to use. */ -}}
{{- define "tf-dashboard.service.name" -}}
{{- include "tf-dashboard.fullname" . }}
{{- end -}}

{{- /* Create the name of the image to use. */ -}}
{{- define "tf-dashboard.image" -}}
{{- $imageRegistry := .Values.image.registry | default .Values.global.imagePrefix | default .Values.global.image.registry }}
{{- $imageRepository := .Values.image.repository }}
{{- $imageTag := .Values.image.tag | default .Chart.AppVersion }}
{{- printf "%s/%s:%s" $imageRegistry $imageRepository $imageTag }}
{{- end -}}

{{- /* Create the nodeSelector of tf-dashboard pods to use. */ -}}
{{- define "tf-dashboard.nodeSelector" -}}
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

{{- /* Create the affinity of tf-dashboard pods to use. */ -}}
{{- define "tf-dashboard.affinity" -}}
podAntiAffinity:
  preferredDuringSchedulingIgnoredDuringExecution:
  - weight: 100
    podAffinityTerm:
      labelSelector:
        matchLabels:
          {{- include "tf-dashboard.selectorLabels" . | nindent 10 }}
      topologyKey: kubernetes.io/hostname
{{- end -}}

{{- /* Create the tolerations of tf-dashboard pods to use. */ -}}
{{- define "tf-dashboard.tolerations" -}}
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
