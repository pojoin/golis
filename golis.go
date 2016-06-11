package golis

import (
	"errors"
	"log"
	"net"
	"sync"
	"time"
)

type ioserv struct {
	sync.WaitGroup
	Handler     IoHandler
	Pkg         IoPackager
	runnable    bool
	rwTimeout   time.Duration
	protocal    string
	ioaddr      string
	filterChain *IoFilterChain
}

func (serv *ioserv) FilterChain() *IoFilterChain {
	return serv.filterChain
}

//create session
func (serv *ioserv) newIoSession(conn net.Conn) *Iosession {
	session := &Iosession{}
	session.Connection = conn
	session.T = time.Now()
	session.writeChan = make(chan interface{}, 16)
	session.serv = serv
	go session.writeDataToConn()
	return session
}

//stop serv
func (serv *ioserv) Stop() {
	serv.runnable = false
}

func (serv *ioserv) resetTimeout(conn net.Conn) {
	if serv.rwTimeout == 0 {
		conn.SetDeadline(time.Time{})
	} else {
		conn.SetDeadline(time.Now().Add(time.Duration(serv.rwTimeout) * time.Second))
	}
}

func (serv *ioserv) readFromData(session *Iosession, data []byte) {
	message := serv.Pkg.Unpacket(data)
	serv.Handler.MessageReceived(session, message)
}

//handle conn
func (serv *ioserv) connectHandle(conn net.Conn) {
	serv.Add(1)
	ioBuffer := NewBuffer()

	buffer := make([]byte, 512)

	//create session
	session := serv.newIoSession(conn)
	serv.Handler.SessionOpened(session)

	defer func() {
		conn.Close()
		ioBuffer = nil
		buffer = nil
		if err := recover(); err != nil {
			log.Println(err)
		}
		serv.Done()
	}()

	var pkgData []byte
	var ok bool
	for serv.runnable && !session.closed {

		n, err := conn.Read(buffer)
		ioBuffer.PutBytes(buffer[:n])
		if err == nil {
			pkgData, ok = serv.Pkg.ReadConnData(ioBuffer)
			if ok {
				serv.readFromData(session, pkgData)
			}
		} else {
			session.Close()
			return
		}
		if !serv.runnable || session.closed {
			session.Close()
			return
		}

	}

}

//core server
type server struct {
	*ioserv
}

//server run
func (s *server) Run() {
	s.runnable = true
	log.Println("golis is starting...")
	netLis, err := net.Listen(s.protocal, s.ioaddr)
	if err != nil {
		log.Fatalln(err)
	}
	defer netLis.Close()
	log.Println(s.ListenInfo())
	log.Println("waiting clients to connect")
	for s.runnable {
		conn, err := netLis.Accept()
		if err != nil {
			continue
		}
		go s.connectHandle(conn)
	}
}

//server run and listen addr port
func (s *server) RunOnPort(protocal, addr string) {
	s.runnable = true
	log.Println("golis is starting...")
	netLis, err := net.Listen(protocal, addr)
	if err != nil {
		log.Fatalln(err)
	}
	defer netLis.Close()
	log.Println(s.ListenInfo())
	log.Println("waiting clients to connect")
	for s.runnable {
		conn, err := netLis.Accept()
		if err != nil {
			continue
		}
		go s.connectHandle(conn)
	}
}

//set port and protocal ,the protocal value can be "tcp" or "udp"
func (s *server) SetPort(protocal, addr string) {
	s.protocal = protocal
	s.ioaddr = addr
}

//get port
func (s *server) Port() string {
	return s.ioaddr
}

//get listen info
func (s *server) ListenInfo() string {
	return "the server listened protocal is " + s.protocal + " and istened addr is " + s.ioaddr
}

type client struct {
	*ioserv
}

// dial to server
func (c *client) Dial(netPro, laddr string) {
	c.runnable = true
	conn, err := net.Dial(netPro, laddr)
	if err != nil {
		log.Fatalln(err)
	}
	go c.connectHandle(conn)
}

//iosession
type Iosession struct {
	serv       *ioserv
	SesionId   int32
	Connection net.Conn
	IsAuth     bool
	closed     bool
	T          time.Time
	writeChan  chan interface{}
}

func (this *Iosession) Write(message interface{}) error {
	if this.serv.runnable && !this.closed {
		this.writeChan <- message
		return nil
	} else {
		this.serv.Handler.MessageSendFail(this, message)
		return errors.New("session is closed")
	}
}

//close iosession
func (this *Iosession) Close() {
	if !this.closed {
		this.serv.Handler.SessionClosed(this)
		this.closed = true
	}
	this.Connection.Close()
}

func (session *Iosession) writeDataToConn() {

	for session.serv.runnable && !session.closed {
		select {
		case data := <-session.writeChan:
			{
				session.serv.Handler.MessageSent(session, data)
				_, err := session.Connection.Write(session.serv.Pkg.Packet(data))
				if err != nil {
					session.serv.Handler.MessageSendFail(session, data)
				}
			}
		}
	}
}

type IoPackager interface {
	ReadConnData(buffer *Buffer) (pkgData []byte, ok bool)
	Unpacket(data []byte) interface{}
	Packet(msg interface{}) []byte
}

type IoHandler interface {
	//session opened
	SessionOpened(session *Iosession)
	//session closed
	SessionClosed(session *Iosession)
	MessageReceived(session *Iosession, message interface{})
	MessageSent(session *Iosession, message interface{})
	MessageSendFail(session *Iosession, message interface{})
}
