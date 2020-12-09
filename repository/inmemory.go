package repository

import (
	"context"
	"errors"

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
				Subject: "7 Quotes by Albert Einstein That Will Change How You Think | Sinem Günel in Age of Awareness",
				From: []domain.Contact{
					{Email: "noreply@medium.com", Name: "Medium"},
				},
				To: []domain.Contact{
					{Email: "foo@bar.com", Name: "Fooabar"},
				},
				TextBody: "hello world",
				HTMLBody: "<p>hello world</p>",
				Attachments: []domain.Attachment{
					{ID: "1lNigjr8fsntbweehfBLpkQRoMh.txt", Filename: "attachment.txt"},
					{ID: "1lNip3BucJnkDGo2uAChhgib20T.png", Filename: "attachment.png"},
				},
			},
			m2: {
				ID:      m1,
				Subject: "Change How You Think | Sinem Günel in Age of Awareness",
				From: []domain.Contact{
					{Email: "noreply@google.com", Name: "google"},
				},
				To: []domain.Contact{
					{Email: "bar@bar.com", Name: "Foobar"},
				},
				TextBody: "hello world",
				HTMLBody: "<p>hello world</p>",
				Attachments: []domain.Attachment{
					{ID: "1lOs2EOqStjiMlDINR44KnMV9De.txt", Filename: "attachment.txt"},
					{ID: "1lOs60waIw6LoEqWNPmL7LRR06M.png", Filename: "attachment.png"},
				},
			},
		},
	}
}

// MessageRepositoryInMemory ...
type MessageRepositoryInMemory struct {
	store *MessageInMemoryStore
}

// NewMessageInMemory ...
func NewMessageInMemory(store *MessageInMemoryStore) *MessageRepositoryInMemory {
	return &MessageRepositoryInMemory{store: store}
}

// Save ...
func (r *MessageRepositoryInMemory) Save(ctx context.Context, m *domain.Message) <-chan error {
	err := make(chan error)

	go func() {
		r.store.mapMessage[m.ID] = *m

		err <- nil
		close(err)
	}()

	return err
}

// FindAll ...
func (r *MessageRepositoryInMemory) FindAll(ctx context.Context) <-chan QueryResult {
	result := make(chan QueryResult)

	go func() {
		messages := []domain.Message{}
		for _, item := range r.store.mapMessage {
			messages = append(messages, item)
		}

		result <- QueryResult{Result: messages}
		close(result)
	}()

	return result
}

// Find ...
func (r *MessageRepositoryInMemory) Find(ctx context.Context, id string) <-chan QueryResult {
	result := make(chan QueryResult)

	go func() {
		for _, item := range r.store.mapMessage {
			if item.ID == id {
				result <- QueryResult{Result: item}
				break
			}
		}

		result <- QueryResult{Error: errors.New("Message not found")}
		close(result)
	}()

	return result
}

// Delete ...
func (r *MessageRepositoryInMemory) Delete(ctx context.Context, message *domain.Message) <-chan error {
	err := make(chan error)

	go func() {
		for _, item := range r.store.mapMessage {
			if item.ID == message.ID {
				delete(r.store.mapMessage, message.ID)
				err <- nil
				break
			}
		}

		err <- errors.New("Message not found")
		close(err)
	}()

	return err
}
