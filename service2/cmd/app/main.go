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

	notifyCtx, notifyCancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer notifyCancel()
	e, egCtx := errgroup.WithContext(notifyCtx)
	e.Go(func() error {
		if brErr := broker.Run(egCtx); brErr != nil && !errors.Is(brErr, kafkaa.ErrConsumerClosed) {
			return brErr
		}
		return nil
	})
	e.Go(func() error {
		<-egCtx.Done()
		if bcErr := broker.Close(); bcErr != nil {
			log.Panicln(bcErr)
		}
		return nil
	})
	if waitErr := e.Wait(); waitErr != nil && !errors.Is(waitErr, context.Canceled) {
		return waitErr
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
