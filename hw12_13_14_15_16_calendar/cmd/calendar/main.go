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

	strg, err := getStorage(config.Storage.Type)
	calendar := app.New(logg, strg)

	server := internalhttp.NewServer(logg, calendar)

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

func getStorage(storageType StorageType) (storage.Storage, error) {
	switch storageType {
	case StorageMemory:
		return memorystorage.New(), nil
	case StorageSql:
		return sqlstorage.New(), nil
	default:
		return nil, errors.New("unknown storage type: " + string(storageType))
	}
}
