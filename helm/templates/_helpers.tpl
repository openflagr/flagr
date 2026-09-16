{{/*
Expand the name of the chart.
*/}}
{{- define "flagr.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "flagr.fullname" -}}
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
{{- define "flagr.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "flagr.labels" -}}
helm.sh/chart: {{ include "flagr.chart" . }}
{{ include "flagr.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels (shared).
*/}}
{{- define "flagr.selectorLabels" -}}
app.kubernetes.io/name: {{ include "flagr.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Primary (SQLite/SQL writer, or the only Deployment).
*/}}
{{- define "flagr.primarySelectorLabels" -}}
{{ include "flagr.selectorLabels" . }}
app.kubernetes.io/component: primary
{{- end }}

{{/*
Eval-only replicas (json_http readers). Separate Service so CRUD never hits them.
*/}}
{{- define "flagr.evalSelectorLabels" -}}
{{ include "flagr.selectorLabels" . }}
app.kubernetes.io/component: eval
{{- end }}

{{- define "flagr.evalFullname" -}}
{{- printf "%s-eval" (include "flagr.fullname" .) | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
URL eval replicas poll. Override evalReplicas.flagsURL if FLAGR_WEB_PREFIX is set.
*/}}
{{- define "flagr.evalFlagsURL" -}}
{{- if .Values.evalReplicas.flagsURL }}
{{- .Values.evalReplicas.flagsURL }}
{{- else }}
{{- printf "http://%s:%v/api/v1/export/eval_cache/json" (include "flagr.fullname" .) .Values.service.port }}
{{- end }}
{{- end }}

{{/*
Container image (tag defaults to Chart.appVersion).
*/}}
{{- define "flagr.image" -}}
{{- printf "%s:%s" .Values.image.repository (.Values.image.tag | default .Chart.AppVersion) }}
{{- end }}

{{/*
Fail closed on incompatible scaling modes.
*/}}
{{- define "flagr.validate" -}}
{{- if and .Values.gitops.enabled (gt (int .Values.evalReplicas.replicaCount) 0) }}
{{- fail "gitops.enabled and evalReplicas.replicaCount>0 cannot be combined: GitOps is one json_http Deployment (scale replicaCount); evalReplicas is SQLite HA only." }}
{{- end }}
{{- $hasDSN := not (empty .Values.gitops.flagsURL) }}
{{- range .Values.env }}
{{- if eq .name "FLAGR_DB_DBCONNECTIONSTR" }}{{- $hasDSN = true }}{{- end }}
{{- end }}
{{- if .Values.envFrom }}{{- $hasDSN = true }}{{- end }}
{{- if and .Values.gitops.enabled (not $hasDSN) }}
{{- fail "gitops.enabled requires gitops.flagsURL (public raw URL) or env/envFrom FLAGR_DB_DBCONNECTIONSTR (private GitHub PAT URL)." }}
{{- end }}
{{- end }}
