package notifychannel

import "github.com/gibsn/duty_bot/internal/cfg"

type Type string

const (
	EmptyChannelType  Type = "empty"
	StdOutChannelType Type = "stdout" // mostly for debugging purposes
	MyTeamChannelType Type = "myteam"
)

func (ch Type) Validate() error {
	switch ch {
	case EmptyChannelType:
		fallthrough
	case StdOutChannelType:
		fallthrough
	case MyTeamChannelType:
		return nil
	}

	return cfg.ErrNotSupported
}
