package encoding

import (
	"encoding/binary"
	"io"
	"net"

	"github.com/pkg/errors"
)

var (
	le = binary.LittleEndian
	be = binary.BigEndian
)

type PortNumber uint16

func (ui PortNumber) Encode(writer io.Writer) error {
	numBuf16 := [2]byte{}
	be.PutUint16(numBuf16[:], uint16(ui))
	_, err := writer.Write(numBuf16[:])
	return errors.Wrap(err, "port number write error")
}

type UInt16 uint16

func (ui UInt16) Encode(writer io.Writer) error {
	numBuf16 := [2]byte{}
	le.PutUint16(numBuf16[:], uint16(ui))
	_, err := writer.Write(numBuf16[:])
	return errors.Wrap(err, "uint16 write error")
}

type UInt32 uint32

func (ui UInt32) Encode(writer io.Writer) error {
	numBuf32 := [4]byte{}
	le.PutUint32(numBuf32[:], uint32(ui))
	_, err := writer.Write(numBuf32[:])
	return errors.Wrap(err, "uint32 write error")
}

type UInt64 uint64

func (ui UInt64) Encode(writer io.Writer) error {
	numBuf64 := [8]byte{}
	le.PutUint64(numBuf64[:], uint64(ui))
	_, err := writer.Write(numBuf64[:])
	return errors.Wrap(err, "uint64 write error")
}

func (s Services) Encode(writer io.Writer) error {
	return UInt64(s).Encode(writer)
}

type IP net.IP

func (ip IP) Encode(writer io.Writer) error {
	_, err := writer.Write(net.IP(ip).To16())
	return errors.Wrap(err, "ip address write error")
}

type RawBytes []byte

func (b RawBytes) Encode(writer io.Writer) error {
	_, err := writer.Write(b)
	return errors.Wrap(err, "raw bytes write error")
}

type UInt8 uint8

func (ui UInt8) Encode(writer io.Writer) error {
	numBuf8 := [1]byte{uint8(ui)}
	_, err := writer.Write(numBuf8[:])
	return errors.Wrap(err, "uint8 write error")
}

type VarInt uint64

func (vi VarInt) Encode(writer io.Writer) error {
	const (
		max1Byte  = 0xFD
		max2Bytes = 0xFFFF
		max4Bytes = 0xFFFFFFFF

		prefix2Bytes = 0xFD
		prefix4Bytes = 0xFE
		prefix8Bytes = 0xFF
	)

	i := uint64(vi)
	switch {
	case i < max1Byte:
		return UInt8(i).Encode(writer)
	case i < max2Bytes:
		return encode(
			writer,
			step("varint_prefix", UInt8(prefix2Bytes)),
			step("varint_num", UInt16(i)),
		)
	case i < max4Bytes:
		return encode(
			writer,
			step("varint_prefix", UInt8(prefix4Bytes)),
			step("varint_num", UInt32(i)),
		)
	default:
		return encode(
			writer,
			step("varint_prefix", UInt8(prefix8Bytes)),
			step("varint_num", UInt64(i)),
		)
	}
}

type VarStr string

func (vs VarStr) Encode(writer io.Writer) error {
	s := string(vs)
	return encode(
		writer,
		step("varint_prefix", VarInt(len(s))),
		step("varint_num", RawBytes([]byte(s))),
	)
}
