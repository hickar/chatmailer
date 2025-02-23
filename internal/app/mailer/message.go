package mailer

import (
	"io"
	"time"
)

type Message struct {
	BodyParts   []BodySegment
	Subject     string
	From        []Address
	To          []Address
	CC          []Address
	BCC         []Address
	ReplyTo     []Address
	Date        time.Time
	Mailbox     string
	UID         uint32
	Attachments []Attachment
}

type BodySegment struct {
	MIMEType       string
	MIMETypeParams map[string]string
	Body           io.Reader
	Size           int64
}

type Attachment struct {
	BodySegment
	Filename         string
	CreationDate     time.Time
	ModificationDate time.Time
	ReadDate         time.Time
}

type Mail struct {
	LastUID         uint32
	LastUIDValidity uint32
	Messages        []*Message
}

type Address struct {
	Address string
	Name    string
}
