package main

import appconfig "github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/config"

type Config = appconfig.Sender

func NewConfig(path string) (Config, error) {
	return appconfig.NewSender(path)
}
