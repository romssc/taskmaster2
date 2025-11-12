package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"
	"taskmaster2/service2/config"
	"taskmaster2/service2/internal/controller/kafkarouter"
	"taskmaster2/service2/internal/pkg/broker/kafkaa"
	"taskmaster2/service2/internal/pkg/json/standartjson"
	"taskmaster2/service2/internal/usecase/update"

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

	json := standartjson.New()

	router := kafkarouter.New(&kafkarouter.Config{
		Update: &update.Usecase{
			Config: config.Router.Update,
		},
	})

	config.Kafka.Handler = router
	config.Kafka.Encoder = json
	config.Kafka.Decoder = json
	broker := kafkaa.New(config.Kafka)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()
	e, c := errgroup.WithContext(ctx)
	e.Go(func() error {
		if err := broker.Run(c); err != nil && !errors.Is(err, kafkaa.ErrConsumerClosed) {
			return err
		}
		return nil
	})
	e.Go(func() error {
		<-c.Done()
		if err := broker.Close(); err != nil {
			log.Panicln(err)
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
		configPath = "service2/config.yaml"
	}
	return configPath
}
