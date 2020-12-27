package cfg

import (
	"flag"
	"fmt"
	"log"
)

const (
	defaultMyTeamToken  = ""
	defaultMyTeamAPIURL = "https://myteam.mail.ru/bot/v1/"
	defaultMyTeamChatID = ""
)

type MyTeamConfig struct {
	MyTeamToken  *string
	MyTeamAPIURL *string
	MyTeamChatID *string
}

func NewMyTeamConfig() *MyTeamConfig {
	config := &MyTeamConfig{
		MyTeamToken:  flag.String("myteam_token", defaultMyTeamToken, "myteam bot token"),
		MyTeamAPIURL: flag.String("myteam_api_url", defaultMyTeamAPIURL, "myteam bot API url"),
		MyTeamChatID: flag.String("myteam_chat_id", defaultMyTeamChatID, "myteam chat id"),
	}

	return config
}

func (cfg *MyTeamConfig) Validate() error {
	if len(*cfg.MyTeamToken) == 0 {
		return fmt.Errorf("invalid myteam_token: %w", ErrMustNotBeEmpty)
	}
	if len(*cfg.MyTeamChatID) == 0 {
		return fmt.Errorf("invalid myteam_chat_id: %w", ErrMustNotBeEmpty)
	}

	return nil
}

func (cfg *MyTeamConfig) Print() {
	log.Printf("myteam_token: %s", *cfg.MyTeamToken)
	log.Printf("myteam_api_url: %s", *cfg.MyTeamAPIURL)
	log.Printf("myteam_chat_id: %s", *cfg.MyTeamChatID)
}
