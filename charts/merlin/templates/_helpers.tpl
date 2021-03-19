{{- define "merlin.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "merlin.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "merlin.fullname" -}}
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

{{- define "mlflow.fullname" -}}
{{- if .Values.mlflow.fullnameOverride -}}
{{- .Values.mlflow.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- printf "%s-%s" .Release.Name .Values.mlflow.name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s-%s" .Release.Name $name .Values.mlflow.name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{- define "merlinPostgresql.host" -}}
{{- if .Values.merlinPostgresql.enabled -}}
{{- printf "%s-postgresql.%s.svc.cluster.local" .Release.Name .Release.Namespace -}}
{{- else -}}
{{- .Values.merlinPostgresql.postgresqlHost -}}
{{- end -}}
{{- end -}}

{{- define "mlflowPostgresql.host" -}}
{{- if .Values.mlflowPostgresql.enabled -}}
{{- printf "%s-postgresql.%s.svc.cluster.local" .Release.Name .Release.Namespace -}}
{{- else -}}
{{- .Values.mlflowPostgresql.postgresqlHost -}}
{{- end -}}
{{- end -}}

{{- define "mlflow.backendStoreUri" -}}
{{- if .Values.mlflowPostgresql.enabled -}}
{{- printf "postgresql://%s:%s@%s:5432/%s" .Values.mlflowPostgresql.postgresqlUsername .Values.mlflowPostgresql.postgresqlPassword (include "mlflowPostgresql.host" .) .Values.mlflowPostgresql.postgresqlDatabase -}}
{{- else -}}
{{- printf .Values.mlflow.backendStoreUri -}}
{{- end -}}
{{- end -}}
