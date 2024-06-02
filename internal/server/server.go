package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/i7tsov/example-microservice/pkg/model"
	"github.com/sirupsen/logrus"
)

const shutdownTimeout = 3 * time.Second

var internalError = model.InternalError{Message: "Internal server error"}

// Server implements application HTTP server.
// Business logic is implemented in server endpoint functions.
type Server struct {
	c      Config
	d      Dependencies
	router *gin.Engine
}

// Config contains configuration options for the Server.
type Config struct {
	Address string `yaml:"address"`
	Port    int    `yaml:"port" validate:"required"`
}

// Config contains all dependencies needed by business logic.
type Dependencies struct {
	UsersRepo usersRepo `validate:"required"`
}

// Dependency we're relying on.
type usersRepo interface {
	AddUser(ctx context.Context, user model.User) (string, error)
	RemoveUser(ctx context.Context, id string) error
	ListUsers(ctx context.Context) ([]model.User, error)
}

// New creates new HTTP server for the microservice.
func New(c Config, d Dependencies) (*Server, error) {
	// Validate config & dependencies.
	validate := validator.New()
	err := validate.Struct(c)
	if err != nil {
		// Error contains sufficient information about the cause, no wrapping needed.
		return nil, err
	}
	err = validate.Struct(d)
	if err != nil {
		// Error contains sufficient information about the cause, no wrapping needed.
		return nil, err
	}

	// Set default values.
	if c.Port == 0 {
		c.Port = 80
	}

	s := &Server{
		c: c,
		d: d,
	}

	s.setupRoutes()

	return s, nil
}

// Set up routes: conect URLs to respective handlers.
func (s *Server) setupRoutes() {
	s.router = gin.Default()
	app := s.router.Group("/v1")

	app.POST("/users", s.createUser)
	app.GET("/users", s.listUsers)
}

// Serve starts serving endpoints. Handles graceful shutdown.
//
// Blocking.
func (s *Server) Serve(ctx context.Context) error {
	srv := http.Server{
		Addr:    fmt.Sprintf("%v:%v", s.c.Address, s.c.Port),
		Handler: s.router,
	}

	go func() {
		<-ctx.Done()
		sdCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		err := srv.Shutdown(sdCtx)
		if err != nil {
			logrus.Errorf("Clean server shutdown failed: %v (%T)", err, err)
		} else {
			logrus.Errorf("Server shut down")
		}
	}()

	fmt.Printf("Serving HTTP at %v:%v\n", s.c.Address, s.c.Port)
	err := srv.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		err = nil
	}
	return err
}
