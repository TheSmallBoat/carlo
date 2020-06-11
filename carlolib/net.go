package carlolib

import "net"

type ConnState int

type BufferedConn interface {
	net.Conn
	Flush() error
}

const (
	StateNew ConnState = iota
	StateClosed
)

type ConnStateHandler interface {
	HandleConnState(conn BufferedConn, state ConnState)
}

type ConnStateHandlerFunc func(conn BufferedConn, state ConnState)

func (fn ConnStateHandlerFunc) HandleConnState(conn BufferedConn, state ConnState) { fn(conn, state) }

var DefaultConnStateHandler ConnStateHandlerFunc = func(conn BufferedConn, state ConnState) {}

type Handler interface {
	HandleMessage(ctx *Context) error
}

type HandlerFunc func(ctx *Context) error

func (fn HandlerFunc) HandleMessage(ctx *Context) error { return fn(ctx) }

var DefaultHandler HandlerFunc = func(ctx *Context) error { return nil }

type HandShaker interface {
	Handshake(conn net.Conn) (BufferedConn, error)
}

type HandShakerFunc func(conn net.Conn) (BufferedConn, error)

func (fn HandShakerFunc) Handshake(conn net.Conn) (BufferedConn, error) { return fn(conn) }


