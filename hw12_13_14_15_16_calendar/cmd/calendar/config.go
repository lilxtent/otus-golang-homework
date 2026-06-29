package main

import "github.com/spf13/viper"

type StorageType string

const (
	StorageMemory StorageType = "MEMORY"
	StorageSql    StorageType = "SQL"
)

// При желании конфигурацию можно вынести в internal/config.
// Организация конфига в main принуждает нас сужать API компонентов, использовать
// при их конструировании только необходимые параметры, а также уменьшает вероятность циклической зависимости.
type Config struct {
	Logger  LoggerConf
	Storage StorageConf
}

type LoggerConf struct {
	Level string
	// TODO
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

// TODO
