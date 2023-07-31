{{- template "settings" .}}{{$t := .Template}}*** Test Cases ***
{{- range $feature := .Features }}
{{- if eq $t "features"}}{{- template "testcases.features" $feature}}{{- end}}
{{- if eq $t "scenarios"}}{{- template "testcases.scenarios" $feature}}{{- end}}
{{- end}}