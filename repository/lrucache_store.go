package repository

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"

	lru "github.com/hashicorp/golang-lru"
	"github.com/purwandi/mail"
	"github.com/purwandi/mail/domain"
	"go.uber.org/zap"
)

type lruCache struct {
	cache *lru.Cache
}

func evictionFn(logger *zap.Logger) func(key interface{}, value interface{}) {
	return func(key interface{}, value interface{}) {
		m, ok := value.(domain.Message)
		if !ok {
			return
		}
		for _, att := range m.Attachments {
			fullPath := fmt.Sprintf("%s/%s", mail.AssetFilePath, att.Filepath)
			err := os.Remove(fullPath)
			if err != nil {
				msg := fmt.Sprintf("failed deleting file: %s from message id: %s", fullPath, key)
				logger.Error(msg, zap.Error(err))
			}
		}
	}

}

func NewLRUCacheStore(size int, logger *zap.Logger) (MessageRepository, error) {
	c, err := lru.NewWithEvict(size, evictionFn(logger))
	if err != nil {
		return nil, err
	}
	return &lruCache{
		cache: c,
	}, nil
}

func (c *lruCache) Save(ctx context.Context, m *domain.Message) <-chan error {
	ch := make(chan error)
	go func() {
		c.cache.Add(m.ID, *m)
		ch <- nil
		close(ch)
	}()

	return ch
}
func (c *lruCache) Delete(ctx context.Context, id string) <-chan error {
	ch := make(chan error)
	go func() {
		present := c.cache.Remove(id)
		if present {
			ch <- nil
			close(ch)
			return
		}
		ch <- errors.New("Message not found")
		close(ch)
	}()

	return ch
}
func (c *lruCache) Reset(ctx context.Context) <-chan error {
	ch := make(chan error)
	go func() {
		c.cache.Purge()
		ch <- nil
		close(ch)
	}()

	return ch
}
func (c *lruCache) Find(ctx context.Context, id string) <-chan QueryResult {
	ch := make(chan QueryResult)
	go func() {
		var qr QueryResult
		inf, ok := c.cache.Get(id)
		if !ok {
			qr.Error = errors.New("Message not found")
			ch <- qr
			return
		}
		qr.Result = inf
		ch <- qr
		close(ch)
	}()
	return ch
}
func (c *lruCache) FindAll(context.Context) <-chan QueryResult {
	ch := make(chan QueryResult)
	go func() {
		var (
			qr   QueryResult
			msgs []domain.Message
		)
		for _, k := range c.cache.Keys() {
			inf, _ := c.cache.Get(k)
			msgs = append(msgs, inf.(domain.Message))
		}
		sort.Sort(ByDateDesc(msgs))
		qr.Result = msgs
		ch <- qr
		close(ch)
	}()
	return ch
}
