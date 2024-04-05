package encoding

import (
	"io"
)

type MsgVerack struct {
	// verack has no body and contains just a header
}

func NewVerackMsg() (*MsgVerack, error) {
	return &MsgVerack{}, nil
}

func (version *MsgVerack) GetCommand() Command {
	return VerackCommand
}

func (version *MsgVerack) Encode(writer io.Writer) error {
	return nil
}

func (version *MsgVerack) Decode(reader io.Reader) error {
	return nil
}
