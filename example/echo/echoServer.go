package main

import (
	"fmt"

	"github.com/hechuangqiang/golis"
)

func main() {
	iobuffer := golis.NewBuffer()
	iobuffer.Cap()
	iobuffer.GetReadPos()

	s := golis.NewServer()
	s.FilterChain().AddLast("test", &filter{})
	s.RunOnPort("tcp", ":9090")
}

type filter struct{}

func (*filter) SessionOpened(session *golis.Iosession) bool {
	fmt.Println("session opened")
	return true
}

func (*filter) SessionClosed(session *golis.Iosession) bool {
	fmt.Println("session closed")
	return true
}

func (*filter) MsgReceived(session *golis.Iosession, message interface{}) bool {
	if msg, ok := message.(*golis.Buffer); ok {
		fmt.Println(msg.GetReadPos(), msg.GetWritePos())
		bs, _ := msg.ReadBytes(msg.GetWritePos() - msg.GetReadPos())
		fmt.Println("received msg : ", string(bs))

	} else {
		fmt.Println("not ok")
	}
	return true
}

func (*filter) MsgSend(session *golis.Iosession, message interface{}) (interface{}, bool) {
	return nil, true
}

func (*filter) ErrorCaught(session *golis.Iosession, err error) bool {
	return true
}
