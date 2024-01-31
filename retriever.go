package main

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/emersion/go-message/mail"
)

// MailRetriever retrieves mail messages from remote mail server.
type MailRetriever interface {
	GetMail(ClientConfig) (ClientConfig, []*Message, error)
}

type MailSourceFunc func(ClientConfig) (ClientConfig, []*Message, error)

func (m MailSourceFunc) GetMail(mailCfg ClientConfig) (ClientConfig, []*Message, error) {
	return m(mailCfg)
}

func IMAPGetMailFunc(client ClientConfig) (ClientConfig, []*Message, error) {
	c, err := imapclient.DialTLS(client.Address, nil)
	if err != nil {
		return client, nil, fmt.Errorf("failed to dial %q IMAP server: %w", client.Address, err)
	}

	if err = c.Login(client.Login, client.Password).Wait(); err != nil {
		return client, nil, fmt.Errorf("IMAP authentication failed: %w", err)
	}
	mailbox, err := c.Select("inbox", nil).Wait()
	if err != nil {
		return client, nil, fmt.Errorf("failed to select mailbox 'inbox': %w", err)
	}
	client.LastUIDValidity = mailbox.UIDValidity

	if areNoNewMessages(mailbox, client) {
		return client, nil, nil
	}
	if client.LastUID == 0 {
		client.LastUID = uint32(mailbox.UIDNext)
		//	return client, nil, nil
	}

	uidRange := imap.UIDSetNum(imap.UID(client.LastUID - 10))
	fetchCmd := c.Fetch(uidRange, fetchOptions)
	defer func() {
		err = errors.Join(err, fetchCmd.Close())
	}()

	var messages []*Message
	for {
		msg := fetchCmd.Next()
		if msg == nil {
			break
		}

		message, err := processMessage(msg, client)
		if err != nil {
			return client, nil, fmt.Errorf("mail message processing failed: %w", err)
		}

		messages = append(messages, message)
	}

	return client, messages, nil
}

func processMessage(msg *imapclient.FetchMessageData, client ClientConfig) (*Message, error) {
	var (
		bodySection imapclient.FetchItemDataBodySection
		ok          bool
		err         error
	)

	for {
		item := msg.Next()
		if item == nil {
			break
		}
		bodySection, ok = item.(imapclient.FetchItemDataBodySection)
		if ok {
			break
		}
	}
	if !ok {
		return nil, errors.New("message body section is nil")
	}

	mr, err := mail.CreateReader(bodySection.Literal)
	if err != nil {
		return nil, err
	}
	defer mr.Close()

	message := &Message{
		From: mr.Header.Values("From"),
		To:   mr.Header.Values("To"),
		CC:   mr.Header.Values("CC"),
	}
	message.Date, _ = mr.Header.Date()
	message.Subject, _ = mr.Header.Text("Subject")

	// Process the message's parts
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read message part: %w", err)
		}

		switch header := part.Header.(type) {
		case *mail.InlineHeader:
			bodyPart, err := processBodyPart(part)
			if err != nil {
				return nil, fmt.Errorf("body segment parsing failed: %w", err)
			}

			message.BodyParts = append(message.BodyParts, bodyPart)
		case *mail.AttachmentHeader:
			if !client.IncludeAttachments {
				break
			}

			attachment, err := processAttachment(part, header)
			if err != nil {
				return nil, fmt.Errorf("attachment parsing failed: %w", err)
			}

			message.Attachments = append(message.Attachments, attachment)
		}
	}

	return message, nil
}

func processBodyPart(part *mail.Part) (BodySegment, error) {
	var bodySegment BodySegment
	bodySegment.Content, _ = io.ReadAll(part.Body)

	headerValue := part.Header.Get("Content-type")
	mimeType, charset, ok := parseContentTypeHeader(headerValue)
	if !ok {
		return bodySegment, fmt.Errorf("failed to parse 'Content-type' header with value %q", headerValue)
	}

	bodySegment.MIMEType = mimeType
	bodySegment.Charset = charset

	return bodySegment, nil
}

func processAttachment(part *mail.Part, header *mail.AttachmentHeader) (Attachment, error) {
	// TODO: implement later
	return Attachment{}, nil
}

func parseContentTypeHeader(header string) (string, string, bool) {
	var mimeType, charset string

	segs := strings.Split(header, ";")
	switch len(segs) {
	case 0:
		return "", "", false
	case 1:
		mimeType = segs[0]
	case 2:
		mimeType, charset = segs[0], segs[1]
	}

	return mimeType, charset, true
}

func areNoNewMessages(mailbox *imap.SelectData, client ClientConfig) bool {
	return client.LastUIDValidity == mailbox.UIDValidity &&
		client.LastUID+1 == uint32(mailbox.UIDNext)
}

var fetchOptions = &imap.FetchOptions{
	Envelope:     true,
	Flags:        true,
	InternalDate: true,
	RFC822Size:   true,
	UID:          true,
	BodySection:  []*imap.FetchItemBodySection{{}},
	ModSeq:       true,
}
