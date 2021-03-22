{{- define "merlin.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "merlin.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version -}}
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

{{- define "database.host" -}}
{{- if .Values.postgresql.enabled }}
{{- printf "%s-postgresql.%s.svc.cluster.local" .Release.Name .Release.Namespace -}}
{{- else -}}
{{- .Values.merlin.database.host -}}
{{- end }}
{{- end -}}

{{- define "database.user" -}}
{{- if .Values.postgresql.enabled }}
{{- .Values.postgresql.postgresqlUsername -}}
{{- else -}}
{{- .Values.merlin.database.user -}}
{{- end }}
{{- end -}}

{{- define "database.name" -}}
{{- if .Values.postgresql.enabled }}
{{- .Values.postgresql.postgresqlDatabase -}}
{{- else -}}
{{- .Values.merlin.database.name -}}
{{- end }}
{{- end -}}

{{- define "database.password" -}}
{{- if .Values.postgresql.enabled }}
valueFrom:
  secretKeyRef:
    name: {{ .Release.Name }}-postgresql
    key: postgresql-password
{{- else -}}
valueFrom:
  secretKeyRef:
    name: {{ .Values.merlin.database.password.secretName }}
    key:  {{ .Values.merlin.database.password.secretKey }}
{{- end }}
{{- end -}}
