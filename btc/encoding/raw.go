package encoding

import (
	"io"

	"github.com/pkg/errors"
)

type MsgRaw struct {
	Header *Header
	Body   []byte
}

func NewRawMsg(header *Header) (*MsgRaw, error) {
	return &MsgRaw{
		Header: header,
	}, nil
}

func (raw *MsgRaw) GetCommand() Command {
	return raw.Header.GetCommand()
}

func (raw *MsgRaw) Encode(writer io.Writer) error {
	return errors.New("not sending raw messages")
}

func (raw *MsgRaw) Decode(reader io.Reader) error {
	raw.Body = make([]byte, raw.Header.PayloadSize)
	_, err := io.ReadFull(reader, raw.Body)
	return errors.Wrap(err, "raw message read error")
}
