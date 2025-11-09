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
{{- define "gpu-exporter.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- /*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/ -}}
{{- define "gpu-exporter.fullname" -}}
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
{{- define "gpu-exporter.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- /* Common labels */ -}}
{{- define "gpu-exporter.labels" -}}
helm.sh/chart: {{ include "gpu-exporter.chart" . }}
{{ include "gpu-exporter.selectorLabels" . }}
{{- with .Chart.AppVersion }}
app.kubernetes.io/version: {{ . | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- with .Values.labels }}
{{- . | toYaml | nindent 0 }}
{{- end }}
{{- end }}

{{- /* Selector labels. */ -}}
{{- define "gpu-exporter.selectorLabels" -}}
app.kubernetes.io/name: {{ include "gpu-exporter.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{- /* Common annotations. */ -}}
{{- define "gpu-exporter.annotations" -}}
{{- with .Values.annotations }}
{{- . | toYaml | nindent 0 }}
{{- end }}
{{- end -}}

{{- /* Create the name of the daemon set to use. */ -}}
{{- define "gpu-exporter.daemonSet.name" -}}
{{- include "gpu-exporter.fullname" . }}
{{- end -}}

{{- /* Create the name of the service to use. */ -}}
{{- define "gpu-exporter.service.name" -}}
{{- include "gpu-exporter.fullname" . }}
{{- end -}}

{{- /* Create the name of the service monitor to use. */ -}}
{{- define "gpu-exporter.serviceMonitor.name" -}}
{{- include "gpu-exporter.fullname" . }}
{{- end -}}

{{- /* Create the name of the image to use. */ -}}
{{- define "gpu-exporter.image" -}}
{{- $imageRegistry := .Values.image.registry | default .Values.global.imagePrefix | default .Values.global.image.registry }}
{{- $imageRepository := .Values.image.repository }}
{{- $imageTag := .Values.image.tag | default .Chart.AppVersion }}
{{- printf "%s/%s:%s" $imageRegistry $imageRepository $imageTag }}
{{- end -}}

{{- /* Create the nodeSelector of gpu-exporter pods to use. */ -}}
{{- define "gpu-exporter.nodeSelector" -}}
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

{{- /* Create the tolerations of gpu-exporter pods to use. */ -}}
{{- define "gpu-exporter.tolerations" -}}
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
