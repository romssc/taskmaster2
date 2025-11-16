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

	notifyCtx, notifyCancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer notifyCancel()
	e, egCtx := errgroup.WithContext(notifyCtx)
	e.Go(func() error {
		if srErr := server.Run(); srErr != nil && !errors.Is(srErr, http.ErrServerClosed) {
			return srErr
		}
		return nil
	})
	e.Go(func() error {
		<-egCtx.Done()
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), config.Server.ShutdownTimeout)
		defer shutdownCancel()
		if ssErr := server.Shutdown(shutdownCtx); ssErr != nil {
			log.Println(ssErr)
		}
		if bcErr := broker.Close(); bcErr != nil {
			log.Println(bcErr)
		}
		storage.Close()
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
		configPath = "service1/config.yaml"
	}
	return configPath
}
