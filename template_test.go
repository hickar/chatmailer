package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRenderDefaultTemplate(t *testing.T) {
	msg := &Message{
		BodyParts: []BodySegment{},
		Subject:   "Testing templates",
		From:      []string{"hickar@icloud.com"},
		To:        []string{"hickar@protonmail.ch"},
		CC:        []string{"recipient1@gmail.com", "recipient2@gmail.com"},
		ReplyTo:   []string{"secret.recipient@gmail.com"},
		Date:      time.Date(1999, time.February, 25, 16, 16, 10, 0, time.Local),
	}

	expectedText := `**From**: [hickar@icloud.com](hickar@icloud.com)
**To**: [hickar@protonmail.ch](hickar@protonmail.ch)
**Reply To**: [secret.recipient@gmail.com](secret.recipient@gmail.com)
**CC**: [recipient1@gmail.com](recipient1@gmail.com), [recipient2@gmail.com](recipient2@gmail.com)
**Subject**: Testing templates
**Date**: Feb 25 1999 16:16:10
`

	renderedText, err := renderTemplate(msg, "")
	assert.NoError(t, err)
	assert.Equal(t, expectedText, renderedText)
}
