{{/*
Expand the name of the chart.
*/}}
{{- define "kserve.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "kserve.fullname" -}}
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

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "kserve.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "kserve.labels" -}}
helm.sh/chart: {{ include "kserve.chart" . }}
{{ include "kserve.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "kserve.selectorLabels" -}}
app.kubernetes.io/name: {{ include "kserve.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "kserve.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "kserve.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}


{{/*
Support scale according to custom metrics
See the doc for details. https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/#scaling-on-custom-metrics
*/}}
{{- define "kserve.isCustomMetrics" -}}
{{- if .Values.scaleMetric }}
    {{- $supportedMetrics := list "cpu" "memory" "rps" "concurrency"}}
    {{- $metrics := .Values.scaleMetric | lower }}
    {{- if has $metrics $supportedMetrics }}
        {{- false }}
    {{- else }}
        {{- true }}
    {{- end -}}
{{- else }}
  {{- false }}
{{- end }}
{{- end -}}

{{- define "setAffinityFunction" -}}
{{- $affinityPolicy := .Values.affinityPolicy -}}
{{- $affinityConstraint := .Values.affinityConstraint -}}

{{- if eq $affinityPolicy "spread" -}}
{{- if eq $affinityConstraint "preferred" -}}
affinity:
  podAntiAffinity:
    preferredDuringSchedulingIgnoredDuringExecution:
      - podAffinityTerm:
          labelSelector:
            matchLabels:
              servingName: "{{ .Values.servingName }}"
          topologyKey: kubernetes.io/hostname
        weight: 100
{{- else if eq $affinityConstraint "required" -}}
affinity:
  podAntiAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
    - labelSelector:
        matchLabels:
          servingName: "{{ .Values.servingName }}"
      topologyKey: kubernetes.io/hostname
{{- end -}}

{{- else if eq $affinityPolicy "binpack" -}}
{{- if eq $affinityConstraint "preferred" -}}
affinity:
  podAffinity:
    preferredDuringSchedulingIgnoredDuringExecution:
      - podAffinityTerm:
          labelSelector:
            matchLabels:
              servingName: "{{ .Values.servingName }}"
          topologyKey: kubernetes.io/hostname
        weight: 100
{{- else if eq $affinityConstraint "required" -}}
affinity:
  podAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
    - labelSelector:
        matchLabels:
          servingName: "{{ .Values.servingName }}"
      topologyKey: kubernetes.io/hostname
{{- end -}}
{{- end -}}
{{- end -}}