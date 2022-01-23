package main

import (
	"log"

	"github.com/gibsn/duty_bot/internal/cfg"
	"github.com/gibsn/duty_bot/internal/dutyscheduler"
)

func main() {
	log.Println("info: starting duty_bot")

	config, err := cfg.NewConfig()
	if err != nil {
		log.Fatalf("fatal: %v", err)
	}

	if err = config.Validate(); err != nil {
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
