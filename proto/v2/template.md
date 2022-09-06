{{range $file := .Files}}

{{range $service := .Services -}}
{{range .Methods -}}

### {{.Name}}

{{ .Description}}

> rpc {{$file.Package}}.{{$service.Name}}/{{.Name}}([{{.RequestLongType}}](#{{.RequestLongType | lower | replace "." "-"}}))
>   returns [{{.ResponseLongType}}](#{{.ResponseLongType | lower | replace "." "-"}})

{{end}}
{{end}}

------
### Messages

{{range .Messages}}

#### {{.LongName}}

{{.Description}}

{{if .HasFields}}
| Field | Type | Description |
| ----- | ---- | ----------- |
{{range .Fields -}}
| {{if .IsOneof}}[oneof](https://developers.google.com/protocol-buffers/docs/proto3#oneof) {{.OneofDecl}}.{{end}}{{.Name}} | [{{if .IsMap}}map {{else if .Label}}{{.Label}} {{end}}{{.LongType}}](#{{.LongType | lower | replace "." "-"}}) | {{if .Description}}{{nobr .Description}}{{else}}N/A{{end}} |
{{end}}
{{end}}
{{end}}

{{end}}
