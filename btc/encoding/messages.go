package encoding

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"net"
	"time"
)

type Header struct {
	Magic       [4]byte // MainNet: [4]byte{0xF9, 0xBE, 0xB4, 0xD9}, Regtest: [4]byte{0xFA, 0xBF, 0xB5, 0xDA}
	Command     [12]byte
	PayloadSize UInt32
	Checksum    [4]byte
}

type MsgVersion struct {
	network Network

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

type Network uint

const (
	NetworkMainnet Network = iota
	NetworkTestnet3
	NetworkRegtest
)

type Command string

const (
	VersionCommand Command = "version"
)

const (
	ProtocolVersion = 70015
	UserAgent       = "/MemeClient:0.0.1/"
)

type NetworkAddress struct {
	Time     UInt32
	Services UInt64
	IP       IP
	Port     PortNumber
}

type Services UInt64

const (
	ServicesNone               Services = 0
	ServicesNodeNetwork        Services = 1
	ServicesNodeGetUTXO        Services = 2
	ServicesNodeBloom          Services = 4
	ServicesNodeWitness        Services = 8
	ServicesNodeXThin          Services = 16
	ServicesNodeCompactFilters Services = 64
	ServicesNodeNetworkLimited Services = 1024
)

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

func NewVersionMsg(
	network Network,
	timestamp time.Time,
	services Services,
	addrRecv *NetworkAddress,
	addrFrom *NetworkAddress,
	nonce uint64,
	startHeight uint32,
) (*MsgVersion, error) {
	version := &MsgVersion{
		network:     network,
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

func NewIP4Address(services Services, addr string) (*NetworkAddress, error) {
	resolved, err := net.ResolveTCPAddr("tcp4", addr)
	if err != nil {
		return nil, fmt.Errorf("invalid address %s: %w", addr, err)
	}
	return &NetworkAddress{
		Time:     UInt32(0),
		Services: UInt64(services),
		IP:       IP(resolved.IP),
		Port:     PortNumber(resolved.Port),
	}, nil
}

func (header *Header) Encode(writer io.Writer) error {
	return encode(writer,
		step("magic", RawBytes(header.Magic[:])),
		step("command", RawBytes(header.Command[:])),
		step("payload size", header.PayloadSize),
		step("checksum", RawBytes(header.Checksum[:])),
	)
}

func (version *MsgVersion) Encode(writer io.Writer) error {
	msgBuf := bytes.NewBuffer(nil)

	err := encode(msgBuf,
		step("version", version.Version),
		step("services", version.Services),
		step("timestamp", version.Timestamp),
		step("addr_recv", version.AddrRecv),
		step("addr_from", version.AddrFrom),
		step("nonce", version.Nonce),
		step("user_agent", version.UserAgent),
		step("start_height", version.StartHeight),
		step("relay", version.Relay),
	)
	if err != nil {
		return fmt.Errorf("error serializing message fields: %w", err)
	}

	header, err := NewHeader(version.network, VersionCommand, msgBuf.Bytes())
	if err != nil {
		return fmt.Errorf("error creating header: %w", err)
	}

	err = header.Encode(writer)
	if err != nil {
		return fmt.Errorf("error serializing header: %w", err)
	}
	_, err = msgBuf.WriteTo(writer)
	if err != nil {
		return fmt.Errorf("error serializing message: %w", err)
	}
	return nil
}

func (addr NetworkAddress) Encode(writer io.Writer) error {
	steps := []encodeStep{
		step("time", addr.Time),
		step("services", addr.Services),
		step("ip", addr.IP),
		step("port", addr.Port),
	}
	// Address timestamp is not used and not sent in version messages
	if addr.Time > 0 {
		return encode(writer, steps...)
	}
	return encode(writer, steps[1:]...)
}

func calculateChecksum(payload []byte) [4]byte {
	firstSHA := sha256.Sum256(payload)
	secondSHA := sha256.Sum256(firstSHA[:])
	var checksum [4]byte
	copy(checksum[:], secondSHA[:4])
	return checksum
}
