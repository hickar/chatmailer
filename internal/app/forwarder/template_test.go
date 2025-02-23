package forwarder

import (
	"strings"
	"testing"
	"time"

	"github.com/hickar/chatmailer/internal/app/mailer"

	"github.com/stretchr/testify/assert"
)

func TestRenderDefaultTemplate(t *testing.T) {
	msg := &mailer.Message{
		BodyParts: []mailer.BodySegment{
			{
				MIMEType: "text/plain",
				Body:     strings.NewReader("MUST NOT BE RENDERED"),
			},
			{
				MIMEType: "text/plain",
				Body:     strings.NewReader("MUST NOT BE RENDERED"),
			},
			{
				MIMEType: "text/html",
				Body:     strings.NewReader("single line first part"),
			},
			{
				MIMEType: "text/html",
				Body:     strings.NewReader("multiple<br/>line<br/>second part"),
			},
			{
				MIMEType: "text/html",
				Body:     strings.NewReader("third part"),
			},
			{
				MIMEType: "text/plain",
				Body:     strings.NewReader("MUST NOT BE RENDERED"),
			},
			{
				MIMEType: "text/plain",
				Body:     strings.NewReader("MUST NOT BE RENDERED"),
			},
		},
		Subject: "Testing templates",
		From:    []mailer.Address{{Address: "hickar@icloud.com"}},
		To:      []mailer.Address{{Address: "hickar@protonmail.ch"}},
		BCC: []mailer.Address{
			{Address: "recipient3@gmail.com"},
			{Address: "recipient4@gmail.com"},
		},
		ReplyTo: []mailer.Address{{Address: "secret.recipient@gmail.com"}},
		Date:    time.Date(1999, time.February, 25, 16, 16, 10, 0, time.Local),
	}

	want := `*From*: [hickar@icloud\.com](mailto://hickar@icloud.com)
*To*: [hickar@protonmail\.ch](mailto://hickar@protonmail.ch)
*Reply To*: [secret\.recipient@gmail\.com](mailto://secret.recipient@gmail.com)
*BCC*: [recipient3@gmail\.com](mailto://recipient3@gmail.com), [recipient4@gmail\.com](mailto://recipient4@gmail.com)
*Subject*: Testing templates
*Date*: Feb 25 1999 16:16:10

>single line first part

>multiple
>line
>second part

>third part`

	got, err := renderTemplate(msg, "")
	assert.NoError(t, err)
	assert.Equal(t, want, got)
}
