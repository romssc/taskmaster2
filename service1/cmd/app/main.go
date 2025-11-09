package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"taskmaster2/service1/config"
	"taskmaster2/service1/internal/adapter/storage/sqlite3"
	"taskmaster2/service1/internal/controller/httprouter"
	"taskmaster2/service1/internal/usecase/create"
	"taskmaster2/service1/internal/usecase/list"
	listid "taskmaster2/service1/internal/usecase/list_id"
	"taskmaster2/service1/pkg/server/httpserver"

	"golang.org/x/sync/errgroup"
)

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

func run() error {
	configPath := loadEnvs()
	config, err := config.New(configPath)
	if err != nil {
		return err
	}
	storage, err := sqlite3.New(config.SQLite3)
	if err != nil {
		return err
	}
	defer storage.Close()
	router := httprouter.New()
	create.New(storage)
	list.New(storage)
	listid.New(storage)
	server := httpserver.New(router, config.Server)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	e, c := errgroup.WithContext(ctx)
	e.Go(func() error {
		if err := server.Run(); err != nil && err != http.ErrServerClosed {
			return err
		}
		return nil
	})
	e.Go(func() error {
		<-c.Done()
		if err := server.Shutdown(config.Server.ShutdownTimeout); err != nil {
			return err
		}
		return nil
	})
	if err := e.Wait(); err != nil && err != context.Canceled {
		return err
	}
	return nil
}

func loadEnvs() string {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "service1_config.yaml"
	}
	return configPath
}
