{{- define "arena.imagePrefix" -}}
{{- if .Values.global.pullImageByVPCNetwork }}
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
{{- if eq .Values.global.clusterProfile "Edge" }}
alibabacloud.com/is-edge-worker: "false"
{{- end }}
{{- end }}

{{- define "arena.version" }}
{{- .Values.binary.tag }}
{{- end }}

{{- define "arena.labels" -}}
helm.sh/chart: arena-artifacts
app.kubernetes.io/managed-by: arena
{{- end }}
