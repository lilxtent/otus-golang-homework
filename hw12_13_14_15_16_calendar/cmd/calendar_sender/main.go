package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/logger"
	"github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/queue/rabbitmq"
	"github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/sender"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "configs/sender_config.yaml", "Path to configuration file")
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

	service := sender.New(broker, log)

	log.Info("calendar sender is running...")
	if err := service.Run(ctx); err != nil {
		log.Error("sender stopped with error: " + err.Error())
		os.Exit(1)
	}
}
