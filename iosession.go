package golis

import (
	"errors"
	"net"
)

//Iosession
type Iosession struct {
	serv   *ioserv
	conn   net.Conn
	closed bool
	Data   interface{}
}

func (this *Iosession) GetConn() net.Conn {
	return this.conn
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
	for session.serv.runnable && !session.closed {
		n, err := session.conn.Read(buffer)
		ioBuffer.PutBytes(buffer[:n])
		if err != nil {
			session.serv.filterChain.errorCaught(session, err)
			session.Close()
			return
		}
		session.serv.filterChain.msgReceived(session, ioBuffer)
	}
}

func (this *Iosession) Write(message interface{}) error {
	if this.serv.runnable && !this.closed {
		if msg, ok := this.serv.filterChain.msgSend(this, message); ok {
			if m, ok := msg.([]byte); ok {
				_, err := this.conn.Write(m)
				if err != nil {
					this.serv.filterChain.errorCaught(this, err)
					return err
				}
			} else {
				err := errors.New("net.Conn.Write fail,param is not []byte type")
				this.serv.filterChain.errorCaught(this, err)
				return err
			}
		}
		return nil
	} else {
		err := errors.New("Iosession is closed")
		this.serv.filterChain.errorCaught(this, err)
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
