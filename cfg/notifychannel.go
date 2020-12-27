package cfg

type NotifyChannelType string

const (
	EmptyChannelType  NotifyChannelType = "empty"
	StdOutChannelType NotifyChannelType = "stdout" // mostly for debugging purposes
	MyTeamChannelType NotifyChannelType = "myteam"
)

func (ch NotifyChannelType) Validate() error {
	switch ch {
	case EmptyChannelType:
		fallthrough
	case StdOutChannelType:
		return nil
	}

	return ErrNotSupported
}
