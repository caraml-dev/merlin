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

{{- define "merlin-postgresql.host" -}}
{{- if index .Values "merlin-postgresql" "enabled" -}}
{{- printf "%s-merlin-postgresql.%s.svc.cluster.local" .Release.Name .Release.Namespace -}}
{{- else -}}
{{- index .Values "merlin-postgresql" "postgresqlHost" -}}
{{- end -}}
{{- end -}}

{{- define "mlflow-postgresql.host" -}}
{{- if index .Values "mlflow-postgresql" "enabled" -}}
{{- printf "%s-mlflow-postgresql" "%s.svc.cluster.local" .Release.Name .Release.Namespace -}}
{{- else -}}
{{- index .Values "mlflow-postgresql" "postgresqlHost" -}}
{{- end -}}
{{- end -}}

{{- define "mlflow.backendStoreUri" -}}
{{- if (index .Values "mlflow-postgresql" "enabled") -}}
{{- printf "postgresql://%s:%s@%s:5432/%s" (index .Values "mlflow-postgresql" "postgresqlUsername") (index .Values "mlflow-postgresql" "postgresqlPassword") (include "mlflow-postgresql.host" .) (index .Values "mlflow-postgresql" "postgresqlDatabase") -}}
{{- else if (index .Values "mlflow-postgresql" "postgresqlHost") -}}
{{- printf "postgresql://%s:%s@%s:5432/%s" (index .Values "mlflow-postgresql" "postgresqlUsername") (index .Values "mlflow-postgresql" "postgresqlPassword") (index .Values "mlflow-postgresql" "postgresqlHost") (index .Values "mlflow-postgresql" "postgresqlDatabase") -}}
{{- else -}}
{{- printf .Values.mlflow.backendStoreUri -}}
{{- end -}}
{{- end -}}
