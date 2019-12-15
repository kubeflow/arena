{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "runai.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "runai.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "runai.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "charts.label-addition"}}
app: {{ template "runai.name" . }}
chart: {{ template "runai.chart" . }}
release: {{ .Release.Name }}
heritage: {{ .Release.Service }}
createdBy: "RunaiJob"
{{- end }}

{{/* Generate basic labels */}}
{{- define "chart.labels" }}
labels:
  {{include "charts.label-addition" . | indent 2}}
  app: {{ template "runai.name" . }}
  chart: {{ template "runai.chart" . }}
  release: {{ .Release.Name }}
  heritage: {{ .Release.Service }}
  createdBy: "RunaiJob"
{{- end }}
