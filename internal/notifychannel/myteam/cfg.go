package myteam

import (
	"fmt"
	"log"
	"time"

	cfgUtil "github.com/gibsn/duty_bot/internal/cfg"
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

type Config struct {
	prefix string

	Token   string
	APIURL  string `mapstructure:"api_url"`
	ChatID  string `mapstructure:"chat_id"`
	Timeout time.Duration
}

func NewConfig(prefix string) Config {
	config := Config{
		prefix: prefix,
	}

	return config
}

func (cfg *Config) Validate() error {
	paramNameFactory := cfg.paramWithPrefix()

	if len(cfg.Token) == 0 {
		return fmt.Errorf(
			"%s: %w", paramNameFactory(cfgMyTeamTokenTitle), cfgUtil.ErrMustNotBeEmpty,
		)
	}
	if len(cfg.ChatID) == 0 {
		return fmt.Errorf(
			"%s: %w", paramNameFactory(cfgMyTeamChatIDTitle), cfgUtil.ErrMustNotBeEmpty,
		)
	}

	if len(cfg.APIURL) == 0 {
		cfg.APIURL = defaultMyTeamAPIURL
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = defaultMyTeamTimeout
	}

	return nil
}

func (cfg Config) Print() {
	paramNameFactory := cfg.paramWithPrefix()

	// log.Printf("myteam_token: %s", *cfg.MyTeamToken) // token is sensitive
	log.Print(fmt.Sprintf("%s: %s", paramNameFactory(cfgMyTeamAPIURLTitle), cfg.APIURL))
	log.Print(fmt.Sprintf("%s: %s", paramNameFactory(cfgMyTeamChatIDTitle), cfg.ChatID))
	log.Print(fmt.Sprintf("%s: %s", paramNameFactory(cfgMyTeamTimeoutTitle), cfg.Timeout))
}

func (cfg Config) paramWithPrefix() func(param string) string {
	return cfgUtil.ParamWithPrefix(cfg.prefix)
}
