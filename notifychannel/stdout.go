package notifychannel

import (
	"fmt"
	"os"
)

type StdOutNotifyChannel struct {
}

func (StdOutNotifyChannel) Send(person string) error {
	_, err := fmt.Fprintln(os.Stdout, person)

	return err
}
