/*
Copyright 2021 The terraform-docs Authors.

Licensed under the MIT license (the "License"); you may not
use this file except in compliance with the License.

You may obtain a copy of the License at the LICENSE file in
the root directory of this source tree.
*/

package format

import (
	gotemplate "text/template"

	"github.com/terraform-docs/terraform-docs/internal/print"
	"github.com/terraform-docs/terraform-docs/internal/template"
	"github.com/terraform-docs/terraform-docs/internal/terraform"
)

const (
	documentHeaderTpl = `
	{{- if .Settings.ShowHeader -}}
		{{- with .Module.Header -}}
			{{ sanitizeHeader . }}
			{{ printf "\n" }}
		{{- end -}}
	{{ end -}}
	`
	documentResourcesTpl = `
	{{- if .Settings.ShowResources -}}
		{{ indent 0 "#" }} Resources
		{{ if not .Module.Resources }}
			No resources.
		{{ else }}
			The following resources are used by this module:
			{{ range .Module.Resources }}
				{{ if eq (len .URL) 0 }}
				- {{ .FullType }}
				{{- else -}}
				- [{{ .FullType }}]({{ .URL }})
				{{- end }}
			{{- end }}
		{{ end }}
	{{ end -}}
	`

	documentRequirementsTpl = `
	{{- if .Settings.ShowRequirements -}}
		{{ indent 0 "#" }} Requirements
		{{ if not .Module.Requirements }}
			No requirements.
		{{ else }}
			The following requirements are needed by this module:
			{{- range .Module.Requirements }}
				{{ $version := ternary (tostring .Version) (printf " (%s)" .Version) "" }}
				- {{ name .Name }}{{ $version }}
			{{- end }}
		{{ end }}
	{{ end -}}
	`

	documentProvidersTpl = `
	{{- if .Settings.ShowProviders -}}
		{{ indent 0 "#" }} Providers
		{{ if not .Module.Providers }}
			No provider.
		{{ else }}
			The following providers are used by this module:
			{{- range .Module.Providers }}
				{{ $version := ternary (tostring .Version) (printf " (%s)" .Version) "" }}
				- {{ name .FullName }}{{ $version }}
			{{- end }}
		{{ end }}
	{{ end -}}
	`

	documentInputsTpl = `
	{{- if .Settings.ShowInputs -}}
		{{- if .Settings.ShowRequired -}}
			{{ indent 0 "#" }} Required Inputs
			{{ if not .Module.RequiredInputs }}
				No required input.
			{{ else }}
				The following input variables are required:
				{{- range .Module.RequiredInputs }}
					{{ template "input" . }}
				{{- end }}
			{{- end }}
			{{ indent 0 "#" }} Optional Inputs
			{{ if not .Module.OptionalInputs }}
				No optional input.
			{{ else }}
				The following input variables are optional (have default values):
				{{- range .Module.OptionalInputs }}
					{{ template "input" . }}
				{{- end }}
			{{ end }}
		{{ else -}}
			{{ indent 0 "#" }} Inputs
			{{ if not .Module.Inputs }}
				No input.
			{{ else }}
				The following input variables are supported:
				{{- range .Module.Inputs }}
					{{ template "input" . }}
				{{- end }}
			{{ end }}
		{{- end }}
	{{ end -}}
	`

	documentInputTpl = `
	{{ printf "\n" }}
	{{ indent 1 "#" }} {{ name .Name }}

	Description: {{ tostring .Description | sanitizeDoc }}

	Type: {{ tostring .Type | type }}

	{{ if or .HasDefault (not isRequired) }}
		Default: {{ default "n/a" .GetValue | value }}
	{{- end }}
	`

	documentOutputsTpl = `
	{{- if .Settings.ShowOutputs -}}
		{{ indent 0 "#" }} Outputs
		{{ if not .Module.Outputs }}
			No output.
		{{ else }}
			The following outputs are exported:
			{{- range .Module.Outputs }}

				{{ indent 1 "#" }} {{ name .Name }}

				Description: {{ tostring .Description | sanitizeDoc }}

				{{ if $.Settings.OutputValues }}
					{{- $sensitive := ternary .Sensitive "<sensitive>" .GetValue -}}
					Value: {{ value $sensitive | sanitizeDoc }}

					{{ if $.Settings.ShowSensitivity -}}
						Sensitive: {{ ternary (.Sensitive) "yes" "no" }}
					{{- end }}
				{{ end }}
			{{ end }}
		{{ end }}
	{{ end -}}
	`

	documentTpl = `
	{{- template "header" . -}}
	{{- template "requirements" . -}}
	{{- template "providers" . -}}
	{{- template "resources" . -}}
	{{- template "inputs" . -}}
	{{- template "outputs" . -}}
	`
)

// MarkdownDocument represents Markdown Document format.
type MarkdownDocument struct {
	template *template.Template
}

// NewMarkdownDocument returns new instance of Document.
func NewMarkdownDocument(settings *print.Settings) print.Engine {
	tt := template.New(settings, &template.Item{
		Name: "document",
		Text: documentTpl,
	}, &template.Item{
		Name: "header",
		Text: documentHeaderTpl,
	}, &template.Item{
		Name: "requirements",
		Text: documentRequirementsTpl,
	}, &template.Item{
		Name: "providers",
		Text: documentProvidersTpl,
	}, &template.Item{
		Name: "resources",
		Text: documentResourcesTpl,
	}, &template.Item{
		Name: "inputs",
		Text: documentInputsTpl,
	}, &template.Item{
		Name: "input",
		Text: documentInputTpl,
	}, &template.Item{
		Name: "outputs",
		Text: documentOutputsTpl,
	})
	tt.CustomFunc(gotemplate.FuncMap{
		"type": func(t string) string {
			result, extraline := printFencedCodeBlock(t, "hcl")
			if !extraline {
				result += "\n"
			}
			return result
		},
		"value": func(v string) string {
			if v == "n/a" {
				return v
			}
			result, extraline := printFencedCodeBlock(v, "json")
			if !extraline {
				result += "\n"
			}
			return result
		},
		"isRequired": func() bool {
			return settings.ShowRequired
		},
	})
	return &MarkdownDocument{
		template: tt,
	}
}

// Print a Terraform module as Markdown document.
func (d *MarkdownDocument) Print(module *terraform.Module, settings *print.Settings) (string, error) {
	rendered, err := d.template.Render(module)
	if err != nil {
		return "", err
	}
	return sanitize(rendered), nil
}

func init() {
	register(map[string]initializerFn{
		"markdown document": NewMarkdownDocument,
		"markdown doc":      NewMarkdownDocument,
		"md document":       NewMarkdownDocument,
		"md doc":            NewMarkdownDocument,
	})
}
