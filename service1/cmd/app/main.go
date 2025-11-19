package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"service1/config"
	"service1/internal/adapter/broker/kafkaa"
	"service1/internal/adapter/storage/inmemory"
	"service1/internal/controller/httprouter"
	"service1/internal/pkg/id/uuidgen"
	"service1/internal/pkg/json/standartjson"
	"service1/internal/pkg/server/httpserver"
	"service1/internal/pkg/timestamp/standarttime"
	"service1/internal/usecase/create"
	"service1/internal/usecase/list"
	"service1/internal/usecase/listid"

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

	config, cnewErr := config.New(configPath)
	if cnewErr != nil {
		return cnewErr
	}
	config.Kafka.Address = brokers

	generator := uuidgen.New()
	timer := standarttime.New()
	json := standartjson.New()

	storage := inmemory.New()

	config.Kafka.Encoder = json
	broker := kafkaa.New(config.Kafka)

	router := httprouter.New(&httprouter.Config{
		Create: &create.Usecase{
			Config:    config.Router.Create,
			Creator:   storage,
			Publisher: broker,
			Generator: generator,
			Timer:     timer,
			Encoder:   json,
			Decoder:   json,
		},
		List: &list.Usecase{
			Config:  config.Router.List,
			Getter:  storage,
			Encoder: json,
		},
		ListID: &listid.Usecase{
			Config:  config.Router.ListID,
			Getter:  storage,
			Encoder: json,
		},
	})

	config.Server.Handler = router
	server := httpserver.New(config.Server)

	egCtx, egCancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer egCancel()
	ewith, ewithCtx := errgroup.WithContext(egCtx)
	ewith.Go(func() error {
		if srunErr := server.Run(); srunErr != nil && !errors.Is(srunErr, http.ErrServerClosed) {
			return srunErr
		}
		return nil
	})
	ewith.Go(func() error {
		<-ewithCtx.Done()
		sshutCtx, sshutCancel := context.WithTimeout(context.Background(), config.Server.ShutdownTimeout)
		defer sshutCancel()
		if sshutErr := server.Shutdown(sshutCtx); sshutErr != nil {
			log.Println(sshutErr)
		}
		if bcloseErr := broker.Close(); bcloseErr != nil {
			log.Println(bcloseErr)
		}
		storage.Close()
		return nil
	})
	if ewaitErr := ewith.Wait(); ewaitErr != nil && !errors.Is(ewaitErr, context.Canceled) {
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
