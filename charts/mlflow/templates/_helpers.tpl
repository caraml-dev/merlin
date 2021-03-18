{{/* vim: set filetype=mustache: */}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
*/}}
{{- define "mlflow.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- printf .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}

{{- define "postgresql.host" -}}
{{- if eq .Values.postgresql.type "external" -}}
{{- .Values.postgresql.postgresqlServer -}}
{{- else -}}
{{- printf "%s-postgresql.%s.svc.cluster.local" .Release.Name .Release.Namespace -}}
{{- end -}}
{{- end -}}

{{- define "mlflow.backendStoreUri" -}}
{{- if .Values.postgresql.enabled -}}
{{- printf "postgresql://%s:%s@%s:5432/%s" .Values.postgresql.postgresqlUsername .Values.postgresql.postgresqlPassword (include "postgresql.host" .) .Values.postgresql.postgresqlDatabase -}}
{{- else -}}
{{- printf .Values.backendStoreUri -}}
{{- end -}}
{{- end -}}
