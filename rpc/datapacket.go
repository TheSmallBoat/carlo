package rpc

import (
	"io"

	"github.com/lithdew/bytesutil"
)

type DataPacket struct {
	StreamID uint32 // stream id
	Data     []byte // payload
}

func (p DataPacket) AppendTo(dst []byte) []byte {
	dst = bytesutil.AppendUint32BE(dst, p.StreamID)
	dst = bytesutil.AppendUint16BE(dst, uint16(len(p.Data)))
	dst = append(dst, p.Data...)
	return dst
}

func UnmarshalDataPacket(buf []byte) (DataPacket, error) {
	var packet DataPacket
	if len(buf) < 4+2 {
		return packet, io.ErrUnexpectedEOF
	}
	packet.StreamID, buf = bytesutil.Uint32BE(buf[:4]), buf[4:]
	var size uint16
	size, buf = bytesutil.Uint16BE(buf[:2]), buf[2:]
	if uint16(len(buf)) < size {
		return packet, io.ErrUnexpectedEOF
	}
	packet.Data, buf = buf[:size], buf[size:]
	return packet, nil
}
