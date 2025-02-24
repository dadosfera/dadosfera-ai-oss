{{/*
Name of the manifest.
*/}}
{{- define "library.metadata.name" -}}
{{- .Values.name | default .Chart.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
All labels
*/}}
{{- define "library.labels.all" -}}
{{ include "library.labels.selector" . }}
{{- end -}}

{{/*
Selector labels
*/}}
{{- define "library.labels.selector" -}}
app.kubernetes.io/part-of: orchest
app.kubernetes.io/name: {{ template "library.metadata.name" . }}
{{- end -}}

{{/*
Generate metadata
*/}}
{{- define "library.metadata" -}}
name: {{ template "library.metadata.name" . }}
namespace: {{ .Release.Namespace }}
labels:
    {{- include "library.labels.all" . | nindent 4 }}
{{- end -}}


{{/*
Get the replicas of the manifest.
*/}}
{{- define "library.spec.replicas" -}}
{{ .Values.replicas | default 1 }}
{{- end -}}

{{/*
Get the image pull policy.
*/}}
{{- define "library.spec.image.pullPolicy" -}}
  {{- if .Values.image.pullPolicy -}}
    {{ .Values.image.pullPolicy }}
  {{- else -}}
    {{ "IfNotPresent" }}
  {{- end }}
{{- end -}}
