package forwarder

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hickar/tg-remailer/internal/app/mailer"
)

func TestRenderDefaultTemplate(t *testing.T) {
	msg := &mailer.Message{
		BodyParts: []mailer.BodySegment{},
		Subject:   "Testing templates",
		From:      []string{"hickar@icloud.com"},
		To:        []string{"hickar@protonmail.ch"},
		CC:        []string{"recipient1@gmail.com", "recipient2@gmail.com"},
		BCC:       []string{"recipient3@gmail.com", "recipient4@gmail.com"},
		ReplyTo:   []string{"secret.recipient@gmail.com"},
		Date:      time.Date(1999, time.February, 25, 16, 16, 10, 0, time.Local),
	}

	expectedText := `**From**: [hickar@icloud.com](hickar@icloud.com)
**To**: [hickar@protonmail.ch](hickar@protonmail.ch)
**Reply To**: [secret.recipient@gmail.com](secret.recipient@gmail.com)
**CC**: [recipient1@gmail.com](recipient1@gmail.com), [recipient2@gmail.com](recipient2@gmail.com)
**BCC**: [recipient3@gmail.com](recipient3@gmail.com), [recipient4@gmail.com](recipient4@gmail.com)
**Subject**: Testing templates
**Date**: Feb 25 1999 16:16:10
`

	renderedText, err := renderTemplate(msg, "")
	assert.NoError(t, err)
	assert.Equal(t, expectedText, renderedText)
}

func TestHTMLToMarkdown(t *testing.T) {
	msgHTML := `<!DOCTYPE html/>
<html>
	<head>
		<meta charset="utf-8">
		<title>Sample page</title>
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<link href="styles.css" rel="stylesheet" media="screen">
	</head>
	<body>
		<h1>Header 1</h1>

		<h2>Header 2</h2>
	
		<h3>Header 3</h3>
		
		<h4>Header 4</h4>
		
		<h5>Header 5</h5>
		
		<h6>Header 6</h6>

		<p>First paragraph</p>
		<br/>
		<br/>
		<br/>
		<p>Second paragraph</p>

		<p>First paragraph</p>
		<p>Second paragraph</p>

		<ol>
			<li>First</li>
			<li>Second</li>
			<li>Third</li>
		</ol>

		<a href="/some/link">Link text</a>

		<p><a href="/inner">Link</a> in paragraph</p>

		<p><strong>Bold text</strong></p>
		<p><strong>Bold *text</strong></p>
		<p><i>Italic text</i></p>
		<p><i>Italic _text</i></p>
		<p><ins>Underlined text</ins></p>
		<p><ins>Underlined __text</ins></p>
		<p><strike>Strikethrough text</strike></p>
		<p><strike>Strikethrough ~text</strike></p>

		<p><strong>Bold <em>Italic <u>Underlined</u></em></strong>

		<blockquote>Block quotation started
Block quotation continued
The last line of the block quotation</blockquote>

		<pre><code class="language-python">pre-formatted fixed-width code block written in the Python programming language</code></pre>

		<pre>	pre-formatted fixed-width code block</pre>

		<code>inline fixed-width code</code>
	</body>
</html>
	`
	expected := `Sample page

# Header 1

## Header 2

### Header 3

#### Header 4

##### Header 5

###### Header 6

First paragraph

Second paragraph

First paragraph

Second paragraph

1. First
2. Second
3. Third

[Link text](/some/link)

[Link](/inner) in paragraph

*Bold text*

*Bold \*text*

_Italic text_

_Italic \_text_

__Underlined text__

__Underlined \_\_text__

~Strikethrough text~

~Strikethrough \~text~

*Bold _Italic __Underlined___*

> Block quotation started
> Block quotation continued
> The last line of the block quotation

` + "```" + `python
pre-formatted fixed-width code block written in the Python programming language
` + "```" + `

` + "```" + `
	pre-formatted fixed-width code block
` + "```" + `

` + " `inline fixed-width code`"

	actual, err := htmlToMarkDown(msgHTML)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}
