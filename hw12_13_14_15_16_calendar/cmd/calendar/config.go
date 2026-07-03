package main

import "github.com/spf13/viper"

type StorageType string

const (
	StorageMemory StorageType = "MEMORY"
	StorageSQL    StorageType = "SQL"
)

type Config struct {
	Logger  LoggerConf
	HTTP    HTTPConf
	GRPC    GRPCConf
	Storage StorageConf
}

type LoggerConf struct {
	Level string
	// TODO
}

type HTTPConf struct {
	Host string
	Port int
}

type GRPCConf struct {
	Host string
	Port int
}

type StorageConf struct {
	Type StorageType
	DSN  string
}

func NewConfig(path string) (Config, error) {
	v := viper.New()
	v.SetConfigFile(path)

	if err := v.ReadInConfig(); err != nil {
		return Config{}, err
	}

	config := Config{}
	if err := v.Unmarshal(&config); err != nil {
		return Config{}, err
	}
	return config, nil
}
