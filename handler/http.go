package handler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/purwandi/mail"
	"github.com/purwandi/mail/domain"
	"github.com/purwandi/mail/repository"
	"go.uber.org/zap"
)

// HTTPHandler ...
type HTTPHandler struct {
	port       string
	logger     *zap.Logger
	echo       *echo.Echo
	auth       *mail.Auth
	repository repository.MessageRepository
}

// NewHTTPHandler ...
func NewHTTPHandler(port string, logger *zap.Logger, auth *mail.Auth, repo repository.MessageRepository) *HTTPHandler {
	e := echo.New()

	e.HideBanner = true
	e.HidePort = true
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Static("/", "public")

	return &HTTPHandler{
		port:       port,
		logger:     logger,
		auth:       auth,
		echo:       e,
		repository: repo,
	}
}

// List ...
func (s *HTTPHandler) List(c echo.Context) error {
	// Process
	result := <-s.repository.FindAll(c.Request().Context())
	if result.Error != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": result.Error.Error(),
		})
	}

	// Response
	return c.JSON(http.StatusOK, result.Result.([]domain.Message))
}

// Delete ...
func (s *HTTPHandler) Delete(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "message id is required",
		})
	}

	// Process
	err := <-s.repository.Delete(c.Request().Context(), id)
	if err != nil {
		s.logger.Error("Unable to delete message", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Unable to delete message",
		})
	}

	// Response
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status": "ok",
	})
}

// DeleteAll ...
func (s *HTTPHandler) DeleteAll(c echo.Context) error {
	// Process
	err := <-s.repository.Reset(c.Request().Context())
	if err != nil {
		s.logger.Error("Unable to clear all messages", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Unable to clear all messages",
		})
	}

	// Response
	return c.JSON(http.StatusOK, "")
}

// Download ...
func (s *HTTPHandler) Download(c echo.Context) error {
	f := c.Param("file")
	if f == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status": "file params is required",
		})
	}

	// Response
	return c.Attachment(fmt.Sprintf("./public/assets/%s", f), f)
}

// Serve for serve
func (s *HTTPHandler) Serve() {
	s.echo.GET("/download/:file", s.Download)
	s.echo.GET("/api/message", s.List)
	s.echo.DELETE("/api/message/:id", s.Delete)
	s.echo.DELETE("/api/reset", s.DeleteAll)

	go func() {
		if err := s.echo.Start(fmt.Sprintf(":%s", s.port)); err != nil {
			s.logger.Info("shutting down http server")
		}
	}()

	s.logger.Info(fmt.Sprintf("http service is running at %s", s.port))
}

// Close ...
func (s *HTTPHandler) Close(ctx context.Context) error {
	return s.echo.Shutdown(ctx)
}
