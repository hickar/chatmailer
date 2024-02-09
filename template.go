package main

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	md "github.com/JohannesKaufmann/html-to-markdown"
)

const defaultMessageTemplateString = `
{{- define "addresses" -}}
	{{- with . -}}
		{{- range $idx, $address := . -}}
			{{- if eq $idx 0 -}}
				{{- printf "[%s](%s)" $address $address -}}
			{{- else -}}
				{{- printf ", [%s](%s)" $address $address -}}
			{{- end -}}
		{{- end -}}
	{{- end -}}
{{- end -}}

**From**: {{template "addresses" .From}}
**To**: {{template "addresses" .To}}
**Reply To**: {{template "addresses" .ReplyTo}}
**CC**: {{template "addresses" .CC}}
{{with .Subject}}**Subject**: {{.}}{{end}}
{{with .Date}}**Date**: {{ .Format "Jan 02 2006 15:04:05" }}{{end}}
`

var (
	templateFuncMap = template.FuncMap{
		"join":     strings.Join,
		"markdown": convertToMarkDown,
	}
	defaultTemplate = template.Must(template.New("").Funcs(templateFuncMap).Parse(defaultMessageTemplateString))
)

func renderTemplate(message *Message, templateStr string) (string, error) {
	var buf bytes.Buffer
	var err error

	if templateStr == "" {
		err = defaultTemplate.Execute(&buf, message)
		return buf.String(), err
	}

	var tmpl *template.Template
	tmpl, err = template.New("custom").Funcs(templateFuncMap).Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("custom template creation error: %w", err)
	}

	err = tmpl.Execute(&buf, message)
	return buf.String(), err
}

func convertToMarkDown(html string) (string, error) {
	c := md.NewConverter("", true, nil)
	return c.ConvertString(html)
}
