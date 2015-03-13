package golis

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
)

//系统变量定义

var (
	GolisHandler IoHandler //事件处理
	GolisMsg     Message   //消息协议
)

//消息协议
type Message interface {
	Unpacket([]byte) interface{}
	Packet(interface{}) []byte
}

//定义session
type Iosession struct {
	conn net.Conn
}

//session写入数据
func (this *Iosession) Write(message *interface{}) {
	//触发消息发送事件
	GolisHandler.MessageSent(this, message)
	data := GolisMsg.Packet(message)
	totalLen := len(data)
	this.conn.Write(append(IntToBytes(totalLen), data...))
}

//事件触发接口定义
type IoHandler interface {
	//session打开
	SessionOpened(session *Iosession)
	//session关闭
	SessionClosed(session *Iosession)
	//收到消息时触发
	MessageReceived(session *Iosession, message interface{})
	//消息发送时触发
	MessageSent(session *Iosession, message interface{})
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
	//创建session
	session := Iosession{conn}
	//触发sessionCreated事件
	GolisHandler.SessionOpened(&session)

	flag := true
	for flag {
		n, err := conn.Read(buffer)
		switch err {
		case nil:
			tmp, data, err := getReadyData(append(tmpBuffer, buffer[:n]...))
			tmpBuffer = tmp
			if err != nil {
				Log(err.Error())
			} else {
				readFromData(&session, data)
			}
		case io.EOF:
			Log("client is disconnected")
			//session关闭
			GolisHandler.SessionClosed(&session)
			flag = false
			break
		default:
			Log("none")
		}
	}

}

//协议中查看协议头是否满足一个协议报
func getReadyData(buffer []byte) ([]byte, []byte, error) {
	length := len(buffer)
	if length >= 32 {
		totalLen := BytesToInt(buffer[0:32]) //get totalLen
		if totalLen == 0 {
			return make([]byte, 0), nil, errors.New("msg is null")
		} else if totalLen <= length-32 {
			return buffer[totalLen+32:], buffer[32:totalLen], nil
		}

	}
	return buffer, nil, errors.New("msg is not ready")
}

//从准备好的数据读取
func readFromData(session *Iosession, data []byte) {
	message := GolisMsg.Unpacket(data) //拆包
	//收到消息时到达
	GolisHandler.MessageReceived(session, message)
}

//整形转换成字节
func IntToBytes(n int) []byte {
	x := int32(n)

	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, x)
	return bytesBuffer.Bytes()
}

//字节转换成整形
func BytesToInt(b []byte) int {
	bytesBuffer := bytes.NewBuffer(b)

	var x int32
	binary.Read(bytesBuffer, binary.BigEndian, &x)

	return int(x)
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
