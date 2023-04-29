package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
)

type message struct {
	typ byte
	a   int32
	b   int32
}

func (m message) String() string {
	return fmt.Sprintf("%c %d %d", m.typ, m.a, m.b)
}

func decodeMessage(buf []byte) (message, error) {
	var err error
	m := message{}
	m.typ = buf[0]
	m.a, err = readInt32(buf[1:5])
	if err != nil {
		return m, err
	}
	m.b, err = readInt32(buf[5:9])
	if err != nil {
		return m, err
	}
	return m, nil
}

func readInt32(buf []byte) (int32, error) {
	var n int32
	if err := binary.Read(bytes.NewReader(buf), binary.BigEndian, &n); err != nil {
		return 0, err
	}
	return n, nil
}

type price struct {
	time  int32
	price int32
}

type session struct {
	prices []price
}

func newSession() *session {
	return &session{prices: []price{}}
}

func (s *session) handleMessage(buf []byte, w io.Writer) error {
	m, err := decodeMessage(buf)
	if err != nil {
		return err
	}
	switch m.typ {
	case 'I':
		return s.handleInsert(m, w)
	case 'Q':
		return s.handleQuery(m, w)
	}
	return fmt.Errorf("unknown type %v", m.typ)
}

func (s *session) handleInsert(m message, w io.Writer) error {
	s.prices = append(s.prices, price{time: m.a, price: m.b})
	return nil
}

func (s *session) handleQuery(m message, w io.Writer) error {
	var mean, sum, count int64
	for _, p := range s.prices {
		if p.time >= m.a && p.time <= m.b {
			sum += int64(p.price)
			count++
		}
	}
	if count > 0 {
		mean = sum / count
	}
	binary.Write(w, binary.BigEndian, int32(mean))
	return nil
}

func main() {
	log.Fatal(listen(":8080"))
}

func listen(addr string) error {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer l.Close()
	for {
		if err := acceptAndHandle(l); err != nil {
			return err
		}
	}
}

func acceptAndHandle(l net.Listener) error {
	conn, err := l.Accept()
	if err != nil {
		return err
	}
	go handleConnection(conn)
	return nil
}

func handleConnection(c net.Conn) {
	defer c.Close()

	s := newSession()
	buf := make([]byte, 9)
	for {
		_, err := io.ReadFull(c, buf)
		if err != nil {
			return
		}
		if err := s.handleMessage(buf, c); err != nil {
			return
		}
	}
}
