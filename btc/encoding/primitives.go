package encoding

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"

	"github.com/pkg/errors"
)

var (
	le = binary.LittleEndian
	be = binary.BigEndian
)

type PortNumber uint16

func (ui *PortNumber) Encode(writer io.Writer) error {
	numBuf16 := [2]byte{}
	be.PutUint16(numBuf16[:], uint16(*ui))
	_, err := writer.Write(numBuf16[:])
	return errors.Wrap(err, "port number write error")
}

func (ui *PortNumber) Decode(reader io.Reader) error {
	numBuf16 := [2]byte{}
	_, err := io.ReadFull(reader, numBuf16[:])
	if err != nil {
		return errors.Wrap(err, "uint16 read error")
	}

	*ui = PortNumber(be.Uint16(numBuf16[:]))
	return nil
}

type UInt16 uint16

func (ui *UInt16) Encode(writer io.Writer) error {
	numBuf16 := [2]byte{}
	le.PutUint16(numBuf16[:], uint16(*ui))
	_, err := writer.Write(numBuf16[:])
	return errors.Wrap(err, "uint16 write error")
}

func (ui *UInt16) Decode(reader io.Reader) error {
	numBuf16 := [2]byte{}
	_, err := io.ReadFull(reader, numBuf16[:])
	if err != nil {
		return errors.Wrap(err, "uint16 read error")
	}

	*ui = UInt16(le.Uint16(numBuf16[:]))
	return nil
}

type UInt32 uint32

func (ui *UInt32) Encode(writer io.Writer) error {
	numBuf32 := [4]byte{}
	le.PutUint32(numBuf32[:], uint32(*ui))
	_, err := writer.Write(numBuf32[:])
	return errors.Wrap(err, "uint32 write error")
}

func (ui *UInt32) Decode(reader io.Reader) error {
	numBuf32 := [4]byte{}
	_, err := io.ReadFull(reader, numBuf32[:])
	if err != nil {
		return errors.Wrap(err, "uint32 read error")
	}

	*ui = UInt32(le.Uint32(numBuf32[:]))
	return nil
}

type UInt64 uint64

func (ui *UInt64) Encode(writer io.Writer) error {
	numBuf64 := [8]byte{}
	le.PutUint64(numBuf64[:], uint64(*ui))
	_, err := writer.Write(numBuf64[:])
	return errors.Wrap(err, "uint64 write error")
}

func (ui *UInt64) Decode(reader io.Reader) error {
	numBuf64 := [8]byte{}
	_, err := io.ReadFull(reader, numBuf64[:])
	if err != nil {
		return errors.Wrap(err, "uint64 read error")
	}

	*ui = UInt64(le.Uint64(numBuf64[:]))
	return nil
}

func (s *Services) Encode(writer io.Writer) error {
	ui := UInt64(*s)
	return (&ui).Encode(writer)
}

func (s *Services) Decode(reader io.Reader) error {
	ui := UInt64(0)
	err := (&ui).Decode(reader)
	if err != nil {
		return errors.Wrap(err, "services read error")
	}
	*s = Services(ui)
	return nil
}

type IP net.IP

func (ip *IP) Encode(writer io.Writer) error {
	_, err := writer.Write(net.IP(*ip).To16())
	return errors.Wrap(err, "ip address write error")
}

func (ip *IP) Decode(reader io.Reader) error {
	buf := [16]byte{}
	_, err := io.ReadFull(reader, buf[:])
	*ip = IP(buf[:])
	return errors.Wrap(err, "ip read error")
}

type RawBytes []byte

func (b RawBytes) Encode(writer io.Writer) error {
	_, err := writer.Write(b)
	return errors.Wrap(err, "raw bytes write error")
}

func (b RawBytes) Decode(reader io.Reader) error {
	_, err := io.ReadFull(reader, b)
	return errors.Wrap(err, "raw bytes read error")
}

type UInt8 uint8

func (ui *UInt8) Encode(writer io.Writer) error {
	numBuf8 := [1]byte{uint8(*ui)}
	_, err := writer.Write(numBuf8[:])
	return errors.Wrap(err, "uint8 write error")
}

func (ui *UInt8) Decode(reader io.Reader) error {
	numBuf8 := [1]byte{}
	_, err := io.ReadFull(reader, numBuf8[:])
	if err != nil {
		return errors.Wrap(err, "uint8 read error")
	}

	*ui = UInt8(numBuf8[0])
	return nil
}

type VarInt uint64

const (
	varIntMax1Byte  = 0xFD
	varIntMax2Bytes = 0xFFFF
	varIntMax4Bytes = 0xFFFFFFFF

	varIntPrefix2Bytes = 0xFD
	varIntPrefix4Bytes = 0xFE
	varIntPrefix8Bytes = 0xFF
)

func (vi *VarInt) Encode(writer io.Writer) error {
	i := uint64(*vi)
	switch {
	case i < varIntMax1Byte:
		ui := UInt8(i)
		return (&ui).Encode(writer)
	case i < varIntMax2Bytes:
		ui := UInt16(i)
		prefix := UInt8(varIntPrefix2Bytes)
		return encode(
			writer,
			step("varint_prefix", &prefix),
			step("varint_num", &ui),
		)
	case i < varIntMax4Bytes:
		ui := UInt32(i)
		prefix := UInt8(varIntPrefix4Bytes)
		return encode(
			writer,
			step("varint_prefix", &prefix),
			step("varint_num", &ui),
		)
	default:
		ui := UInt64(i)
		prefix := UInt8(varIntPrefix8Bytes)
		return encode(
			writer,
			step("varint_prefix", &prefix),
			step("varint_num", &ui),
		)
	}
}

//nolint:funlen //many decode cases
func (vi *VarInt) Decode(reader io.Reader) error {
	widthPrefix := UInt8(0)
	err := (&widthPrefix).Decode(reader)
	if err != nil {
		return errors.Wrap(err, "varint width read error")
	}
	switch {
	case widthPrefix < varIntMax1Byte:
		*vi = VarInt(widthPrefix)
		return nil
	case widthPrefix == varIntPrefix2Bytes:
		ui := UInt16(0)
		err := (&ui).Decode(reader)
		if err != nil {
			return errors.Wrap(err, "varint num read error")
		}
		*vi = VarInt(ui)
		return nil
	case widthPrefix == varIntPrefix4Bytes:
		ui := UInt32(0)
		err := (&ui).Decode(reader)
		if err != nil {
			return errors.Wrap(err, "varint num read error")
		}
		*vi = VarInt(ui)
		return nil
	case widthPrefix == varIntPrefix8Bytes:
		ui := UInt64(0)
		err := (&ui).Decode(reader)
		if err != nil {
			return errors.Wrap(err, "varint num read error")
		}
		*vi = VarInt(ui)
		return nil
	default:
		return fmt.Errorf("invalid varint prefix: %d", widthPrefix)
	}
}

type VarStr string

func (vs *VarStr) Encode(writer io.Writer) error {
	s := string(*vs)
	vi := VarInt(len(s))
	return encode(
		writer,
		step("varint_prefix", &vi),
		step("varint_num", RawBytes([]byte(s))),
	)
}

func (vs *VarStr) Decode(reader io.Reader) error {
	s := string(*vs)
	vi := VarInt(len(s))
	err := (&vi).Decode(reader)
	if err != nil {
		return errors.Wrap(err, "var_str length read error")
	}
	buf := make([]byte, uint64(vi))
	err = RawBytes(buf).Decode(reader)
	if err != nil {
		return errors.Wrap(err, "var_str contents read error")
	}
	*vs = VarStr(string(buf))
	return nil
}

type NetworkAddress struct {
	Time     UInt32
	Services UInt64
	IP       IP
	Port     PortNumber
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

func (addr *NetworkAddress) Encode(writer io.Writer) error {
	steps := []*encodeStep{
		step("time", &addr.Time),
		step("services", &addr.Services),
		step("ip", &addr.IP),
		step("port", &addr.Port),
	}
	// Address timestamp is not used and not sent in version messages
	if addr.Time > 0 {
		return encode(writer, steps...)
	}
	return encode(writer, steps[1:]...)
}

func (addr *NetworkAddress) Decode(reader io.Reader) error {
	return decode(reader,
		step("services", &addr.Services),
		step("ip", &addr.IP),
		step("port", &addr.Port),
	)
}
