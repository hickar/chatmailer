package forwarder

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hickar/chatmailer/internal/app/mailer"
)

func TestRenderDefaultTemplate(t *testing.T) {
	msg := &mailer.Message{
		BodyParts: []mailer.BodySegment{},
		Subject:   "Testing templates",
		From:      []mailer.Address{{Address: "hickar@icloud.com"}},
		To:        []mailer.Address{{Address: "hickar@protonmail.ch"}},
		BCC: []mailer.Address{
			{Address: "recipient3@gmail.com"},
			{Address: "recipient4@gmail.com"},
		},
		ReplyTo: []mailer.Address{{Address: "secret.recipient@gmail.com"}},
		Date:    time.Date(1999, time.February, 25, 16, 16, 10, 0, time.Local),
	}

	expectedText := `*From*: [hickar@icloud\.com](mailto://hickar@icloud.com)
*To*: [hickar@protonmail\.ch](mailto://hickar@protonmail.ch)
*Reply To*: [secret\.recipient@gmail\.com](mailto://secret.recipient@gmail.com)
*BCC*: [recipient3@gmail\.com](mailto://recipient3@gmail.com), [recipient4@gmail\.com](mailto://recipient4@gmail.com)
*Subject*: Testing templates
*Date*: Feb 25 1999 16:16:10`

	renderedText, err := renderTemplate(msg, "")
	assert.NoError(t, err)
	assert.Equal(t, expectedText, renderedText)
}
