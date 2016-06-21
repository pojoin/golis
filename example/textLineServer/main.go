package main

import (
	"fmt"

	"github.com/pojoin/golis"
)

type myFilter struct {
	golis.IoFilterAdapter
}

func (*myFilter) MsgReceived(session *golis.Iosession, message interface{}) bool {
	if msg, ok := message.([]byte); ok {
		fmt.Printf("received msg : %s\n", string(msg))
		session.Write([]byte("copy\n"))
	}
	return false
}

func main() {
	s := golis.NewServer()
	s.FilterChain().AddLast("codec", &TextLineCodecFilter{})
	s.FilterChain().AddLast("myFilter", &myFilter{})
	s.RunOnPort("tcp", ":9090")
}
