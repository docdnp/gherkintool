{{- define "testcases.features"}}{{- $f := .}}
{{ cleantitle $f.feature.name }}
    [Documentation]    {{ rmnewln $f.feature.description  }}
    [Tags]    {{range $tag := $f.feature.tags }}{{ puretag $tag.name }}    {{end}}{{ $f.uri }}
{{- range $scen_idx, $child := $f.feature.children }}
    Perform Scenario {{purename $f.uri }}: {{ $child.scenario.name }}
{{- range $step_idx, $step := $child.scenario.steps }}
    {{$step.keyword}}{{$step.text}}
{{- end}}
{{- end}}
{{end}}