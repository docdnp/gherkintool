{{- template "settings" .}}*** Test Cases ***
{{- range $feature := .Features }}
{{- template "testcases.scenarios" $feature}}
{{- end}}