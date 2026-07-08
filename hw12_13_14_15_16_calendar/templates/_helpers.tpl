{{- define "calendar.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "calendar.fullname" -}}
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

{{- define "calendar.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "calendar.labels" -}}
helm.sh/chart: {{ include "calendar.chart" . }}
app.kubernetes.io/name: {{ include "calendar.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- with .Values.commonLabels }}
{{ toYaml . }}
{{- end }}
{{- end -}}

{{- define "calendar.selectorLabels" -}}
app.kubernetes.io/name: {{ include "calendar.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{- define "calendar.componentName" -}}
{{- printf "%s-%s" (include "calendar.fullname" .root) .component | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "calendar.componentLabels" -}}
{{ include "calendar.labels" .root }}
app.kubernetes.io/component: {{ .component }}
{{- end -}}

{{- define "calendar.componentSelectorLabels" -}}
{{ include "calendar.selectorLabels" .root }}
app.kubernetes.io/component: {{ .component }}
{{- end -}}

{{- define "calendar.image" -}}
{{- printf "%s:%s" .Values.image.repository (.Values.image.tag | default .Chart.AppVersion) -}}
{{- end -}}

{{- define "calendar.componentImage" -}}
{{- $repository := default .root.Values.image.repository .config.image.repository -}}
{{- $tag := default (default .root.Chart.AppVersion .root.Values.image.tag) .config.image.tag -}}
{{- printf "%s:%s" $repository $tag -}}
{{- end -}}
