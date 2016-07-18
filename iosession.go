package golis

import (
	"errors"
	"net"
)

//Iosession
type Iosession struct {
	id     uint64
	serv   *ioserv
	conn   net.Conn
	closed bool
	Data   interface{}
	dataCh chan interface{}
}

func (s *Iosession) Id() uint64 {
	return s.id
}

func (s *Iosession) Conn() net.Conn {
	return s.conn
}

func (s *Iosession) dealDataCh() {
	s.serv.wg.Add(1)
	defer s.serv.wg.Done()
	var msg interface{}
	for s.serv.runnable && !s.closed {
		select {
		case msg = <-s.dataCh:
			s.serv.filterChain.msgReceived(s, msg)
		}
	}
}

func (session *Iosession) readData() {
	session.serv.wg.Add(1)
	ioBuffer := NewBuffer()
	buffer := make([]byte, 512)
	defer func() {
		if !session.closed {
			session.Close()
		}
		session.serv.wg.Done()
	}()
	var n int
	var err error
	for session.serv.runnable && !session.closed {
		n, err = session.conn.Read(buffer)
		ioBuffer.PutBytes(buffer[:n])
		if err != nil {
			session.serv.filterChain.errorCaught(session, err)
			session.Close()
			return
		}
		err = session.serv.codecer.Decode(ioBuffer, session.dataCh)
		if err != nil {
			session.serv.filterChain.errorCaught(session, err)
		}
		session.serv.filterChain.msgReceived(session, ioBuffer)
	}
}

func (session *Iosession) Write(message interface{}) error {
	if session.serv.runnable && !session.closed {
		if msg, ok := session.serv.filterChain.msgSend(session, message); ok {
			bs, err := session.serv.codecer.Encode(msg)
			if err != nil {
				return err
			}
			_, err = session.conn.Write(bs)
			if err != nil {
				session.serv.filterChain.errorCaught(session, err)
			}
		}
		return nil
	} else {
		err := errors.New("Iosession is closed")
		session.serv.filterChain.errorCaught(session, err)
		return err
	}
}

//close iosession
func (this *Iosession) Close() {
	if !this.closed {
		this.serv.filterChain.sessionClosed(this)
		this.closed = true
	}
	this.conn.Close()
}
