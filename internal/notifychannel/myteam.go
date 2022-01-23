package notifychannel

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gibsn/duty_bot/internal/cfg"

	botgolang "github.com/mail-ru-im/bot-golang" //nolint: goimports
)

type MyTeamNotifyChannel struct {
	bot *botgolang.Bot

	chatID string
}

func NewMyTeamNotifyChannel(config *cfg.MyTeamConfig) (*MyTeamNotifyChannel, error) {
	// currently myteam library has no option to provide a timeout
	http.DefaultClient.Timeout = config.Timeout

	bot, err := botgolang.NewBot(config.Token, botgolang.BotApiURL(config.APIURL))
	if err != nil {
		return nil, fmt.Errorf("could not create myteam bot: %w", err)
	}

	log.Printf("info: myteam: connected to bot '%s'", bot.Info.Nick)

	return &MyTeamNotifyChannel{
		bot:    bot,
		chatID: config.ChatID,
	}, nil
}

func (ch *MyTeamNotifyChannel) Send(text string) error {
	msg := ch.bot.NewTextMessage(ch.chatID, text)

	if err := msg.Send(); err != nil {
		return fmt.Errorf("could not send message to chat '%s': %w", ch.chatID, err)
	}

	if err := msg.Pin(); err != nil {
		log.Printf("error: myteam: could not pin message in chat '%s': %v", ch.chatID, err)
	}

	return nil
}

func (ch *MyTeamNotifyChannel) Shutdown() error {
	return nil
}
