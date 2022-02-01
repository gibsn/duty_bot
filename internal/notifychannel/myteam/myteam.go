package myteam

import (
	"fmt"
	"log"
	"net/http"

	botgolang "github.com/mail-ru-im/bot-golang"
)

type NotifyChannel struct {
	bot *botgolang.Bot

	chatID string
}

func NewNotifyChannel(config Config) (*NotifyChannel, error) {
	// currently myteam library has no option to provide a timeout
	http.DefaultClient.Timeout = config.Timeout

	bot, err := botgolang.NewBot(config.Token, botgolang.BotApiURL(config.APIURL))
	if err != nil {
		return nil, fmt.Errorf("could not create myteam bot: %w", err)
	}

	log.Printf("info: myteam: connected to bot '%s'", bot.Info.Nick)

	return &NotifyChannel{
		bot:    bot,
		chatID: config.ChatID,
	}, nil
}

func (ch *NotifyChannel) Send(text string) error {
	msg := ch.bot.NewTextMessage(ch.chatID, text)

	if err := msg.Send(); err != nil {
		return fmt.Errorf("could not send message to chat '%s': %w", ch.chatID, err)
	}

	if err := msg.Pin(); err != nil {
		log.Printf("error: myteam: could not pin message in chat '%s': %v", ch.chatID, err)
	}

	return nil
}

func (ch *NotifyChannel) Shutdown() error {
	return nil
}
