package main

import (
	"context"
	"errors"
	"log"
	"os/signal"
	"syscall"

	"service2/config"
	"service2/internal/broker/kafkaa"
	"service2/internal/controller/kafkarouter"
	"service2/internal/usecase/update"
	"service2/internal/utils/json/standartjson"

	"golang.org/x/sync/errgroup"
)

// ENVIRONMENT VARIABLES:
//   - KAFKA_BROKERS = SPECIFIES KAFKA ADDRESSES, DEFAULTS TO []string{"0.0.0.0:9092"}
//   - KAFKA_TOPIC = SPECIFIES KAFKA TOPIC, DEFAULTS TO "tasks"
//   - KAFKA_GROUP_ID = SPECIFIES KAFKA GROUP ID, DEFAULTS TO "tasks_group"
func main() {
	if err := run(); err != nil {
		log.Fatalf("exit with failure: %v", err)
	}
}

func run() error {
	config, err := config.New()
	if err != nil {
		return err
	}

	json := standartjson.New()

	router := kafkarouter.New(&kafkarouter.Routes{
		Update: &update.Usecase{
			Config:  config.Router.Update,
			Decoder: json,
		},
	})

	config.Kafka.Handler = router
	broker := kafkaa.New(config.Kafka)

	sigCtx, sigCancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer sigCancel()
	eg, egCtx := errgroup.WithContext(sigCtx)

	eg.Go(func() error {
		if err := broker.Run(egCtx); err != nil && !errors.Is(err, kafkaa.ErrOperationCanceled) {
			return err
		}
		return nil
	})

	eg.Go(func() error {
		<-egCtx.Done()
		if err := broker.Shutdown(); err != nil {
			log.Println(err)
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		return err
	}

	return nil
}
