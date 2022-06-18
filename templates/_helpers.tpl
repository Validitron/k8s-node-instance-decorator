{{/*
Expand the name of the chart.
*/}}
{{- define "node-instance-decorator.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "node-instance-decorator.fullname" -}}
{{- $chartName := include "node-instance-decorator.name" . }}
{{- default $chartName .Values.fullNameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "node-instance-decorator.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "node-instance-decorator.labels" -}}
helm.sh/chart: {{ include "node-instance-decorator.chart" . }}
{{ include "node-instance-decorator.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- if .Values.context.domainUsername }}
app.kubernetes.io/created-by: {{ .Values.context.domainUsername | trunc 63 | quote }}
{{- end }}
{{- if .Values.context.correlationId }}
tron/correlationId: {{ .Values.context.correlationId | trunc 63 | quote }}
{{- else }}
{{- $correlationId := include "node-instance-decorator.fullname" . | b64enc | replace "=" "" | trunc 63 | quote }}
tron/correlationId: {{ $correlationId }}
{{- end }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "node-instance-decorator.selectorLabels" -}}
app.kubernetes.io/name: {{ include "node-instance-decorator.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
