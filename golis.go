package golis

import (
	"log"
	"net"
	"sync"
	"time"
)

type ioserv struct {
	wg          sync.WaitGroup
	runnable    bool
	filterChain *IoFilterChain
}

func (serv *ioserv) FilterChain() *IoFilterChain {
	return serv.filterChain
}

//create session
func (serv *ioserv) newIoSession(conn net.Conn) *Iosession {
	session := &Iosession{}
	session.conn = conn
	session.T = time.Now()
	session.serv = serv
	session.serv.filterChain.sessionOpened(session)
	return session
}

//stop serv
func (serv *ioserv) Stop() {
	serv.runnable = false
}

//core server
type server struct {
	ioserv
	protocal string
	ioaddr   string
}

func NewServer() *server {
	s := &server{}
	s.protocal = "tcp"
	s.ioaddr = "10010"
	s.filterChain = &IoFilterChain{}
	return s
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
		go s.newIoSession(conn).readData()
	}
}

//server run and listen addr port
func (s *server) RunOnPort(protocal, addr string) {
	s.protocal = protocal
	s.ioaddr = addr
	s.Run()
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
	return "the server listened protocal is " + s.protocal + " and listened addr is " + s.ioaddr
}

type client struct {
	ioserv
}

func NewClient() *client {
	c := &client{}
	c.filterChain = &IoFilterChain{}
	return c
}

// dial to server
func (c *client) Dial(netPro, laddr string) {
	c.runnable = true
	conn, err := net.Dial(netPro, laddr)
	if err != nil {
		log.Fatalln(err)
	}
	go c.newIoSession(conn).readData()
}
