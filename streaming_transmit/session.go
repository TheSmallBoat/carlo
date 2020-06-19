package streaming_transmit

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/lithdew/bytesutil"
	"github.com/oasisprotocol/ed25519"
	"github.com/oasisprotocol/ed25519/extra/x25519"
	"golang.org/x/crypto/blake2b"
)

var _ BufferedConn = (*SessionConn)(nil)

// SessionConn is not safe for concurrent use. It decrypts on reads and encrypts on writes
// via a provided cipher.AEAD suite for a given conn that implements net.Conn. It assumes
// all packets sent/received are to be prefixed with a 32-bit unsigned integer that
// designates the length of each individual packet.
type SessionConn struct {
	suite cipher.AEAD
	conn  net.Conn

	bw *bufio.Writer
	br *bufio.Reader

	rb []byte // read buffer
	wb []byte // write buffer
	wn uint64 // write nonce
	rn uint64 // read nonce
}

func NewSessionConn(suite cipher.AEAD, conn net.Conn) *SessionConn {
	return &SessionConn{
		suite: suite,
		conn:  conn,

		bw: bufio.NewWriter(conn),
		br: bufio.NewReader(conn),
	}
}
func (s *SessionConn) Read(b []byte) (int, error) {
	var err error
	s.rb, err = ReadSized(s.rb[:0], s.br, cap(b))
	if err != nil {
		return 0, err
	}

	s.rb = bytesutil.ExtendSlice(s.rb, len(s.rb)+s.suite.NonceSize())
	for i := len(s.rb) - s.suite.NonceSize(); i < len(s.rb); i++ {
		s.rb[i] = 0
	}
	binary.BigEndian.PutUint64(s.rb[len(s.rb)-s.suite.NonceSize():], s.rn)
	s.rn++

	s.rb, err = s.suite.Open(
		s.rb[:0],
		s.rb[len(s.rb)-s.suite.NonceSize():],
		s.rb[:len(s.rb)-s.suite.NonceSize()],
		nil,
	)
	if err != nil {
		return 0, err
	}
	return copy(b, s.rb), err
}

func (s *SessionConn) Write(b []byte) (int, error) {
	s.wb = bytesutil.ExtendSlice(s.wb, s.suite.NonceSize()+len(b)+s.suite.Overhead())
	binary.BigEndian.PutUint64(s.wb[:8], s.wn)
	for i := 8; i < s.suite.NonceSize(); i++ {
		s.wb[i] = 0
	}
	s.wn++

	s.wb = s.suite.Seal(
		s.wb[s.suite.NonceSize():s.suite.NonceSize()],
		s.wb[:s.suite.NonceSize()],
		b,
		nil,
	)

	err := WriteSized(s.bw, s.wb)
	if err != nil {
		return 0, err
	}

	return len(s.wb), nil
}

func (s *SessionConn) Flush() error { return s.bw.Flush() }

func (s *SessionConn) Close() error                       { return s.conn.Close() }
func (s *SessionConn) LocalAddr() net.Addr                { return s.conn.LocalAddr() }
func (s *SessionConn) RemoteAddr() net.Addr               { return s.conn.RemoteAddr() }
func (s *SessionConn) SetDeadline(t time.Time) error      { return s.conn.SetDeadline(t) }
func (s *SessionConn) SetReadDeadline(t time.Time) error  { return s.conn.SetReadDeadline(t) }
func (s *SessionConn) SetWriteDeadline(t time.Time) error { return s.conn.SetWriteDeadline(t) }

// Session is not safe for concurrent use.
type Session struct {
	suite     cipher.AEAD
	ourPub    []byte
	ourPriv   []byte
	theirPub  []byte
	sharedKey []byte
}

func NewSession() (Session, error) {
	var session Session

	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		return session, err
	}

	sessionPub, ok := x25519.EdPublicKeyToX25519(publicKey)
	if !ok {
		return session, errors.New("unable to derive ed25519 key to x25519 key")
	}

	sessionPriv := x25519.EdPrivateKeyToX25519(privateKey)

	session.ourPub = sessionPub
	session.ourPriv = sessionPriv

	return session, nil
}

func (s *Session) Suite() cipher.AEAD {
	return s.suite
}

func (s *Session) SharedKey() []byte {
	return s.sharedKey
}

func (s *Session) Write(conn net.Conn) error {
	err := Write(conn, s.ourPub)
	if err != nil {
		return fmt.Errorf("failed to write session public key: %w", err)
	}
	return nil
}

func (s *Session) Read(conn net.Conn) error {
	publicKey, err := Read(make([]byte, x25519.PointSize), conn)
	if err != nil {
		return fmt.Errorf("failed to read peer session public key: %w", err)
	}
	s.theirPub = publicKey
	return nil
}

func (s *Session) Establish() error {
	if s.theirPub == nil {
		return errors.New("did not read peer session public key yet")
	}
	sharedKey, err := x25519.X25519(s.ourPriv, s.theirPub)
	if err != nil {
		return fmt.Errorf("failed to derive shared session key: %w", err)
	}
	derivedKey := blake2b.Sum256(sharedKey)
	block, err := aes.NewCipher(derivedKey[:])
	if err != nil {
		return fmt.Errorf("failed to init aes cipher: %w", err)
	}
	suite, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("failed to init aead suite: %w", err)
	}
	s.sharedKey = derivedKey[:]
	s.suite = suite
	return nil
}

func (s *Session) DoClient(conn net.Conn) error {
	err := s.Write(conn)
	if err == nil {
		err = s.Read(conn)
	}
	if err == nil {
		err = s.Establish()
	}
	return err
}

func (s *Session) DoServer(conn net.Conn) error {
	err := s.Read(conn)
	if err == nil {
		err = s.Write(conn)
	}
	if err == nil {
		err = s.Establish()
	}
	return err
}
