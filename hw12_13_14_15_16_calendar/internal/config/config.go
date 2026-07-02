package config

import (
	"time"

	"github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/queue/rabbitmq"
	"github.com/spf13/viper"
)

type StorageType string

const (
	StorageMemory StorageType = "MEMORY"
	StorageSQL    StorageType = "SQL"
)

type Calendar struct {
	Logger  LoggerConf  `mapstructure:"logger"`
	HTTP    HTTPConf    `mapstructure:"http"`
	GRPC    GRPCConf    `mapstructure:"grpc"`
	Storage StorageConf `mapstructure:"storage"`
}

type Scheduler struct {
	Logger    LoggerConf      `mapstructure:"logger"`
	Storage   StorageConf     `mapstructure:"storage"`
	Queue     rabbitmq.Config `mapstructure:"queue"`
	Scheduler SchedulerConf   `mapstructure:"scheduler"`
}

type LoggerConf struct {
	Level string `mapstructure:"level"`
}

type HTTPConf struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type GRPCConf struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type StorageConf struct {
	Type StorageType `mapstructure:"type"`
	DSN  string      `mapstructure:"dsn"`
}

type SchedulerConf struct {
	ScanInterval    time.Duration `mapstructure:"scan_interval"`
	CleanupInterval time.Duration `mapstructure:"cleanup_interval"`
}

func NewCalendar(path string) (Calendar, error) {
	config := Calendar{}
	if err := Load(path, &config); err != nil {
		return Calendar{}, err
	}

	return config, nil
}

func NewScheduler(path string) (Scheduler, error) {
	config := Scheduler{}
	if err := Load(path, &config); err != nil {
		return Scheduler{}, err
	}

	return config, nil
}

func Load(path string, target any) error {
	v := viper.New()
	v.SetConfigFile(path)

	if err := v.ReadInConfig(); err != nil {
		return err
	}

	return v.Unmarshal(target)
}
