package main

type FilterFunc func([]string, []Message) ([]Message, error)

func filterMessages(filters []string, messages []Message) ([]Message, error) {
	return messages, nil
}
