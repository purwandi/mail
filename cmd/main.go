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

	db := repository.NewMessageInMemory(repository.NewMessageInMemoryStore())
	auth := &mail.Auth{
		Enable:   true,
		Username: "username",
		Password: "password",
	}
	tls := &mail.TLS{
		Enable:   true,
		CertFile: "./cert/certificate.crt",
		KeyFile:  "./cert/certificate.key",
	}
	smtpd := handler.NewSMTPHandler("2525", logger, auth, tls, db)
	httpd := handler.NewHTTPHandler("8080", logger, auth, db)

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

	smtpd.Serve()
	httpd.Serve()
	<-shutdown
}
