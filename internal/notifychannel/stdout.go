package notifychannel

import (
	"fmt"
	"os"
)

// StdOutNotifyChannel is a notify channel that prints out every
// notification to stdout. Mostly used for debugging
type StdOutNotifyChannel struct {
}

func (StdOutNotifyChannel) Send(person string) error {
	_, err := fmt.Fprintln(os.Stdout, person)

	return err
}

func (StdOutNotifyChannel) Shutdown() error {
	return nil
}
