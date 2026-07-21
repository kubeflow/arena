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
{{- define "arena.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- /*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/ -}}
{{- define "arena.fullname" -}}
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
{{- define "arena.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- /* Common labels */ -}}
{{- define "arena.labels" -}}
helm.sh/chart: {{ include "arena.chart" . }}
{{ include "arena.selectorLabels" . }}
{{- with .Chart.AppVersion }}
app.kubernetes.io/version: {{ . | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{- /* Selector labels. */ -}}
{{- define "arena.selectorLabels" -}}
app.kubernetes.io/name: {{ include "arena.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{- /* Common annotations. */ -}}
{{- define "arena.annotations" -}}
{{- with .Values.annotations }}
{{- . | toYaml | nindent 0 }}
{{- end }}
{{- end -}}

{{- /* Create the name of the service account to use. */ -}}
{{- define "arena.serviceAccount.name" -}}
{{- include "arena.fullname" . }}
{{- end -}}

{{- /* Create the name of the cluster role to use. */ -}}
{{- define "arena.clusterRole.name" -}}
{{- include "arena.fullname" . }}
{{- end -}}

{{- /* Create the name of the cluster role binding to use. */ -}}
{{- define "arena.clusterRoleBinding.name" -}}
{{- include "arena.fullname" . }}
{{- end -}}

{{- /* Create the name of the role to use. */ -}}
{{- define "arena.role.name" -}}
{{- include "arena.fullname" . }}
{{- end -}}

{{- /* Create the name of the role binding to use. */ -}}
{{- define "arena.roleBinding.name" -}}
{{- include "arena.fullname" . }}
{{- end -}}

{{- /* Create the name of the pre-upgrade hook job. */ -}}
{{- define "arena.pre-upgrade.job.name" -}}
{{- include "arena.fullname" . -}}-pre-upgrade
{{- end -}}

{{- /* Create the name of the installer job. */ -}}
{{- define "arena.installer.job.name" -}}
{{- include "arena.fullname" . -}}-installer
{{- end -}}

{{- /* Create the image registry used by the installer job. */ -}}
{{- define "arena.installer.imageRegistry" -}}
{{ .Values.binary.image.registry | default .Values.global.imagePrefix | default .Values.global.image.registry }}
{{- end -}}

{{- /* Create the image of used by the installer job. */ -}}
{{- define "arena.installer.image" -}}
{{- $imageRegistry := .Values.binary.image.registry | default .Values.global.imagePrefix | default .Values.global.image.registry }}
{{- $imageRepository := .Values.binary.image.repository }}
{{- $imageTag := .Values.binary.image.tag | default .Chart.Version }}
{{- printf "%s/%s:%s" $imageRegistry $imageRepository $imageTag }}
{{- end -}}
