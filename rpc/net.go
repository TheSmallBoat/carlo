package rpc

import (
	"net"
	"strconv"
)

const ChunkSize = 2048

type OpCode = uint8

const (
	OpCodeHandshake OpCode = iota
	OpCodeServiceRequest
	OpCodeServiceResponse
	OpCodeData
	OpCodeFindNodeRequest
	OpCodeFindNodeResponse
)

type BindFunc func() (net.Listener, error)

func BindTCPAnyPort() BindFunc {
	return func() (net.Listener, error) { return net.Listen("tcp", ":0") }
}

func BindTCP(addr string) BindFunc {
	return func() (net.Listener, error) { return net.Listen("tcp", addr) }
}

func BindTCPv4(addr string) BindFunc {
	return func() (net.Listener, error) { return net.Listen("tcp4", addr) }
}

func BindTCPv6(addr string) BindFunc {
	return func() (net.Listener, error) { return net.Listen("tcp6", addr) }
}

func HostAddr(host net.IP, port uint16) string {
	h := ""
	if len(host) > 0 {
		h = host.String()
	}
	p := strconv.FormatUint(uint64(port), 10)
	return net.JoinHostPort(h, p)
}
