package dutybot

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/gibsn/duty_bot/internal/app/dutybot/cfg"
	"github.com/gibsn/duty_bot/internal/dutyscheduler"
	"github.com/gibsn/duty_bot/internal/productioncal"
	"github.com/gibsn/duty_bot/internal/statedumper"
)

// TODO comment
type DutyBot struct {
	cfg cfg.Config

	schedulers    []*dutyscheduler.DutyScheduler
	stateDumper   *statedumper.FileDumper
	productionCal *productioncal.ProductionCal

	shutdownOnce *sync.Once
	finished     chan struct{}
}

// TODO comment
func NewDutyBot(cfg cfg.Config) (*DutyBot, error) {
	bot := &DutyBot{
		cfg:          cfg,
		shutdownOnce: new(sync.Once),
		finished:     make(chan struct{}, 1),
	}

	if cfg.ProductionCal.Enabled {
		bot.initProductionCal()
	}

	if err := bot.initStateDumper(); err != nil {
		return nil, fmt.Errorf("could not init state dumper: %w", err)
	}

	for _, projectCfg := range cfg.Projects {
		sch, err := dutyscheduler.NewDutyScheduler(
			projectCfg, bot.stateDumper, bot.productionCal,
		)
		if err != nil {
			return nil, fmt.Errorf("could not init project '%s': %w", projectCfg.Name, err)
		}

		bot.schedulers = append(bot.schedulers, sch)
	}

	go bot.signalHandler()

	return bot, nil
}

func (bot *DutyBot) initProductionCal() {
	bot.productionCal = productioncal.NewProductionCal(bot.cfg.ProductionCal)

	if err := bot.productionCal.Init(); err != nil {
		log.Printf("error: could not initialise production calendar: %v", err)
		log.Println("warning: day offs recognition will be unavailable until next refetch")
	} else {
		log.Println("info: initialised production calendar")
	}

	go bot.productionCal.Routine()
}

//nolint: unparam
func (bot *DutyBot) initStateDumper() error {
	stateDumper, err := statedumper.NewFileDumper()
	if err != nil {
		return err
	}

	bot.stateDumper = stateDumper

	return nil
}

func (bot *DutyBot) signalHandler() {
	signalQ := make(chan os.Signal, 1)
	signal.Notify(signalQ, syscall.SIGTERM, syscall.SIGINT)

	for s := range signalQ {
		log.Printf("info: received %s", s)
		bot.Shutdown()
	}
}

// Wait waits for the bot finish gracefully
func (bot *DutyBot) Wait() {
	<-bot.finished
}

func (bot *DutyBot) Shutdown() {
	log.Println("info: shutting down")

	for _, sch := range bot.schedulers {
		sch.Shutdown()
	}

	bot.stateDumper.Shutdown()

	log.Println("info: shutdown finished")

	bot.shutdownOnce.Do(func() { close(bot.finished) })
}
