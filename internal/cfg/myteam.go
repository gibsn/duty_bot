package cfg

import (
	"flag"
	"fmt"
	"log"
	"time"
)

const (
	defaultMyTeamToken   = ""
	defaultMyTeamAPIURL  = "https://myteam.mail.ru/bot/v1/"
	defaultMyTeamChatID  = ""
	defaultMyTeamTimeout = 5 * time.Second
)

const (
	cfgMyTeamPrefix = "myteam"

	cfgMyTeamTokenTitle   = cfgMyTeamPrefix + ".token"
	cfgMyTeamAPIURLTitle  = cfgMyTeamPrefix + ".api_url"
	cfgMyTeamChatIDTitle  = cfgMyTeamPrefix + ".chat_id"
	cfgMyTeamTimeoutTitle = cfgMyTeamPrefix + ".timeout"
)

type MyTeamConfig struct {
	MyTeamToken   *string
	MyTeamAPIURL  *string
	MyTeamChatID  *string
	MyTeamTimeout *time.Duration
}

func NewMyTeamConfig() *MyTeamConfig {
	config := &MyTeamConfig{
		MyTeamToken:   flag.String(cfgMyTeamTokenTitle, defaultMyTeamToken, "myteam bot token"),
		MyTeamAPIURL:  flag.String(cfgMyTeamAPIURLTitle, defaultMyTeamAPIURL, "myteam bot API url"),
		MyTeamChatID:  flag.String(cfgMyTeamChatIDTitle, defaultMyTeamChatID, "myteam chat id"),
		MyTeamTimeout: flag.Duration(cfgMyTeamTimeoutTitle, defaultMyTeamTimeout, "myteam timeout"),
	}

	return config
}

func (cfg *MyTeamConfig) Validate() error {
	if len(*cfg.MyTeamToken) == 0 {
		return fmt.Errorf("invalid %s: %w", cfgMyTeamTokenTitle, ErrMustNotBeEmpty)
	}
	if len(*cfg.MyTeamChatID) == 0 {
		return fmt.Errorf("invalid %s: %w", cfgMyTeamChatIDTitle, ErrMustNotBeEmpty)
	}

	return nil
}

func (cfg *MyTeamConfig) Print() {
	// log.Printf("myteam_token: %s", *cfg.MyTeamToken) // token is sensitive
	log.Print(cfgMyTeamAPIURLTitle+":", *cfg.MyTeamAPIURL)
	log.Print(cfgMyTeamChatIDTitle+":", *cfg.MyTeamChatID)
	log.Print(cfgMyTeamTimeoutTitle+":", *cfg.MyTeamTimeout)
}
