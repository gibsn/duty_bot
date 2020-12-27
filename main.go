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

	sch, err := dutyscheduler.NewDutyScheduler(config)
	if err != nil {
		log.Fatalf("fatal: could not initialise scheduler: %v", err)
	}

	sch.Routine()

	log.Println("info: exiting")
}
