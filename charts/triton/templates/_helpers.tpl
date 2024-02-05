{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "nvidia-triton-server.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "nvidia-triton-server.fullname" -}}
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
{{- define "nvidia-triton-server.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Return tritonserver image
*/}}
{{- define "triton.image" -}}
{{- if .Values.image }}
{{- .Values.image -}}
{{- else }}
{{- if eq .Values.backend "vllm" }}
{{- "nvcr.io/nvidia/tritonserver:24.01-vllm-python-py3" -}}
{{- else if eq .Values.backend "trt-llm" }}
{{- "nvcr.io/nvidia/tritonserver:24.01-trtllm-python-py3" -}}
{{- else }}
{{- "nvcr.io/nvidia/tritonserver:24.01-py3" -}}
{{- end }}
{{- end }}
{{- end -}}
