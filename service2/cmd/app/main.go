package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	"service2/config"
	"service2/internal/controller/kafkarouter"
	"service2/internal/pkg/broker/kafkaa"
	"service2/internal/pkg/json/standartjson"
	"service2/internal/usecase/update"

	"golang.org/x/sync/errgroup"
)

// ENVIRONMENT VARIABLES:
// - CONFIG_PATH - SPECIFIES CONFIG .YAML FILE TO USE : DEFAULTS TO "service1/config.yaml"
// - KAFKA_ADDR - SPECIFIES KAFKA ADDRESS (EG. "0.0.0.0:9092") : DEFAULTS TO "[]string{"0.0.0.0:9092"}"

func main() {
	if err := run(); err != nil {
		log.Fatalf("critical: %v", err)
	}
}

func run() error {
	configPath, brokers := loadEnvs()

	config, err := config.New(configPath)
	if err != nil {
		return err
	}
	config.Kafka.Brokers = brokers

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
		if runErr := broker.Run(egCtx); runErr != nil && !errors.Is(err, kafkaa.ErrConsumerClosed) {
			return runErr
		}
		return nil
	})
	e.Go(func() error {
		<-egCtx.Done()
		if closeErr := broker.Close(); closeErr != nil && !errors.Is(err, kafkaa.ErrConsumerClosed) {
			return closeErr
		}
		return nil
	})
	if waitErr := e.Wait(); waitErr != nil && !errors.Is(waitErr, context.Canceled) {
		return waitErr
	}

	return nil
}

func loadEnvs() (string, []string) {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "service2/config.yaml"
	}
	brokers := make([]string, 0, 1)
	kafkaAddr := os.Getenv("KAFKA_ADDR")
	if kafkaAddr == "" {
		kafkaAddr = "0.0.0.0:9092"
	}
	brokers = append(brokers, kafkaAddr)
	return configPath, brokers
}
