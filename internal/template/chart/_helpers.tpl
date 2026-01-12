{{/*
Expand the name of the chart.
*/}}
{{- define "app.name" -}}
{{ .Chart.Name }}
{{- end -}}

{{/*
Create a fullname using the release name and chart name.
*/}}
{{- define "app.fullname" -}}
{{ .Release.Name }}-{{ .Chart.Name }}
{{- end -}}

