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
	internalgrpc "github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/server/grpc"
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
		closer, ok := strg.(interface {
			Close() error
		})
		if !ok {
			return
		}
		if err := closer.Close(); err != nil {
			log.Error("failed to close storage: " + err.Error())
		}
	}()

	calendar := app.New(strg)

	httpServer := internalhttp.NewServer(log, config.HTTP.Host, config.HTTP.Port, calendar)
	grpcServer := internalgrpc.NewServer(log, config.GRPC.Host, config.GRPC.Port, calendar)

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	go func() {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()

		if err := httpServer.Stop(ctx); err != nil {
			log.Error("failed to stop http server: " + err.Error())
		}
		if err := grpcServer.Stop(ctx); err != nil {
			log.Error("failed to stop grpc server: " + err.Error())
		}
	}()

	log.Info("calendar is running...")

	errCh := make(chan error, 2)
	go func() {
		if err := httpServer.Start(ctx); err != nil {
			errCh <- fmt.Errorf("failed to start http server: %w", err)
		}
	}()
	go func() {
		if err := grpcServer.Start(ctx); err != nil {
			errCh <- fmt.Errorf("failed to start grpc server: %w", err)
		}
	}()

	select {
	case <-ctx.Done():
	case err := <-errCh:
		log.Error(err.Error())
		cancel()
		os.Exit(1) //nolint:gocritic
	}
}

func getStorage(config StorageConf) (storage.Storage, error) {
	switch config.Type {
	case StorageMemory:
		return memorystorage.New(), nil
	case StorageSQL:
		db := sqlstorage.New(config.DSN)
		if err := db.Connect(); err != nil {
			return nil, err
		}

		return db, nil
	default:
		return nil, errors.New("unknown storage type: " + string(config.Type))
	}
}
