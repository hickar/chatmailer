package forwarder

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"html"
	"io"
	"strings"
	"text/template"

	"github.com/hickar/chatmailer/internal/app/mailer"

	"jaytaylor.com/html2text"
)

const defaultTemplateContent = `
{{- define "addresses" -}}
	{{- range $idx, $address := . -}}
		{{ $escapedAddress := escapeMarkdown $address.Address }}

		{{- if ne $idx 0 -}}
			{{- printf ", " -}}
		{{- end -}}

		{{- printf "[%s](mailto://%s)" $escapedAddress $address.Address -}}
	{{- end -}}
{{- end -}}

{{ define "html-body" }}{{ range $part := . }}{{ if eq $part.MIMEType "text/html" }}
{{ htmlstring $part.Body | escapeMarkdown | quoteMarkdown }}
{{ end }}{{ end }}{{ end }}

{{ define "text-body" }}{{ range $part := . }}{{ if eq $part.MIMEType "text/plain" }}
{{ htmlstring $part.Body | escapeMarkdown | quoteMarkdown }}
{{ end }}{{ end }}{{ end }}

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
{{ end }}

{{- $hasParts := gt (len .BodyParts) 0 -}}
{{- $hasHTMLParts := containsMIMEType .BodyParts "text/html" -}}
{{- $hasTextParts := containsMIMEType .BodyParts "text/plain" -}}

{{ if and $hasParts $hasHTMLParts }}{{ template "html-body" .BodyParts }}
{{ else if and $hasParts $hasTextParts }}{{ template "text-body" .BodyParts }}
{{ else -}}TEXT MESSAGE CAN NOT BE REPRESENTED
{{ end -}}`

var (
	defaultTemplateFuncs = template.FuncMap{
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
		"quoteMarkdown":    quoteMarkdown,
	}
	defaultTemplateName = "default"
	defaultTemplate     = template.Must(
		template.
			New(defaultTemplateName).
			Funcs(defaultTemplateFuncs).
			Parse(defaultTemplateContent),
	)
)

func renderTemplate(message *mailer.Message, templateContent string) (string, error) {
	var buf bytes.Buffer

	switch {
	case templateContent == "":
		if err := defaultTemplate.Execute(&buf, message); err != nil {
			return "", fmt.Errorf("default template rendering: %w", err)
		}
	default:
		tmpl, err := template.
			New(templateHash(templateContent)).
			Funcs(defaultTemplateFuncs).
			Parse(templateContent)
		if err != nil {
			return "", fmt.Errorf("custom template parsing: %w", err)
		}

		if err = tmpl.Execute(&buf, message); err != nil {
			return "", fmt.Errorf("custom template rendering: %w", err)
		}
	}

	// I don't know who the fuck designed standard Go template engine syntax,
	// but I was not able to think of a way to remove newline at the very end.
	return strings.TrimSpace(buf.String()), nil
}

func templateHash(s string) string {
	return string(fnv.New32().Sum([]byte(s)))
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

	return false
}

// quoteMarkdown wraps provided text in markdown 'quote' block
// by prepending each line with '>' symbol.
func quoteMarkdown(s string) string {
	br := bytes.NewBufferString(s)
	bw := bytes.NewBuffer(make([]byte, 0, len(s)))

	for {
		line, err := br.ReadString('\n')
		if line == "" && err != nil {
			break
		}

		_, _ = fmt.Fprintf(bw, ">%s", line)
	}

	return bw.String()
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
