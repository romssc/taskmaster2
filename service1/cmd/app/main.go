package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"

	"service1/config"
	"service1/internal/adapter/broker/kafkaa"
	"service1/internal/adapter/storage/inmemory"
	"service1/internal/controller/httprouter"
	"service1/internal/server/httpserver"
	"service1/internal/usecase/create"
	"service1/internal/usecase/list"
	"service1/internal/usecase/listid"
	"service1/internal/utils/id/uuidgen"
	"service1/internal/utils/json/standartjson"
	"service1/internal/utils/timestamp/standarttime"

	"golang.org/x/sync/errgroup"
)

// ENVIRONMENT VARIABLES:
//   - SERVER_HOST = SPECIFIES SERVER HOST, DEFAULTS TO "0.0.0.0"
//   - SERVER_PORT = SPECIFIES SERVER PORT, DEFAULTS TO "8081"
//   - KAFKA_ADDRESS = SPECIFIES KAFKA ADDRESSES, DEFAULTS TO []string{"0.0.0.0:9092"}
//   - KAFKA_TOPIC = SPECIFIES KAFKA TOPIC, DEFAULTS TO "tasks"
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

	generator := uuidgen.New()
	timer := standarttime.New()
	json := standartjson.New()

	storage := inmemory.New()

	config.Kafka.Encoder = json
	broker := kafkaa.New(config.Kafka)

	router := httprouter.New(&httprouter.Routes{
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

	sigCtx, sigCancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer sigCancel()

	eg, egCtx := errgroup.WithContext(sigCtx)

	eg.Go(func() error {
		if err := server.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	})

	eg.Go(func() error {
		<-egCtx.Done()
		toCtx, toCancel := context.WithTimeout(context.Background(), config.Server.ShutdownTimeout)
		defer toCancel()
		if err := server.Shutdown(toCtx); err != nil {
			log.Println(err)
		}
		if err := broker.Close(); err != nil {
			log.Println(err)
		}
		storage.Close()
		return nil
	})

	if err := eg.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}

	return nil
}
