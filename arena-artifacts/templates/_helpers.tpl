{{- define "arena.imagePrefix" -}}
{{- if eq .Values.global.clusterProfile "Edge" }}
{{- .Values.global.imagePrefix }}
{{- else if .Values.global.pullImageByVPCNetwork }}
{{- .Values.global.imagePrefix | replace "registry." "registry-vpc." }}
{{- else }}
{{- .Values.global.imagePrefix }}
{{- end }}
{{- end }}

{{- define "arena.nodeSelector" }}
{{- range $nodeKey,$nodeVal := .Values.nodeSelector }}
{{ $nodeKey }}: "{{ $nodeVal }}"
{{- end }}
{{- range $nodeKey,$nodeVal := .Values.global.nodeSelector }}
{{ $nodeKey }}: "{{ $nodeVal }}"
{{- end }}
{{- end }}

{{- define "arena.nonEdgeNodeSelector" }}
{{- if eq .Values.global.clusterProfile "Edge" }}
alibabacloud.com/is-edge-worker: "false"
{{- end }}
{{- end }}

{{- define "arena.tolerateNonEdgeNodeSelector" }}
{{- if eq .Values.global.clusterProfile "Edge" }}
- key: node-role.alibabacloud.com/addon
  operator: Exists
  effect: NoSchedule
{{- end }}
{{- end }}

{{- define "arena.version" }}
{{- .Values.binary.tag }}
{{- end }}

{{- define "arena.labels" -}}
helm.sh/chart: arena-artifacts
app.kubeflow.org/managed-by: arena
{{- end }}

{{- define "crd.api" }}
{{- if .Capabilities.APIVersions.Has "apiextensions.k8s.io/v1beta1" -}}
v1beta1
{{- else -}}
v1 
{{- end }}
{{- end }}
