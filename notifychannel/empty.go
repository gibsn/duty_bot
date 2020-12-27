package notifychannel

type EmptyNotifyChannel struct {
}

func (EmptyNotifyChannel) Send(_ string) error {
	return nil
}

func (EmptyNotifyChannel) Shutdown() error {
	return nil
}
