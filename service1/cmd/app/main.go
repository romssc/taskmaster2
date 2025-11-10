package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"taskmaster2/service1/config"

	"taskmaster2/service1/internal/adapter/broker/kafkaa"
	"taskmaster2/service1/internal/adapter/storage/inmemory"
	"taskmaster2/service1/internal/controller/httprouter"
	"taskmaster2/service1/internal/pkg/server/httpserver"
	"taskmaster2/service1/internal/usecase/create"
	"taskmaster2/service1/internal/usecase/list"
	"taskmaster2/service1/internal/usecase/listid"

	"golang.org/x/sync/errgroup"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("critical: %v", err)
	}
}

func run() error {
	configPath := loadEnvs()
	config, err := config.New(configPath)
	if err != nil {
		return err
	}
	storage := inmemory.New()
	broker := kafkaa.New(config.Kafka)
	router := httprouter.New()
	create.New(storage, broker)
	list.New(storage)
	listid.New(storage)
	server := httpserver.New(router, config.Server)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
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
		ctx, cancel := context.WithTimeout(context.Background(), config.Server.ShutdownTimeout)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			return err
		}
		if err := broker.Close(); err != nil {
			return err
		}
		storage.Close()
		return nil
	})
	if err := e.Wait(); err != nil && !errors.Is(err, context.Canceled) {
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
