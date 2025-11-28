package httpserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

var (
	ErrTryingToListen = errors.New("server: failed while trying to listen and serve")
	ErrShuttingDown   = errors.New("server: failed while trying to shut down")
)

type Config struct {
	Host         string        `mapstructure:"host"`
	Port         string        `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`

	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`

	Handler http.Handler
}

type Server struct {
	server *http.Server
}

func New(c Config) *Server {
	return &Server{
		server: &http.Server{
			Addr:         strings.Join([]string{c.Host, c.Port}, ":"),
			Handler:      c.Handler,
			ReadTimeout:  c.ReadTimeout,
			WriteTimeout: c.WriteTimeout,
		},
	}
}

func (s *Server) Run() error {
	err := s.server.ListenAndServe()
	if err != nil {
		return fmt.Errorf("%v: %w", ErrTryingToListen, err)
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	err := s.server.Shutdown(ctx)
	if err != nil {
		return fmt.Errorf("%v: %w", ErrShuttingDown, err)
	}
	return nil
}
