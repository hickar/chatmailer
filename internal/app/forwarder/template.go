package forwarder

import (
	"bytes"
	"fmt"
	"html"
	"io"
	"strings"
	"text/template"

	"github.com/hickar/chatmailer/internal/app/mailer"

	"jaytaylor.com/html2text"
)

const defaultMessageTemplateString = `
{{- define "addresses" -}}
	{{- range $idx, $address := . -}}
		{{ $escapedAddress := escapeMarkdown $address.Address }}

		{{- if ne $idx 0 -}}
			{{- printf ", " -}}
		{{- end -}}

		{{- printf "[%s](mailto://%s)" $escapedAddress $address.Address -}}
	{{- end -}}
{{- end -}}

{{- if .From }}*From*: {{ template "addresses" .From }}
{{ end }}
{{- if .To }}*To*: {{ template "addresses" .To }}
{{ end }}
{{- if .ReplyTo }}*Reply To*: {{ template "addresses" .ReplyTo }}
{{ end }}
{{- if .CC }}*CC*: {{ template "addresses" .CC }}
{{ end }}
{{- if .BCC }}*BCC*: {{ template "addresses" .BCC }}
{{ end }}
{{- if .Subject }}*Subject*: {{ escapeMarkdown .Subject }}
{{ end }}
{{- if .Date }}*Date*: {{ .Date.Format "Jan 02 2006 15:04:05" }}
{{- end }}`

var templateFuncs = template.FuncMap{
	"escapeMarkdown":   escapeMarkdown,
	"escapeHTML":       escapeHTML,
	"join":             strings.Join,
	"replace":          strings.Replace,
	"replaceAll":       strings.ReplaceAll,
	"upper":            strings.ToUpper,
	"lower":            strings.ToLower,
	"contains":         strings.Contains,
	"trim":             strings.Trim,
	"trimSpace":        strings.TrimSpace,
	"bytestring":       bytesToString,
	"htmlstring":       htmlToText,
	"containsMIMEType": containsMIMEType,
}

var defaultTemplate = template.Must(template.New("").Funcs(templateFuncs).Parse(defaultMessageTemplateString))

func renderTemplate(message *mailer.Message, templateStr string) (string, error) {
	var buf bytes.Buffer

	if templateStr == "" {
		if err := defaultTemplate.Execute(&buf, message); err != nil {
			return "", fmt.Errorf("default template rendering: %w", err)
		}

		return buf.String(), nil
	}

	tmpl, err := template.New("custom").Funcs(templateFuncs).Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("custom template parsing: %w", err)
	}

	if err = tmpl.Execute(&buf, message); err != nil {
		return "", fmt.Errorf("custom template rendering: %w", err)
	}

	return buf.String(), nil
}

func escapeMarkdown(s string) string {
	return escapeCharacters(s, markdownSpecialChars)
}

func escapeHTML(s string) string {
	return html.EscapeString(s)
}

func bytesToString(payload any) string {
	switch v := payload.(type) {
	case []byte:
		return string(v)

	case io.Reader:
		b, err := io.ReadAll(v)
		if err != nil {
			break
		}

		return string(b)
	}

	return ""
}

var defaultHTMLToTextOpts = html2text.Options{TextOnly: true}

func htmlToText(payload any) string {
	var output string

	switch v := payload.(type) {
	case string:
		output, _ = html2text.FromString(v, defaultHTMLToTextOpts)
	case io.Reader:
		output, _ = html2text.FromReader(v, defaultHTMLToTextOpts)
	}

	return output
}

func containsMIMEType(parts []mailer.BodySegment, mimeType string) bool {
	for _, part := range parts {
		if part.MIMEType == mimeType {
			return true
		}
	}

	return true
}

func escapeCharacters(s string, charMap map[rune]struct{}) string {
	var (
		buf strings.Builder
		ok  bool
	)

	buf.Grow(len(s))

	for _, c := range s {
		if _, ok = charMap[c]; ok {
			buf.WriteString(`\`)
		}

		buf.WriteRune(c)
	}

	return buf.String()
}

var markdownSpecialChars = map[rune]struct{}{
	'_': {},
	'*': {},
	'[': {},
	']': {},
	'(': {},
	')': {},
	'~': {},
	'`': {},
	'>': {},
	'#': {},
	'+': {},
	'-': {},
	'=': {},
	'|': {},
	'{': {},
	'}': {},
	'.': {},
	'!': {},
	'"': {},
}
