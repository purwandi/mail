package handler

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/purwandi/mail"
	"github.com/purwandi/mail/domain"
	"github.com/purwandi/mail/repository"
	"go.uber.org/zap"
)

const sessionName = "mailbox"

// HTTPHandler ...
type HTTPHandler struct {
	port          string
	logger        *zap.Logger
	echo          *echo.Echo
	auth          *mail.Auth
	authorization *AuthMiddleware
	repository    repository.MessageRepository
}

// NewHTTPHandler ...
func NewHTTPHandler(port, secret string, logger *zap.Logger, auth *mail.Auth, repo repository.MessageRepository) *HTTPHandler {
	e := echo.New()

	e.HideBanner = true
	e.HidePort = true

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(session.Middleware(sessions.NewCookieStore([]byte(secret))))

	e.Static("/", "public")

	return &HTTPHandler{
		port:          port,
		logger:        logger,
		auth:          auth,
		authorization: NewAuthMiddleware(auth),
		echo:          e,
		repository:    repo,
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
	return c.Attachment(fmt.Sprintf("%s/%s", mail.AssetFilePath, f), f)
}

// Login ...
func (s *HTTPHandler) Login(c echo.Context) error {
	file, err := ioutil.ReadFile("./public/login.html")
	if err != nil {
		return c.HTML(http.StatusInternalServerError, "Unable to get login template file")
	}

	return c.HTML(http.StatusOK, string(file))
}

// LoginProcess ...
func (s *HTTPHandler) LoginProcess(c echo.Context) error {
	// given
	username := c.FormValue("username")
	password := c.FormValue("password")

	if username == "" {
		return c.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
			"message": "username can't be blank",
		})
	}

	if password == "" {
		return c.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
			"message": "password can't be blank",
		})
	}

	if s.auth == nil || s.auth.Enable != true {
		return c.JSON(http.StatusForbidden, map[string]interface{}{
			"message": "auth is not enabled",
		})
	}

	if username != s.auth.Username || password != s.auth.Password {
		return c.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
			"message": "credential is not match",
		})
	}

	sess, err := session.Get(sessionName, c)
	if err != nil {
		return err
	}

	sess.Options = &sessions.Options{
		Path:   "/",
		MaxAge: 86400 * 7,
	}

	sess.Values["authorize"] = "ok"
	sess.Save(c.Request(), c.Response())

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "ok",
	})
}

// Serve for serve
func (s *HTTPHandler) Serve() {
	s.echo.GET("/login", s.Login)
	s.echo.POST("/login", s.LoginProcess)

	s.echo.GET("/download/:file", s.Download, s.authorization.Check)
	s.echo.GET("/api/message", s.List, s.authorization.Check)
	s.echo.DELETE("/api/message/:id", s.Delete, s.authorization.Check)
	s.echo.DELETE("/api/reset", s.DeleteAll, s.authorization.Check)

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

// AuthMiddleware ...
type AuthMiddleware struct {
	auth *mail.Auth
}

// NewAuthMiddleware ...
func NewAuthMiddleware(auth *mail.Auth) *AuthMiddleware {
	return &AuthMiddleware{auth: auth}
}

// Check ...
func (m *AuthMiddleware) Check(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if m.auth != nil && m.auth.Enable == true {
			sess, err := session.Get(sessionName, c)
			if err != nil {
				return c.Redirect(http.StatusTemporaryRedirect, "/")
			}

			if sess.Values["authorize"] == nil {
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{
					"message": "is not authorized",
				})
			}
		}

		return next(c)
	}
}
