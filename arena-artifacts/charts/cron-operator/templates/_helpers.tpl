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

{{- /* Common labels. */ -}}
{{- define "cron-operator.labels" -}}
helm.sh/chart: {{ include "cron-operator.chart" . }}
{{ include "cron-operator.selectorLabels" . }}
{{- with .Chart.AppVersion }}
app.kubernetes.io/version: {{ . | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{- /* Selector labels. */ -}}
{{- define "cron-operator.selectorLabels" -}}
app.kubernetes.io/name: {{ include "cron-operator.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{- /* Name of the service account. */ -}}
{{- define "cron-operator.serviceAccount.name" -}}
{{- include "cron-operator.fullname" . }}
{{- end }}

{{- /* Name of the cluster role. */ -}}
{{- define "cron-operator.clusterRole.name" -}}
{{- include "cron-operator.fullname" . }}
{{- end }}

{{- /* Name of the cluster role binding. */ -}}
{{- define "cron-operator.clusterRoleBinding.name" -}}
{{- include "cron-operator.fullname" . }}
{{- end }}

{{- /* Name of the deployment. */ -}}
{{- define "cron-operator.deployment.name" -}}
{{- include "cron-operator.fullname" . }}
{{- end }}

{{- /* Cron operator image. */ -}}
{{- define "cron-operator.image" -}}
{{- printf "%s/%s:%s" .Values.image.registry .Values.image.repository (.Values.image.tag | default .Chart.AppVersion | default .Chart.Version) -}}
{{- end -}}

{{- /* Name of the service. */ -}}
{{- define "cron-operator.service.name" -}}
{{- include "cron-operator.fullname" . }}
{{- end }}
