package main

import (
	"log"

	"github.com/gibsn/duty_bot/cfg"
	"github.com/gibsn/duty_bot/dutyscheduler"
	"github.com/gibsn/duty_bot/notifychannel"
)

func main() {
	log.Println("info: starting duty_bot")

	config := cfg.NewConfig()

	if err := config.Validate(); err != nil {
		log.Fatalf("error: config is invalid: %v", err)
	}

	config.Print()

	log.Println("info: will start scheduling now")

	sch := dutyscheduler.NewDutyScheduler(config, notifychannel.StdOutNotifyChannel{})
	sch.Routine()

	log.Println("info: exiting")
}
