{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "tfjob.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "tfjob.fullname" -}}
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

{{- define "policy.api" }}
{{- if .Capabilities.APIVersions.Has "policy/v1beta1" -}}
v1beta1
{{- else -}}
v1
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "tfjob.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}


{{- define "setPSAffinityFunction" -}}
{{- $affinityPolicy := .Values.psAffinityPolicy -}}
{{- $affinityConstraint := .Values.psAffinityConstraint -}}

{{- if eq $affinityPolicy "spread" -}}
{{- if eq $affinityConstraint "preferred" -}}
affinity:
  podAntiAffinity:
    preferredDuringSchedulingIgnoredDuringExecution:
      - weight: 100
        podAffinityTerm:
          topologyKey: kubernetes.io/hostname
          labelSelector:
            matchExpressions:
              - key: release
                operator: In
                values:
                  - "{{ .Release.Name }}"
              - key: group-name
                operator: In
                values:
                  - "kubeflow.org"
              - key: tf-replica-type
                operator: In
                values:
                  - ps
{{- else if eq $affinityConstraint "required" -}}
affinity:
  podAntiAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
      - topologyKey: kubernetes.io/hostname
        labelSelector:
          matchExpressions:
            - key: release
              operator: In
              values:
                - "{{ .Release.Name }}"
            - key: group-name
              operator: In
              values:
                - "kubeflow.org"
            - key: tf-replica-type
              operator: In
              values:
                - ps
{{- end -}}

{{- else if eq $affinityPolicy "binpack" -}}
{{- if eq $affinityConstraint "preferred" -}}
affinity:
  podAffinity:
    preferredDuringSchedulingIgnoredDuringExecution:
    - weight: 100
      podAffinityTerm:
        topologyKey: kubernetes.io/hostname
        labelSelector:
          matchExpressions:
            - key: release
              operator: In
              values:
                - "{{ .Release.Name }}"
            - key: group-name
              operator: In
              values:
                - "kubeflow.org"
    - weight: 60
      podAffinityTerm:
        topologyKey: kubernetes.io/hostname
        labelSelector:
          matchExpressions:
            - key: tf-replica-type
              operator: In
              values:
                - worker
    - weight: 30
      podAffinityTerm:
        topologyKey: kubernetes.io/hostname
        labelSelector:
          matchExpressions:
            - key: tf-replica-type
              operator: In
              values:
                - ps
{{- else if eq $affinityConstraint "required" -}}
affinity:
  podAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
    - topologyKey: kubernetes.io/hostname
      labelSelector:
        matchExpressions:
          - key: release
            operator: In
            values:
              - "{{ .Release.Name }}"
          - key: group-name
            operator: In
            values:
              - "kubeflow.org"
          - key: tf-replica-type
            operator: In
            values:
              - ps
{{- end -}}
{{- end -}}
{{- end -}}


{{- define "setWorkerAffinityFunction" -}}
{{- $affinityPolicy := .Values.workerAffinityPolicy -}}
{{- $affinityConstraint := .Values.workerAffinityConstraint -}}

{{- if eq $affinityPolicy "spread" -}}
{{- if eq $affinityConstraint "preferred" -}}
affinity:
  podAntiAffinity:
    preferredDuringSchedulingIgnoredDuringExecution:
      - weight: 100
        podAffinityTerm:
          topologyKey: kubernetes.io/hostname
          labelSelector:
            matchExpressions:
              - key: release
                operator: In
                values:
                  - "{{ .Release.Name }}"
              - key: group-name
                operator: In
                values:
                  - "kubeflow.org"
              - key: tf-replica-type
                operator: In
                values:
                  - worker
{{- else if eq $affinityConstraint "required" -}}
affinity:
  podAntiAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
      - topologyKey: kubernetes.io/hostname
        labelSelector:
          matchExpressions:
            - key: release
              operator: In
              values:
                - "{{ .Release.Name }}"
            - key: group-name
              operator: In
              values:
                - "kubeflow.org"
            - key: tf-replica-type
              operator: In
              values:
                - worker
{{- end -}}
{{- else if eq $affinityPolicy "binpack" -}}
{{- if eq $affinityConstraint "preferred" -}}
affinity:
  podAffinity:
    preferredDuringSchedulingIgnoredDuringExecution:
    - weight: 100
      podAffinityTerm:
        topologyKey: kubernetes.io/hostname
        labelSelector:
          matchExpressions:
            - key: release
              operator: In
              values:
                - "{{ .Release.Name }}"
            - key: group-name
              operator: In
              values:
                - "kubeflow.org"
    - weight: 60
      podAffinityTerm:
        topologyKey: kubernetes.io/hostname
        labelSelector:
          matchExpressions:
            - key: tf-replica-type
              operator: In
              values:
                - ps
    - weight: 30
      podAffinityTerm:
        topologyKey: kubernetes.io/hostname
        labelSelector:
          matchExpressions:
            - key: tf-replica-type
              operator: In
              values:
                - worker
{{- else if eq $affinityConstraint "required" -}}
affinity:
  podAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
    - topologyKey: kubernetes.io/hostname
      labelSelector:
        matchExpressions:
          - key: release
            operator: In
            values:
              - "{{ .Release.Name }}"
          - key: group-name
            operator: In
            values:
              - "kubeflow.org"
          - key: tf-replica-type
            operator: In
            values:
              - worker
{{- end -}}
{{- end -}}
{{- end -}}