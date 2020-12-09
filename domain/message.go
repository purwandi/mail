package domain

import (
	"time"
)

// Message ...
type Message struct {
	ID          string       `json:"id"`
	Subject     string       `json:"subject"`
	Sender      string       `json:"sender"`
	From        []Contact    `json:"from"`
	To          []Contact    `json:"to"`
	HTMLBody    string       `json:"html_body"`
	TextBody    string       `json:"text_body"`
	Attachments []Attachment `json:"attachments"`
	Date        time.Time    `json:"date"`
}
