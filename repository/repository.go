package repository

import (
	"context"

	"github.com/purwandi/mail/domain"
)

// QueryResult ...
type QueryResult struct {
	Result interface{}
	Error  error
}

// MessageRepository ...
type MessageRepository interface {
	Save(context.Context, *domain.Message) <-chan error
	Delete(context.Context, *domain.Message) <-chan error
	Reset(context.Context) <-chan error
	Find(context.Context, string) <-chan QueryResult
	FindAll(context.Context) <-chan QueryResult
}
