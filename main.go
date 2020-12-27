package main

import (
	"log"

	"github.com/gibsn/duty_bot/cfg"
	"github.com/gibsn/duty_bot/dutyscheduler"
)

func main() {
	log.Println("info: starting duty_bot")

	config := cfg.NewConfig()

	if err := config.Validate(); err != nil {
		log.Fatalf("error: config is invalid: %v", err)
	}

	config.Print()

	log.Println("info: will start scheduling now")

	sch := dutyscheduler.NewDutyScheduler(config)
	sch.Routine()

	log.Println("info: exiting")
}
