package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/app"
	"github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/logger"
	internalhttp "github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/server/http"
	"github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/storage"
	memorystorage "github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/storage/memory"
	sqlstorage "github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/storage/sql"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "configs/config.yaml", "Path to configuration file")
}

func main() {
	flag.Parse()

	if flag.Arg(0) == "version" {
		printVersion()
		return
	}

	config, err := NewConfig(configFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	logg, err := logger.New(config.Logger.Level)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	strg, closeStorage, err := getStorage(config.Storage)
	if err != nil {
		logg.Error("failed to initialize storage: " + err.Error())
		os.Exit(1)
	}
	defer func() {
		if err := closeStorage(); err != nil {
			logg.Error("failed to close storage: " + err.Error())
		}
	}()

	calendar := app.New(logg, strg)

	server := internalhttp.NewServer(logg, config.HTTP.Host, config.HTTP.Port, calendar)

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	go func() {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()

		if err := server.Stop(ctx); err != nil {
			logg.Error("failed to stop http server: " + err.Error())
		}
	}()

	logg.Info("calendar is running...")

	if err := server.Start(ctx); err != nil {
		logg.Error("failed to start http server: " + err.Error())
		cancel()
		os.Exit(1) //nolint:gocritic
	}
}

func getStorage(config StorageConf) (storage.Storage, func() error, error) {
	switch config.Type {
	case StorageMemory:
		return memorystorage.New(), func() error { return nil }, nil
	case StorageSql:
		db := sqlstorage.New(config.DSN)
		if err := db.Connect(); err != nil {
			return nil, nil, err
		}

		return db, db.Close, nil
	default:
		return nil, nil, errors.New("unknown storage type: " + string(config.Type))
	}
}
