package encoding

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
)

type Header struct {
	Magic       [4]byte // MainNet: [4]byte{0xF9, 0xBE, 0xB4, 0xD9}, Regtest: [4]byte{0xFA, 0xBF, 0xB5, 0xDA}
	Command     [12]byte
	PayloadSize UInt32
	Checksum    [4]byte
}

type Network uint

const (
	NetworkMainnet Network = iota
	NetworkTestnet3
	NetworkRegtest
)

type Command string

const (
	VersionCommand Command = "version"
	VerackCommand  Command = "verack"
)

const (
	ProtocolVersion = 70015
	UserAgent       = "/MemeClient:0.0.1/"
)

type Message interface {
	Encodable

	GetCommand() Command
}

func NewHeader(network Network, command Command, payload []byte) (*Header, error) {
	var magic [4]byte
	switch network {
	case NetworkMainnet:
		magic = [4]byte{0xF9, 0xBE, 0xB4, 0xD9}
	case NetworkTestnet3:
		magic = [4]byte{0x0B, 0x11, 0x09, 0x07}
	case NetworkRegtest:
		magic = [4]byte{0xFA, 0xBF, 0xB5, 0xDA}
	default:
		return nil, fmt.Errorf("unknown network: %d", network)
	}

	commandBytes := [12]byte{}
	copy(commandBytes[:], command)

	return &Header{
		Magic:       magic,
		Command:     commandBytes,
		PayloadSize: UInt32(len(payload)),
		Checksum:    calculateChecksum(payload),
	}, nil
}

func (header *Header) Encode(writer io.Writer) error {
	return encode(writer,
		step("magic", RawBytes(header.Magic[:])),
		step("command", RawBytes(header.Command[:])),
		step("payload size", &header.PayloadSize),
		step("checksum", RawBytes(header.Checksum[:])),
	)
}

func (header *Header) GetCommand() Command {
	name := string(bytes.TrimRight(header.Command[:], "\x00"))
	return Command(name)
}

func (header *Header) Decode(reader io.Reader) error {
	return decode(reader,
		step("magic", RawBytes(header.Magic[:])),
		step("command", RawBytes(header.Command[:])),
		step("payload size", &header.PayloadSize),
		step("checksum", RawBytes(header.Checksum[:])),
	)
}

// Builds a header and sends it and the message to the writer.
func SendMessage(network Network, message Message, writer io.Writer) error {
	msgBuf := bytes.NewBuffer(nil)

	err := message.Encode(msgBuf)
	if err != nil {
		return fmt.Errorf("error encoding message: %w", err)
	}
	header, err := NewHeader(network, message.GetCommand(), msgBuf.Bytes())
	if err != nil {
		return fmt.Errorf("error creating header: %w", err)
	}

	err = header.Encode(writer)
	if err != nil {
		return fmt.Errorf("error encoding header: %w", err)
	}
	_, err = msgBuf.WriteTo(writer)
	if err != nil {
		return fmt.Errorf("error encoding message: %w", err)
	}
	return nil
}

func createMessage(header *Header) (Message, error) {
	switch header.GetCommand() {
	case VersionCommand:
		return &MsgVersion{}, nil
	case VerackCommand:
		return &MsgVerack{}, nil
	default:
		return NewRawMsg(header)
	}
}

func ReceiveMessage(reader io.Reader) (*Header, Message, error) {
	header := &Header{}
	err := header.Decode(reader)
	if err != nil {
		return nil, nil, fmt.Errorf("error decoding header: %w", err)
	}

	msg, err := createMessage(header)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating message: %w", err)
	}
	err = msg.Decode(reader)
	if err != nil {
		return nil, nil, fmt.Errorf("error decoding message: %w", err)
	}
	return header, msg, nil
}

func calculateChecksum(payload []byte) [4]byte {
	firstSHA := sha256.Sum256(payload)
	secondSHA := sha256.Sum256(firstSHA[:])
	var checksum [4]byte
	copy(checksum[:], secondSHA[:4])
	return checksum
}
