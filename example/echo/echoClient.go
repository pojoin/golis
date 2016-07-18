package main

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/pojoin/golis"
)

func main() {
	c := golis.NewClient()
	c.FilterChain().AddLast("clientFilter", &filter{})
	c.SetCodecer(&echoProtocalCodec{})
	c.Dial("tcp", "127.0.0.1:9090")
}

type echoProtocalCodec struct {
	golis.ProtocalCodec
}

func (*echoProtocalCodec) Decode(buffer *golis.Buffer, dataCh chan<- interface{}) error {
	bs, _ := buffer.ReadBytes(buffer.GetWritePos() - buffer.GetReadPos())
	buffer.ResetWrite()
	buffer.ResetRead()
	dataCh <- bs
	return nil
}

func (*echoProtocalCodec) Encode(message interface{}) ([]byte, error) {
	if bs, ok := message.([]byte); ok {
		return bs, nil
	}
	return nil, errors.New("failed")
}

type filter struct {
	golis.IoFilterAdapter
}

func (*filter) SessionClosed(session *golis.Iosession) bool {
	fmt.Println("client session closed")
	return true
}

func (*filter) SessionOpened(session *golis.Iosession) bool {
	fmt.Println("client session opened")
	for i := 0; i < 10; i++ {
		msg := fmt.Sprintf("client send num %d", i)
		session.Write([]byte(msg))
		time.Sleep(time.Second * 1)
	}
	session.Write([]byte("close"))
	return true
}

func (*filter) MsgReceived(session *golis.Iosession, message interface{}) bool {
	if bs, ok := message.([]byte); ok {
		msg := string(bs)
		fmt.Println(msg)
		if strings.Contains(msg, "close") {
			session.Close()
		}
	}
	return true
}

func (*filter) MsgSent(session *golis.Iosession, message interface{}) bool {
	fmt.Println("client msg sent")
	return true
}
