package main

import (
	"log"

	"github.com/gibsn/duty_bot/internal/app/dutybot"
	"github.com/gibsn/duty_bot/internal/app/dutybot/cfg"

	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	log.Println("info: starting duty_bot")

	config, err := cfg.NewConfig()
	if err != nil {
		log.Fatalf("fatal: %v", err)
	}

	if err = config.Validate(); err != nil {
		log.Fatalf("error: config is invalid: %v", err)
	}

	config.Print()

	bot, err := dutybot.NewDutyBot(config)
	if err != nil {
		log.Fatalf("fatal: could not initialise scheduler: %v", err)
	}

	bot.Wait()

	log.Println("info: exiting")
}
