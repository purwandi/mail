package repository

import (
	"context"
	"errors"
	"sort"

	"github.com/purwandi/mail/domain"
)

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

		// sort
		sort.Sort(ByDateDesc(messages))

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

// Reset ...
func (r *MessageRepositoryInMemory) Reset(ctx context.Context) <-chan error {
	err := make(chan error)

	go func() {
		r.store.mapMessage = map[string]domain.Message{}
		err <- nil
		close(err)
	}()

	return err
}
