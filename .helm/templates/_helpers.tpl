{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "jfrog-cleaner.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "cmd" -}}
{{- printf " jfrog-cleaner delete -H %s -u %s -p %s -f %s  -n %d -c %s -r %s" .Values.settings.host .Values.settings.user .Values.settings.password .Values.settings.folder (.Values.settings.recentNumber | int) .Values.settings.period .Values.settings.repo -}}
{{ $length := len .Values.settings.filterTags }} {{- if gt $length 0 }}
{{- print " -j " }}
{{- join " -j " .Values.settings.filterTags }}
{{- end }}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "jfrog-cleaner.fullname" -}}
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
{{- define "jfrog-cleaner.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "jfrog-cleaner.labels" -}}
helm.sh/chart: {{ include "jfrog-cleaner.chart" . }}
{{ include "jfrog-cleaner.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{/*
Selector labels
*/}}
{{- define "jfrog-cleaner.selectorLabels" -}}
app.kubernetes.io/name: {{ include "jfrog-cleaner.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{/*
Create the name of the service account to use
*/}}
{{- define "jfrog-cleaner.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
    {{ default (include "jfrog-cleaner.fullname" .) .Values.serviceAccount.name }}
{{- else -}}
    {{ default "default" .Values.serviceAccount.name }}
{{- end -}}
{{- end -}}
