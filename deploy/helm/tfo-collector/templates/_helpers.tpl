{{/*
Expand the name of the chart.
*/}}
{{- define "tfo-collector.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
Truncate at 63 chars (Kubernetes name limit).
*/}}
{{- define "tfo-collector.fullname" -}}
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
Create chart label value (name-version).
*/}}
{{- define "tfo-collector.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels applied to all resources.
*/}}
{{- define "tfo-collector.labels" -}}
helm.sh/chart: {{ include "tfo-collector.chart" . }}
{{ include "tfo-collector.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- with .Values.commonLabels }}
{{ toYaml . }}
{{- end }}
{{- end }}

{{/*
Selector labels (used in matchLabels and pod selectors — must be stable).
*/}}
{{- define "tfo-collector.selectorLabels" -}}
app.kubernetes.io/name: {{ include "tfo-collector.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
ServiceAccount name — use custom name if provided, else generate from fullname.
*/}}
{{- define "tfo-collector.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "tfo-collector.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Container image reference — tag defaults to appVersion.
*/}}
{{- define "tfo-collector.image" -}}
{{- $tag := .Values.image.tag | default .Chart.AppVersion }}
{{- printf "%s:%s" .Values.image.repository $tag }}
{{- end }}

{{/*
Name of the credentials Secret.
*/}}
{{- define "tfo-collector.secretName" -}}
{{- printf "%s-credentials" (include "tfo-collector.fullname" .) }}
{{- end }}

{{/*
Name of the ConfigMap for the collector config.
*/}}
{{- define "tfo-collector.configMapName" -}}
{{- printf "%s-config" (include "tfo-collector.fullname" .) }}
{{- end }}

{{/*
Name of the queue PVC.
*/}}
{{- define "tfo-collector.queuePvcName" -}}
{{- printf "%s-queue" (include "tfo-collector.fullname" .) }}
{{- end }}
