package cfg

import (
	"fmt"
	"log"
	"time"
)

const (
	defaultMyTeamAPIURL  = "https://myteam.mail.ru/bot/v1/"
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
	prefix string

	Token   string
	APIURL  string `mapstructure:"api_url"`
	ChatID  string `mapstructure:"chat_id"`
	Timeout time.Duration
}

func NewMyTeamConfig(prefix string) *MyTeamConfig {
	config := &MyTeamConfig{
		prefix: prefix,
	}

	return config
}

func (cfg *MyTeamConfig) Validate() error {
	paramNameFactory := cfg.paramWithPrefix()

	if len(cfg.Token) == 0 {
		return fmt.Errorf("%s: %w", paramNameFactory(cfgMyTeamTokenTitle), ErrMustNotBeEmpty)
	}
	if len(cfg.ChatID) == 0 {
		return fmt.Errorf("%s: %w", paramNameFactory(cfgMyTeamChatIDTitle), ErrMustNotBeEmpty)
	}

	if len(cfg.APIURL) == 0 {
		cfg.APIURL = defaultMyTeamAPIURL
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = defaultMyTeamTimeout
	}

	return nil
}

func (cfg MyTeamConfig) Print() {
	paramNameFactory := cfg.paramWithPrefix()

	// log.Printf("myteam_token: %s", *cfg.MyTeamToken) // token is sensitive
	log.Print(fmt.Sprintf("%s: %s", paramNameFactory(cfgMyTeamAPIURLTitle), cfg.APIURL))
	log.Print(fmt.Sprintf("%s: %s", paramNameFactory(cfgMyTeamChatIDTitle), cfg.ChatID))
	log.Print(fmt.Sprintf("%s: %s", paramNameFactory(cfgMyTeamTimeoutTitle), cfg.Timeout))
}

func (cfg MyTeamConfig) paramWithPrefix() func(param string) string {
	return paramWithPrefix(cfg.prefix)
}
