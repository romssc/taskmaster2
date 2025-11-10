package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"
	"taskmaster2/service2/config"
	"taskmaster2/service2/internal/adapter/storage/inmemory"
	"taskmaster2/service2/internal/controller/kafkarouter"
	"taskmaster2/service2/internal/pkg/broker/kafkaa"

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
	router := kafkarouter.New(storage)
	broker := kafkaa.New(router, config.Kafka)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()
	e, c := errgroup.WithContext(ctx)
	e.Go(func() error {
		if err := broker.Run(c); err != nil {
			return err
		}
		return nil
	})
	e.Go(func() error {
		<-c.Done()
		defer func() {
			storage.Close()
		}()
		if err := broker.Close(); err != nil {
			return err
		}
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
		configPath = "service2_config.yaml"
	}
	return configPath
}
