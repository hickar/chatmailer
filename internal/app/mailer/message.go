package mailer

import "time"

type Message struct {
	BodyParts   []BodySegment
	Subject     string
	From        []string
	To          []string
	CC          []string
	ReplyTo     []string
	Date        time.Time
	Mailbox     string
	UID         uint32
	Attachments []Attachment
}

type BodySegment struct {
	MIMEType string
	Content  []byte
	Charset  string
}

type Attachment struct {
	Filename string
	MIMEType string
	Body     []byte
}

type MailResponse struct {
	LastUID         uint32
	LastUIDValidity uint32
	Messages        []*Message
}
