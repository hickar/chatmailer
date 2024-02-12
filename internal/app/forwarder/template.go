package forwarder

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"

	"github.com/hickar/tg-remailer/internal/app/mailer"
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
**BCC**: {{template "addresses" .BCC}}
{{with .Subject}}**Subject**: {{.}}{{end}}
{{with .Date}}**Date**: {{ .Format "Jan 02 2006 15:04:05" }}{{end}}
{{- range $part := .BodyParts  -}}
	{{- if eq $part.MIMETYPE "text/html" -}}
		{{- htmltomarkdown $part.Content -}}
	{{- else if eq $part.MIMETYPE "text/plain" -}}
		{{- $part.Content -}}
	{{- end -}}
{{- end -}}
`

var (
	templateFuncMap = template.FuncMap{
		"join":           strings.Join,
		"htmltomarkdown": htmlToMarkDown,
	}
	defaultTemplate = template.Must(template.New("").Funcs(templateFuncMap).Parse(defaultMessageTemplateString))
)

func renderTemplate(message *mailer.Message, templateStr string) (string, error) {
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

func htmlToMarkDown(html string) (string, error) {
	c := md.NewConverter("", true, &md.Options{
		// Fence:              "",
		// EmDelimiter:        "",
		StrongDelimiter: "*",
	}).AddRules(extendedMDRules...)
	return c.ConvertString(html)
}

var extendedMDRules = []md.Rule{
	// Rule for underlined text conversion
	{
		Filter: []string{"u", "ins"},
		Replacement: func(content string, _ *goquery.Selection, _ *md.Options) *string {
			return md.String("__" + content + "__")
		},
	},
	// Rule for striked through text conversion
	{
		Filter: []string{"s", "strike", "del"},
		Replacement: func(content string, _ *goquery.Selection, _ *md.Options) *string {
			content = strings.ReplaceAll(content, "~", "\\~")
			return md.String("~" + content + "~")
		},
	},
}
