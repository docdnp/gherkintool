{{- define "testcases.scenarios"}}{{- $f := .}}
{{- range $scen_idx, $child := $f.feature.children }}
{{ purename $f.uri }}= {{ cleantitle $child.scenario.name }}
    [Documentation]    {{ rmnewln $f.feature.description  }}
    [Tags]    {{range $tag := $child.scenario.tags }}{{ puretag $tag.name }}    {{end}}{{ $f.uri }}
    Perform Scenario {{ purename $f.uri }}: {{ $f.feature.name }}
{{range $s := $child.scenario.steps }}    {{$s.keyword}}{{$s.text}}
{{end}}
{{- end}}
{{- end}}