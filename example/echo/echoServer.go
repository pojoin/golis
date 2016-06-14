package main

import (
	"fmt"

	"github.com/hechuangqiang/golis"
)

func main() {
	s := golis.NewServer()
	s.FilterChain().AddLast("test", &filter{})
	s.RunOnPort("tcp", ":9090")
}

type filter struct{}

func (*filter) SessionOpened(session *golis.Iosession) bool {
	fmt.Println("session opened,the client is ", session.GetConn().RemoteAddr().String())
	return true
}

func (*filter) SessionClosed(session *golis.Iosession) bool {
	fmt.Println("session closed")
	return true
}

func (*filter) MsgReceived(session *golis.Iosession, message interface{}) (interface{}, bool) {
	if msg, ok := message.(*golis.Buffer); ok {
		bs, _ := msg.ReadBytes(msg.GetWritePos() - msg.GetReadPos())
		fmt.Println("received msg :", string(bs))
		replayMsg := fmt.Sprintf("echoServer received msg : %v", string(bs))
		session.Write([]byte(replayMsg))
		msg.ResetRead()
		msg.ResetWrite()
	} else {
		fmt.Println("not ok")
	}
	return message, true
}

func (*filter) MsgSend(session *golis.Iosession, message interface{}) (interface{}, bool) {
	return message, true
}

func (*filter) ErrorCaught(session *golis.Iosession, err error) bool {
	return true
}
