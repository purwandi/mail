package main

import (
	"context"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/purwandi/mail"
	"github.com/purwandi/mail/handler"
	"github.com/purwandi/mail/repository"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	shutdowns []func() error
	auth      *mail.Auth
	tls       *mail.TLS
	smtpPort  = "2525"
	httpPort  = "8080"
	hostname  = "localhost"
	ctx       = context.Background()
	logger, _ = zap.NewProduction(
		zap.AddStacktrace(zapcore.FatalLevel),
		zap.AddCallerSkip(1),
		zap.Fields(zap.String("type", "main")),
	)
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	godotenv.Load()

	run()
}

func run() {
	var shutdown = make(chan struct{})

	if os.Getenv("SMTP_PORT") != "" {
		smtpPort = os.Getenv("SMTP_PORT")
	}

	if os.Getenv("HTTP_PORT") != "" {
		httpPort = os.Getenv("HTTP_PORT")
	}

	db := repository.NewMessageInMemory(repository.NewMessageInMemoryStore())

	if os.Getenv("MAIL_AUTH") == "true" {
		auth = &mail.Auth{
			Enable:   true,
			Username: os.Getenv("MAIL_USERNAME"),
			Password: os.Getenv("MAIL_PASSWORD"),
		}
	}

	if os.Getenv("MAIL_TLS") == "true" {
		tls = &mail.TLS{
			Enable:   true,
			CertFile: os.Getenv("MAIL_TLS_CERT"),
			KeyFile:  os.Getenv("MAIL_TLS_KEY"),
		}
	}

	if os.Getenv("MAIL_HOSTNAME") != "" {
		hostname = os.Getenv("MAIL_HOSTNAME")
	}

	smtpd := handler.NewSMTPHandler(smtpPort, hostname, logger, auth, tls, db)
	httpd := handler.NewHTTPHandler(httpPort, logger, auth, db)

	// app service
	go func() {
		quit := make(chan os.Signal)
		signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
		<-quit

		logger.Info("shutting down server gracefully")
		smtpd.Close()
		httpd.Close(ctx)

		// close another module
		for i := range shutdowns {
			shutdowns[i]()
		}

		close(shutdown)
	}()

	// start the engine
	smtpd.Serve()
	httpd.Serve()
	<-shutdown
}
