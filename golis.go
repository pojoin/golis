package golis

import (
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type ioserv struct {
	generator_id uint64
	wg           sync.WaitGroup
	runnable     bool
	filterChain  *IoFilterChain
	codecer      Codecer
}

func (serv *ioserv) FilterChain() *IoFilterChain {
	return serv.filterChain
}

func (serv *ioserv) SetCodecer(codecer Codecer) {
	serv.codecer = codecer
}

//create session
func (serv *ioserv) newIoSession(conn net.Conn) *Iosession {
	session := &Iosession{}
	session.conn = conn
	session.serv = serv
	session.closed = false
	session.extraData = make(map[string]interface{})
	session.dataCh = make(chan interface{}, 16)
	session.id = atomic.AddUint64(&serv.generator_id, 1)
	go session.dealDataCh()
	go session.readData()
	go session.serv.filterChain.sessionOpened(session)
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

//default port is 10086
func NewServer() *server {
	s := &server{}
	s.protocal = "tcp"
	s.ioaddr = "10086"
	s.filterChain = &IoFilterChain{}
	return s
}

//server run
func (s *server) Run() {
	s.runnable = true
	fmt.Println("golis is starting...")
	netLis, err := net.Listen(s.protocal, s.ioaddr)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer netLis.Close()
	fmt.Println(s.ListenInfo())
	fmt.Println("waiting clients to connect...")
	for s.runnable {
		conn, err := netLis.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		s.newIoSession(conn)
	}
	s.wg.Wait()
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
	return fmt.Sprintf("server listened protocal is %v and listened addr is %v \n",
		s.protocal, s.ioaddr)
}

type Client struct {
	ioserv
}

func NewClient() *Client {
	c := &Client{}
	c.filterChain = &IoFilterChain{}
	return c
}

// dial to server
func (c *Client) Dial(netPro, laddr string) error {
	conn, err := net.Dial(netPro, laddr)
	if err != nil {
		fmt.Println(err)
		return err
	}
	go func() {
		c.runnable = true
		c.newIoSession(conn)
		time.Sleep(20 * time.Millisecond)
		c.wg.Wait()
	}()
	return nil
}
