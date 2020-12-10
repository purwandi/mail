package handler

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/jhillyerd/enmime"
	"github.com/mhale/smtpd"
	"github.com/purwandi/mail"
	"github.com/purwandi/mail/domain"
	"github.com/purwandi/mail/repository"
	"github.com/segmentio/ksuid"
	"go.uber.org/zap"
)

// SMTPHandler ...
type SMTPHandler struct {
	port       string
	logger     *zap.Logger
	smtpd      *smtpd.Server
	repository repository.MessageRepository
	auth       *mail.Auth
	tls        *mail.TLS
}

// NewSMTPHandler for creating new smtp
func NewSMTPHandler(port string, logger *zap.Logger, auth *mail.Auth, tls *mail.TLS, repo repository.MessageRepository) *SMTPHandler {
	smtp := &smtpd.Server{
		Addr:     fmt.Sprintf(":%s", port),
		Appname:  "Mailbox",
		Hostname: "onelabs.dev",
	}

	return &SMTPHandler{
		port:       port,
		logger:     logger,
		smtpd:      smtp,
		auth:       auth,
		tls:        tls,
		repository: repo,
	}
}

// Serve for serve
func (s *SMTPHandler) Serve() {
	s.smtpd.Handler = s.mailHandler

	if s.auth != nil {
		s.smtpd.AuthRequired = s.auth.Enable
		s.smtpd.AuthHandler = s.authHandler
		s.smtpd.AuthMechs = map[string]bool{
			"PLAIN":    true,
			"LOGIN":    true,
			"CRAM-MD5": true,
		}
	}

	if s.tls != nil && s.tls.Enable == true {
		if err := s.smtpd.ConfigureTLS(s.tls.CertFile, s.tls.KeyFile); err != nil {
			s.logger.Error("Unable to configure tls", zap.Error(err))
		}
	}

	go func() {
		if err := s.smtpd.ListenAndServe(); err != nil {
			s.logger.Info("shutting down smtp server")
		}
	}()

	s.logger.Info(fmt.Sprintf("smtp service is running at %s", s.port))
}

// Close for close operation
func (s *SMTPHandler) Close() error {
	return nil
}

func (s *SMTPHandler) mailHandler(origin net.Addr, sender string, to []string, data []byte) {
	m, err := enmime.ReadEnvelope(bytes.NewReader(data))
	if err != nil {
		s.logger.Error("Unable to read message", zap.Error(err))
	}

	mID := ksuid.New().String()

	alist, _ := m.AddressList("To")
	tos := []domain.Contact{}
	for _, addr := range alist {
		tos = append(tos, domain.Contact{
			Name:  addr.Name,
			Email: addr.Address,
		})
	}

	alist, _ = m.AddressList("From")
	froms := []domain.Contact{}
	for _, addr := range alist {
		froms = append(froms, domain.Contact{
			Name:  addr.Name,
			Email: addr.Address,
		})
	}

	attch := []domain.Attachment{}
	for _, item := range m.Attachments {
		fid := ksuid.New().String()
		attch = append(attch, domain.Attachment{
			ID:       fid,
			Filename: item.FileName,
			Filepath: fmt.Sprintf("%s%s", fid, filepath.Ext(item.FileName)),
			Type:     item.ContentType,
		})

		// write into file
		f, err := os.Create(fmt.Sprintf("./public/assets/%s%s", fid, filepath.Ext(item.FileName)))
		if err != nil {
			fmt.Println(err)
			return
		}

		f.Write(item.Content)
		f.Close()
	}

	message := domain.Message{
		ID:          mID,
		Subject:     m.GetHeader("Subject"),
		Sender:      sender,
		From:        froms,
		To:          tos,
		Attachments: attch,
		HTMLBody:    m.HTML,
		TextBody:    m.Text,
		Date:        time.Now(),
	}

	err = <-s.repository.Save(context.Background(), &message)
	if err != nil {
		s.logger.Error("Unable to insert db", zap.Error(err))
	}

	return
}

func (s *SMTPHandler) authHandler(remote net.Addr, mechanism string, username, password, shared []byte) (bool, error) {
	return string(username) == s.auth.Username && string(password) == s.auth.Password, nil
}
