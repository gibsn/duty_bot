package notifychannel

import (
	"io"
)

type Pipe struct {
	*io.PipeReader

	w *io.PipeWriter
}

func NewPipe() *Pipe {
	r, w := io.Pipe()

	return &Pipe{r, w}
}

func (p *Pipe) Send(text string) error {
	_, err := io.WriteString(p.w, text+"\n")

	return err
}
