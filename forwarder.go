package main

import "context"

type TelegramForwarder struct{}

func NewTelegramForwarder() *TelegramForwarder {
	return &TelegramForwarder{}
}

func (n *TelegramForwarder) Forward(_ context.Context, _ ClientConfig, _ []*Message) error {
	// TODO: implement
	return nil
}
