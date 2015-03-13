package golis

import (
	"fmt"
	"io"
	"net"
	"os"
)

//系统变量定义

var (
	ioHandler IoHandler    //事件处理
	unpacket  UnpacketType //协议解包函数
	packet    PacketType   //协议封包函数
)

//定义解包函数类型
type UnpacketType func([]byte, chan interface{}) []byte

//定义封包函数类型
type PacketType func(interface{}) []byte

//定义session
type Iosession struct {
	conn net.Conn
}

//session写入数据
func (this *Iosession) Write(message *interface{}) {
	//触发消息发送事件
	ioHandler.MessageSent(this, message)
	this.conn.Write(packet(message))
}

//事件触发接口定义
type IoHandler interface {
	//session打开
	SessionOpened(session *Iosession)
	//session关闭
	SessionClosed(session *Iosession)
	//收到消息时触发
	MessageReceived(session *Iosession, message *interface{})
	//消息发送时触发
	MessageSent(session *Iosession, message *interface{})
}

//运行golis
//netPro：运行协议参数，tcp/udp
//laddr ：程序监听ip和端口，如127.0.0.1:8080
func Run(netPro, laddr string) {
	Log("初始化系统完成")
	netLis, err := net.Listen(netPro, laddr)
	CheckError(err)
	defer netLis.Close()
	Log("等待客户端连接...")
	for {
		conn, err := netLis.Accept()
		if err != nil {
			continue
		}
		go connectHandle(conn)
	}
}

//处理新连接
func connectHandle(conn net.Conn) {
	//声明一个临时缓冲区，用来存储被截断的数据
	tmpBuffer := make([]byte, 0)
	buffer := make([]byte, 1024)
	readChannel := make(chan interface{}, 16)
	//创建session
	session := Iosession{conn}
	//触发sessionCreated事件
	ioHandler.SessionOpened(&session)

	//通过readDataFromChannel监听客户发过来的数据
	go readDataFromChannel(&session, readChannel)
	flag := true
	for flag {
		n, err := conn.Read(buffer)
		switch err {
		case nil:
			tmpBuffer = unpacket(append(tmpBuffer, buffer[:n]...), readChannel)
		case io.EOF:
			Log("client is disconnected")
			flag = false
			break
		default:
			Log("none")
		}
	}

}

func readDataFromChannel(session *Iosession, readChannel chan interface{}) {
	for {
		select {
		case data := <-readChannel:
			//收到消息时到达
			ioHandler.MessageReceived(session, &data)
		}
	}
}

//简单日志输出
func Log(v ...interface{}) {
	fmt.Println(v...)
}

//检查错误并退出程序
func CheckError(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "Fatal error:%s", err.Error())
		os.Exit(1)
	}
}
