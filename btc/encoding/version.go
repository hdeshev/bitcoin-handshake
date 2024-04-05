package encoding

import (
	"fmt"
	"io"
	"time"
)

type MsgVersion struct {
	Version     UInt32 // 70015
	Services    Services
	Timestamp   UInt64
	AddrRecv    NetworkAddress
	AddrFrom    NetworkAddress
	Nonce       UInt64
	UserAgent   VarStr
	StartHeight UInt32
	Relay       UInt8
}

func NewVersionMsg(
	timestamp time.Time,
	services Services,
	addrRecv *NetworkAddress,
	addrFrom *NetworkAddress,
	nonce uint64,
	startHeight uint32,
) (*MsgVersion, error) {
	version := &MsgVersion{
		Version:     ProtocolVersion,
		UserAgent:   VarStr(UserAgent),
		Services:    services,
		Timestamp:   UInt64(timestamp.Unix()),
		AddrRecv:    *addrRecv,
		AddrFrom:    *addrFrom,
		Nonce:       UInt64(nonce),
		StartHeight: UInt32(startHeight),
		Relay:       UInt8(0),
	}
	return version, nil
}

func (version *MsgVersion) GetCommand() Command {
	return VersionCommand
}

func (version *MsgVersion) Encode(writer io.Writer) error {
	err := encode(writer,
		step("version", &version.Version),
		step("services", &version.Services),
		step("timestamp", &version.Timestamp),
		step("addr_recv", &version.AddrRecv),
		step("addr_from", &version.AddrFrom),
		step("nonce", &version.Nonce),
		step("user_agent", &version.UserAgent),
		step("start_height", &version.StartHeight),
		step("relay", &version.Relay),
	)
	if err != nil {
		return fmt.Errorf("error encoding version fields: %w", err)
	}
	return nil
}

func (version *MsgVersion) Decode(reader io.Reader) error {
	err := decode(reader,
		step("version", &version.Version),
		step("services", &version.Services),
		step("timestamp", &version.Timestamp),
		step("addr_recv", &version.AddrRecv),
		step("addr_from", &version.AddrFrom),
		step("nonce", &version.Nonce),
		step("user_agent", &version.UserAgent),
		step("start_height", &version.StartHeight),
		step("relay", &version.Relay),
	)
	if err != nil {
		return fmt.Errorf("error decoding version fields: %w", err)
	}
	return nil
}
