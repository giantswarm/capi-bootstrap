{{/* vim: set filetype=mustache: */}}
{{/* Expand the name of the chart. This is suffixed with -alertmanager, which means subtract 13 from longest 63 available */}}
{{- define "capi-bootstrap.name" -}}
{{- .Chart.Name | trunc 50 | trimSuffix "-" -}}
{{- end }}
