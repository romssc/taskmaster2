package config

import (
	"errors"
	"fmt"
	"strings"

	"service2/internal/broker/kafkaa"
	"service2/internal/controller/kafkarouter"

	"github.com/spf13/viper"
)

var (
	ErrReadingConfig = errors.New("config: failed to configure")
)

type Config struct {
	Kafka  kafkaa.Config      `mapstructure:"kafka"`
	Router kafkarouter.Config `mapstructure:"router"`
}

func New() (Config, error) {
	v := viper.New()

	v.SetDefault("kafka.brokers", []string{"0.0.0.0:9092"})
	v.SetDefault("kafka.topic", "tasks")
	v.SetDefault("kafka.group_id", "tasks-group")

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./service2/")

	if err := v.ReadInConfig(); err != nil {
		return Config{}, fmt.Errorf("%w: %v", ErrReadingConfig, err)
	}

	v.SetEnvPrefix("")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	var c Config
	if err := v.Unmarshal(&c); err != nil {
		return Config{}, fmt.Errorf("%w: %v", ErrReadingConfig, err)
	}

	return c, nil
}
