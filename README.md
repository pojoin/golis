# golis
golis 是一个简单的构建网络消息传输的脚手架，类型java中的mina，简单方便。

根据需要自定义消息报文格式，可以参考例子。

golis 通过IoFilterChain 处理定义好的IoFilter,类似mina。

##Quick Start
######Download and install

    go get github.com/hechuangqiang/golis

####example echoServer
######Create file `echoServer.go`
```go
package main

import (
	"fmt"

	"github.com/hechuangqiang/golis"
)

func main() {
	s := golis.NewServer()
	s.FilterChain().AddLast("codec", &codecFilter{}).AddLast("test", &filter{})
	s.RunOnPort("tcp", ":9090")
}

type codecFilter struct {
	golis.IoFilterAdapter
}

func (*codecFilter) Decode(message interface{}) (interface{}, bool) {
	if buffer, ok := message.(*golis.Buffer); ok {
		bs, _ := buffer.ReadBytes(buffer.GetWritePos() - buffer.GetReadPos())
		buffer.ResetRead()
		buffer.ResetWrite()
		return bs, true
	}
	return message, false
}

func (*codecFilter) Encode(message interface{}) (interface{}, bool) {
	return message, true
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

func (*filter) MsgReceived(session *golis.Iosession, message interface{}) bool {
	if bs, ok := message.([]byte); ok {
		fmt.Println("received msg :", string(bs))
		replayMsg := fmt.Sprintf("echoServer received msg : %v", string(bs))
		session.Write([]byte(replayMsg))
	}
	return true
}

func (*filter) MsgSend(session *golis.Iosession, message interface{}) bool {
	return true
}

func (*filter) ErrorCaught(session *golis.Iosession, err error) bool {
	return true
}
```
######Build and run
```bash
    go run echoServer.go
```
######test
```bash
    telnet 127.0.0.1 9090
````
More [examples](https://github.com/hechuangqiang/golis/tree/master/example)
