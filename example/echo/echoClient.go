package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/hechuangqiang/golis"
)

func main() {
	c := golis.NewClient()
	c.FilterChain().AddLast("default", &golis.IoFilterAdapter{}).AddLast("clientFilter", &filter{})
	c.Dial("tcp", "127.0.0.1:9090")
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

func (*filter) MsgReceived(session *golis.Iosession, message interface{}) (interface{}, bool) {
	fmt.Println("MsgReceived")
	if bs, ok := message.([]byte); ok {
		msg := string(bs)
		fmt.Println(msg)
		if strings.Contains(msg, "close") {
			session.Close()
		}
	}
	return message, true
}

func (*filter) MsgSent(session *golis.Iosession, message interface{}) (interface{}, bool) {
	fmt.Println("client msg sent")
	return message, true
}
