package main

import appconfig "github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/config"

const StorageSQL = appconfig.StorageSQL

type Config = appconfig.Scheduler
type StorageConf = appconfig.StorageConf

func NewConfig(path string) (Config, error) {
	return appconfig.NewScheduler(path)
}
