{{/*
Expand the name of the chart.
*/}}
{{- define "redroid-operator.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "redroid-operator.fullname" -}}
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
Create chart label value.
*/}}
{{- define "redroid-operator.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels.
*/}}
{{- define "redroid-operator.labels" -}}
helm.sh/chart: {{ include "redroid-operator.chart" . }}
{{ include "redroid-operator.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: redroid-operator
{{- end }}

{{/*
Selector labels.
*/}}
{{- define "redroid-operator.selectorLabels" -}}
app.kubernetes.io/name: {{ include "redroid-operator.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
ServiceAccount name.
*/}}
{{- define "redroid-operator.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "redroid-operator.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Manager container image (repo:tag, tag defaults to appVersion).
*/}}
{{- define "redroid-operator.image" -}}
{{- printf "%s:%s" .Values.image.repository (default .Chart.AppVersion .Values.image.tag) }}
{{- end }}

{{/*
kmsg-tools image (repo:tag, tag defaults to operator image tag or appVersion).
*/}}
{{- define "redroid-operator.kmsgToolsImage" -}}
{{- $kmsgRepo := .Values.kmsgToolsImage.repository | default "ghcr.io/isning/redroid-operator/kmsg-tools" -}}
{{- $kmsgTag := (or .Values.kmsgToolsImage.tag .Values.image.tag .Chart.AppVersion) -}}
{{- printf "%s:%s" $kmsgRepo $kmsgTag -}}
{{- end }}
