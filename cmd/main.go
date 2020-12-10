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
	smtpPort  = "2525"
	httpPort  = "8080"
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

	// TODO get username and password from env
	auth := &mail.Auth{
		Enable:   true,
		Username: "username",
		Password: "password",
	}

	// TODO get certfile and keyfile from env
	tls := &mail.TLS{
		Enable:   true,
		CertFile: "./cert/certificate.crt",
		KeyFile:  "./cert/certificate.key",
	}

	smtpd := handler.NewSMTPHandler(smtpPort, logger, auth, tls, db)
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
