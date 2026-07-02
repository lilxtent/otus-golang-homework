package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/logger"
	"github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/queue/rabbitmq"
	"github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/scheduler"
	sqlstorage "github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/storage/sql"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "configs/scheduler_config.yaml", "Path to configuration file")
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

	log, err := logger.New(config.Logger.Level)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	strg, err := getStorage(config.Storage)
	if err != nil {
		log.Error("failed to initialize storage: " + err.Error())
		os.Exit(1)
	}
	defer func() {
		if err := strg.Close(); err != nil {
			log.Error("failed to close storage: " + err.Error())
		}
	}()

	broker, err := rabbitmq.New(config.Queue)
	if err != nil {
		log.Error("failed to initialize queue: " + err.Error())
		os.Exit(1)
	}
	defer func() {
		if err := broker.Close(); err != nil {
			log.Error("failed to close queue: " + err.Error())
		}
	}()

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	service := scheduler.New(strg, broker, log, scheduler.Config{
		ScanInterval:    config.Scheduler.ScanInterval,
		CleanupInterval: config.Scheduler.CleanupInterval,
	})

	log.Info("calendar scheduler is running...")
	if err := service.Run(ctx); err != nil {
		log.Error("scheduler stopped with error: " + err.Error())
		os.Exit(1)
	}
}

func getStorage(config StorageConf) (*sqlstorage.SQLStorage, error) {
	switch config.Type {
	case StorageSQL:
		db := sqlstorage.New(config.DSN)
		if err := db.Connect(); err != nil {
			return nil, errors.Join(err, db.Close())
		}

		return db, nil
	default:
		return nil, errors.New("unknown storage type: " + string(config.Type))
	}
}
