package repository

import (
	"time"

	"github.com/purwandi/mail/domain"
	"github.com/segmentio/ksuid"
)

// MessageInMemoryStore ...
type MessageInMemoryStore struct {
	mapMessage map[string]domain.Message
}

// NewMessageInMemoryStore ...
func NewMessageInMemoryStore() *MessageInMemoryStore {
	m1 := ksuid.New().String()
	m2 := ksuid.New().String()

	return &MessageInMemoryStore{
		mapMessage: map[string]domain.Message{
			m1: {
				ID:      m1,
				Subject: "7 Quotes by Albert Einstein That Will Change How You Think ",
				Sender:  "noreply@medium.com",
				From: []domain.Contact{
					{Email: "noreply@medium.com", Name: "Medium"},
				},
				To: []domain.Contact{
					{Email: "foo@bar.com", Name: "Fooabar"},
				},
				TextBody: "We’ve almost made it to 2021. Have you built any particularly cool projects this year? We’d love to hear about them.",
				HTMLBody: "<p>hello world</p>",
				Attachments: []domain.Attachment{
					{ID: "1lNigjr8fsntbweehfBLpkQRoMh", Filename: "attachment.txt", Filepath: "1lNigjr8fsntbweehfBLpkQRoMh.txt"},
					{ID: "1lNip3BucJnkDGo2uAChhgib20T", Filename: "attachment.png", Filepath: "1lNip3BucJnkDGo2uAChhgib20T.png"},
				},
				Date: time.Now().Add(-20 * time.Minute),
			},
			m2: {
				ID:      m2,
				Subject: "Change How You Think | Sinem Günel in Age of Awareness",
				Sender:  "noreply@google.com",
				From: []domain.Contact{
					{Email: "noreply@google.com", Name: "google"},
				},
				To: []domain.Contact{
					{Email: "bar@bar.com", Name: "Foobar"},
				},
				TextBody: "Collaborating with multiple people can be difficult, especially with lots of back and forth across email and chat.",
				HTMLBody: "<p>hello world</p>",
				Attachments: []domain.Attachment{
					{ID: "1lOs2EOqStjiMlDINR44KnMV9De", Filename: "attachment.txt", Filepath: "1lOs2EOqStjiMlDINR44KnMV9D.txt"},
					{ID: "1lOs60waIw6LoEqWNPmL7LRR06M", Filename: "attachment.png", Filepath: "1lOs60waIw6LoEqWNPmL7LRR06M.png"},
				},
				Date: time.Now(),
			},
		},
	}
}

// ByDateDesc ...
type ByDateDesc []domain.Message

func (d ByDateDesc) Len() int           { return len(d) }
func (d ByDateDesc) Less(i, j int) bool { return d[i].Date.UnixNano() > d[j].Date.UnixNano() }
func (d ByDateDesc) Swap(i, j int)      { d[i], d[j] = d[j], d[i] }
