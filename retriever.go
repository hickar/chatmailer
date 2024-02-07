package main

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"strings"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/emersion/go-message/charset"
	"github.com/emersion/go-message/mail"
)

// MailRetriever retrieves mail messages from remote mail server.
type MailRetriever interface {
	GetMail(ClientConfig) (MailResponse, error)
}

type MailSourceFunc func(ClientConfig) (MailResponse, error)

func (m MailSourceFunc) GetMail(mailCfg ClientConfig) (MailResponse, error) {
	return m(mailCfg)
}

type MailResponse struct {
	LastUID         uint32
	LastUIDValidity uint32
	Messages        []*Message
}

func IMAPGetMailFunc(client ClientConfig) (MailResponse, error) {
	var mailResp MailResponse

	c, err := imapclient.DialTLS(client.Address, &imapclient.Options{
		DebugWriter:           nil,
		UnilateralDataHandler: &imapclient.UnilateralDataHandler{},
		WordDecoder:           &mime.WordDecoder{CharsetReader: charset.Reader},
	})
	if err != nil {
		return mailResp, err
	}

	if err = c.Login(client.Login, client.Password).Wait(); err != nil {
		return mailResp, err
	}

	mailbox, err := c.Select("inbox", nil).Wait()
	if err != nil {
		return mailResp, err
	}
	mailResp.LastUIDValidity = mailbox.UIDValidity
	mailResp.LastUID = uint32(mailbox.UIDNext)

	if areNoNewMessages(mailbox, client) {
		return mailResp, nil
	}
	if client.LastUIDNext == 0 {
		return mailResp, nil
	}

	capabilities, err := c.Capability().Wait()
	if err != nil {
		return mailResp, err
	}

	var uidSet imap.UIDSet
	if len(client.Filters) > 0 && capabilities.Has(imap.CapESearch) {
		uidSet, err = getUIDSetBySearchCriteria(c, client)
		if err != nil {
			return mailResp, err
		}
	} else {
		uidSet = imap.UIDSet{imap.UIDRange{
			Start: imap.UID(mailResp.LastUID) - 10,
			Stop:  imap.UID(mailResp.LastUID),
		}}
	}

	fetchCmd := c.Fetch(uidSet, fetchOptions)
	defer func() {
		err = errors.Join(err, fetchCmd.Close())
	}()

	for {
		msg := fetchCmd.Next()
		if msg == nil {
			break
		}

		message, err := processMessage(msg, client)
		if err != nil {
			return mailResp, err
		}
		// if !capabilities.Has(imap.CapESearch) {
		// 	// TODO: handle message filtering
		// }

		mailResp.Messages = append(mailResp.Messages, message)
	}

	return mailResp, err
}

func getUIDSetBySearchCriteria(c *imapclient.Client, client ClientConfig) (imap.UIDSet, error) {
	searchCriteria, err := buildSearchCriteria(client.Filters, client.LastUIDNext)
	if err != nil {
		return nil, fmt.Errorf("search criteria parsing failed: %w", err)
	}

	searchCmd, err := c.Search(searchCriteria, &imap.SearchOptions{
		ReturnMin:   false,
		ReturnMax:   false,
		ReturnAll:   false,
		ReturnCount: false,
		ReturnSave:  false,
	}).Wait()
	if err != nil {
		return nil, err
	}

	uidSet := imap.UIDSetNum(searchCmd.AllUIDs()...)
	return uidSet, nil
}

func processMessage(msg *imapclient.FetchMessageData, client ClientConfig) (*Message, error) {
	var (
		uidSection  imapclient.FetchItemDataUID
		bodySection imapclient.FetchItemDataBodySection
		ok          bool
	)

	item := msg.Next()
	if item == nil {
		return nil, errors.New("first message part is nil")
	}
	uidSection, ok = item.(imapclient.FetchItemDataUID)
	if !ok {
		return nil, errors.New("first message part is not UID section")
	}

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
		UID:  uint32(uidSection.UID),
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
		client.LastUIDNext == uint32(mailbox.UIDNext)
}

func buildSearchCriteria(filters []string, lastClientUIDNext uint32) (*imap.SearchCriteria, error) {
	var searchCriteria *imap.SearchCriteria

	for _, filterExpr := range filters {
		if strings.TrimSpace(filterExpr) == "" {
			continue
		}

		newCriteria, err := parseFilter(filterExpr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse filter expression %q: %w", filterExpr, err)
		}

		if searchCriteria == nil {
			searchCriteria = newCriteria
			continue
		}

		searchCriteria.And(newCriteria)
	}
	if searchCriteria != nil {
		searchCriteria.And(&imap.SearchCriteria{
			UID: []imap.UIDSet{{imap.UIDRange{
				Start: imap.UID(lastClientUIDNext),
			}}},
		})
	}

	return searchCriteria, nil
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
