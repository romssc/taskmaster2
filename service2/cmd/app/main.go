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

	config, cnErr := config.New(configPath)
	if cnErr != nil {
		return cnErr
	}
	config.Kafka.Brokers = brokers

	json := standartjson.New()

	router := kafkarouter.New(&kafkarouter.Config{
		Update: &update.Usecase{
			Config:  config.Router.Update,
			Decoder: json,
		},
	})

	config.Kafka.Handler = router
	broker := kafkaa.New(config.Kafka)

	snCtx, snCancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer snCancel()
	ewith, ewithCtx := errgroup.WithContext(snCtx)
	ewith.Go(func() error {
		if brunErr := broker.Run(); brunErr != nil && !errors.Is(brunErr, kafkaa.ErrOperationCanceled) {
			return brunErr
		}
		return nil
	})
	ewith.Go(func() error {
		<-ewithCtx.Done()
		bshutCtx, bshutCancel := context.WithTimeout(context.Background(), config.Kafka.ShutdownTimeout)
		defer bshutCancel()
		if bshutErr := broker.Shutdown(bshutCtx); bshutErr != nil {
			log.Println(bshutErr)
		}
		return nil
	})
	if ewaitErr := ewith.Wait(); ewaitErr != nil {
		return ewaitErr
	}

	return nil
}

func loadEnvs() (string, []string) {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yaml"
	}
	brokers := make([]string, 0, 1)
	kafkaAddr := os.Getenv("KAFKA_ADDR")
	if kafkaAddr == "" {
		kafkaAddr = "0.0.0.0:9092"
	}
	brokers = append(brokers, kafkaAddr)
	return configPath, brokers
}
