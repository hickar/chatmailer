package retriever

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"strings"
	"time"

	"github.com/hickar/chatmailer/internal/app/config"
	"github.com/hickar/chatmailer/internal/app/mailer"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/emersion/go-message"
	"github.com/emersion/go-message/charset"
	"github.com/emersion/go-message/mail"
)

type ImapDialer interface {
	DialTLS(address string, options *imapclient.Options) (*imapclient.Client, error)
}

type ImapDialerFunc func(string, *imapclient.Options) (*imapclient.Client, error)

func (f ImapDialerFunc) DialTLS(address string, options *imapclient.Options) (*imapclient.Client, error) {
	return f(address, options)
}

type imapRetriever struct {
	dialer ImapDialer
	logger *slog.Logger
}

func NewIMAPRetriever(dialer ImapDialer, logger *slog.Logger) *imapRetriever {
	return &imapRetriever{
		dialer: dialer,
		logger: logger,
	}
}

// GetMail retrieves email messages from an IMAP server for specificied client.
//
// Execution flow:
// 1. Connect to the IMAP server using TLS.
// 2. Authenticate with the provided login credentials.
// 3. Select the "inbox" mailbox (optionally marking messages as seen).
// 4. Check if there are new messages based on UID validity and UIDNext comparison.
// 5. If there are new messages:
//   - Retrieve capabilities to determine if extended search is supported.
//   - If filters are provided and extended search is supported:
//   - Build a search criteria based on the filters and the client's LastUIDNext.
//   - Perform a search on the server to get the UIDs of matching messages.
//   - Otherwise, fetch all messages since the client's LastUIDNext (inclusive).
//
// 6. For each message:
//   - Extract the UID, sender, recipients, CC recipients, date, and subject.
//   - Process each part of the message (text body or attachment):
//   - For text parts:
//   - Read the content and determine MIME type and character set.
//   - Add the body segment to the message.
//   - For attachments (if inclusion is enabled):
//   - Parse the attachment information (not yet implemented).
//   - Add the attachment to the message (not yet implemented).
//
// 7. Return the retrieved messages and any encountered errors.
// Lacks of appropriate error and behaviour handling, need to handle such cases:
//   - Dial TLS failure
//   - Login failure
//   - Read mainbox failure
//   - LastUIDNext === 0 case
//   - Capability error
//   - Search builder and Lexer errors
//   - Message, Headers and Attachments errors
func (r *imapRetriever) GetMail(ctx context.Context, cfg config.ClientConfig) (mailer.Mail, error) {
	// 1. TODO(hickar): pass context.Context and handle cancellation with it.
	// 2. TODO(hickar): handle IMAP connection reuse.
	// 3. TODO(hickar): consider IMAP IDLE command to receive
	// server-side notifications on new messages arrival.
	// 4. TODO(hickar): wrap error within custom error types
	// to utilize different error handling strategies on the upper level.
	var mail mailer.Mail

	client, err := r.dialer.DialTLS(cfg.Address, &imapclient.Options{
		DebugWriter:           nil,
		UnilateralDataHandler: &imapclient.UnilateralDataHandler{},
		WordDecoder:           &mime.WordDecoder{CharsetReader: charset.Reader},
	})
	if err != nil {
		return mail, fmt.Errorf("dial TLS: %w", err)
	}

	if err = client.Login(cfg.Login, cfg.Password).Wait(); err != nil {
		return mail, fmt.Errorf("login: %w", err)
	}

	mailbox, err := client.Select("inbox", &imap.SelectOptions{
		ReadOnly: cfg.MarkAsSeen,
	}).Wait()
	if err != nil {
		return mail, fmt.Errorf("select: %w", err)
	}
	mail.LastUIDValidity = mailbox.UIDValidity
	mail.LastUID = uint32(mailbox.UIDNext)

	if areNoNewMessages(mailbox, cfg) {
		return mail, nil
	}
	if cfg.LastUIDNext == 0 {
		return mail, nil
	}

	capabilities, err := client.Capability().Wait()
	if err != nil {
		return mail, fmt.Errorf("get capabilities: %w", err)
	}

	uids := imap.UIDSet{imap.UIDRange{
		Start: imap.UID(cfg.LastUIDNext),
		Stop:  imap.UID(mail.LastUID),
	}}
	if len(cfg.Filters) > 0 && capabilities.Has(imap.CapESearch) {
		uids, err = getUIDsByCriteria(client, cfg)
		if err != nil {
			return mail, fmt.Errorf("get UID set by search criteria: %w", err)
		}
	}

	fetchCmd := client.Fetch(uids, fetchOptions)
	defer func() {
		_ = fetchCmd.Close()
	}()

	for {
		msg := fetchCmd.Next()
		if msg == nil {
			break
		}

		var message *mailer.Message
		message, err = parseMessage(msg, cfg)
		if err != nil {
			return mail, fmt.Errorf("process message: %w", err)
		}
		// TODO(hickar): handle message filtering in case of remote IMAP server inability
		// to filter messages based on sent search criteria
		//
		// if !capabilities.Has(imap.CapESearch) {
		// 	...
		// }

		mail.Messages = append(mail.Messages, message)
	}

	return mail, err
}

func getUIDsByCriteria(c *imapclient.Client, client config.ClientConfig) (imap.UIDSet, error) {
	criteria, err := buildSearchCriteria(client.Filters, client.LastUIDNext)
	if err != nil {
		return nil, fmt.Errorf("build search criteria: %w", err)
	}

	cmd, err := c.UIDSearch(criteria, nil).Wait()
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}

	uids := imap.UIDSetNum(cmd.AllUIDs()...)
	return uids, nil
}

func parseMessage(msg *imapclient.FetchMessageData, client config.ClientConfig) (*mailer.Message, error) {
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
		return nil, fmt.Errorf("create reader: %w", err)
	}
	defer func() {
		_ = mr.Close()
	}()

	message := &mailer.Message{
		UID:  uint32(uidSection.UID),
		From: parseAddress(mr.Header, "From"),
		To:   parseAddress(mr.Header, "To"),
		CC:   parseAddress(mr.Header, "CC"),
		BCC:  parseAddress(mr.Header, "BCC"),
	}
	message.Date, _ = mr.Header.Date()
	message.Subject, _ = mr.Header.Text("Subject")

	// Process the message's parts
	for {
		part, err := mr.NextPart()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read message part: %w", err)
		}

		switch header := part.Header.(type) {
		case *mail.InlineHeader:
			bodyPart, err := parseBodyPart(part, header.Header)
			if err != nil {
				return nil, fmt.Errorf("body segment parsing: %w", err)
			}

			message.BodyParts = append(message.BodyParts, bodyPart)
		case *mail.AttachmentHeader:
			if !client.IncludeAttachments {
				break
			}

			attachment, err := parseAttachment(part, header)
			if err != nil {
				return nil, fmt.Errorf("attachment parsing: %w", err)
			}

			if attachment.Size > int64(client.MaximumAttachmentsSize) {
				break
			}

			message.Attachments = append(message.Attachments, attachment)
		}
	}

	return message, nil
}

func parseAttachment(part *mail.Part, header *mail.AttachmentHeader) (mailer.Attachment, error) {
	var attachment mailer.Attachment
	var err error

	attachment.BodySegment, err = parseBodyPart(part, header.Header)
	if err != nil {
		return attachment, fmt.Errorf("parse body part: %w", err)
	}

	params := attachment.MIMETypeParams
	if v, ok := params["creation-date"]; ok {
		attachment.CreationDate, err = time.Parse(time.RFC822, v)
		if err != nil {
			return attachment, fmt.Errorf("parse 'creation-date': %w", err)
		}
	}

	if v, ok := params["modification-date"]; ok {
		attachment.ModificationDate, err = time.Parse(time.RFC822, v)
		if err != nil {
			return attachment, fmt.Errorf("parse 'modification-date': %w", err)
		}
	}

	if v, ok := params["read-date"]; ok {
		attachment.ReadDate, err = time.Parse(time.RFC822, v)
		if err != nil {
			return attachment, fmt.Errorf("parse 'read-date': %w", err)
		}
	}

	attachment.Filename, err = header.Filename()
	if err != nil {
		return attachment, fmt.Errorf("get filename: %w", err)
	}

	return attachment, nil
}

func parseBodyPart(part *mail.Part, header message.Header) (mailer.BodySegment, error) {
	var segment mailer.BodySegment
	var buf bytes.Buffer
	var err error

	segment.Size, err = buf.ReadFrom(part.Body)
	if err != nil {
		return segment, fmt.Errorf("read from: %w", err)
	}
	segment.Body = &buf

	segment.MIMEType, segment.MIMETypeParams, err = header.ContentType()
	if err != nil {
		return segment, fmt.Errorf("get 'Content-Type': %w", err)
	}

	return segment, nil
}

func parseAddress(header mail.Header, addressListName string) []mailer.Address {
	addrList, _ := header.AddressList(addressListName)
	addrs := make([]mailer.Address, 0, len(addrList))

	for _, addr := range addrList {
		addrs = append(addrs, mailer.Address{
			Name:    addr.Name,
			Address: addr.Address,
		})
	}

	return addrs
}

func areNoNewMessages(mailbox *imap.SelectData, client config.ClientConfig) bool {
	return client.LastUIDValidity == mailbox.UIDValidity &&
		client.LastUIDNext == uint32(mailbox.UIDNext)
}

func buildSearchCriteria(filters []string, lastClientUIDNext uint32) (*imap.SearchCriteria, error) {
	uids := []imap.UIDSet{{imap.UIDRange{
		Start: imap.UID(lastClientUIDNext),
	}}}
	criteria := imap.SearchCriteria{UID: uids}

	for _, filterExpr := range filters {
		if strings.TrimSpace(filterExpr) == "" {
			continue
		}

		newCriteria, err := ParseFilter(filterExpr)
		if err != nil {
			return nil, fmt.Errorf("parse filter expression %q: %w", filterExpr, err)
		}

		criteria.And(newCriteria)
	}

	return &criteria, nil
}

func setUIDs(criteria *imap.SearchCriteria, uids []imap.UIDSet) {
	if criteria == nil {
		return
	}

	criteria.UID = uids
	for i := range criteria.Or {
		setUIDs(&criteria.Or[i][0], uids)
		setUIDs(&criteria.Or[i][1], uids)
	}
	for i := range criteria.Not {
		setUIDs(&criteria.Not[i], uids)
	}
}

var fetchOptions = &imap.FetchOptions{
	Envelope:     true,
	Flags:        true,
	InternalDate: true,
	RFC822Size:   true,
	UID:          true,
	BodySection:  []*imap.FetchItemBodySection{{Peek: true}},
	ModSeq:       true,
}
