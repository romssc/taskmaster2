package httpserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"taskmaster2/service1/internal/usecase/create"
	"taskmaster2/service1/internal/usecase/list"
	"taskmaster2/service1/internal/usecase/listid"
	"time"
)

var (
	ErrTryingToListen = errors.New("server: failed while trying to listen and serve")
	ErrShuttingDown   = errors.New("server: failed while trying to shut down")
)

type Config struct {
	Host         string        `yaml:"host"`
	Port         string        `yaml:"port"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`

	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`

	Routes Routes `yaml:"routes"`
}

type Routes struct {
	Create create.Config `yaml:"create"`
	List   list.Config   `yaml:"list"`
	ListID listid.Config `yaml:"list_id"`
}

type Server struct {
	server *http.Server
}

func New(h http.Handler, c Config) *Server {
	return &Server{
		server: &http.Server{
			Addr:         strings.Join([]string{c.Host, c.Port}, ":"),
			Handler:      h,
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
