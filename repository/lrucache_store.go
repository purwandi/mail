package repository

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
	"time"

	lru "github.com/hashicorp/golang-lru"
	"github.com/purwandi/mail"
	"github.com/purwandi/mail/domain"
	"github.com/segmentio/ksuid"
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
	m1 := ksuid.New().String()
	m2 := ksuid.New().String()
	c.Add(m1, domain.Message{
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
	})
	c.Add(m2, domain.Message{
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
	})
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
