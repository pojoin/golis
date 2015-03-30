package golis

import (
	"log"
	"net"
)

//系统变量定义
var (
	GolisHandler IoHandler //事件处理
	GolisPackage Packager  //拆包封包处理接口
)

//定义session
type Iosession struct {
	SesionId   int      //session唯一表示
	Connection net.Conn //连接
}

//session写入数据
func (this *Iosession) Write(message interface{}) {
	//触发消息发送事件
	GolisHandler.MessageSent(this, message)
	this.Connection.Write(GolisPackage.Packet(message))
}

//关闭连接
func (this *Iosession) Close() {
	GolisHandler.SessionClosed(this)
	this.Connection.Close()
}

//拆包封包接口定义
type Packager interface {
	//读取连接数据
	//connData 从连接当中读取数据
	//packageChan 准备好包后交给该chan
	//remain 读取connData的剩余数据返回
	ReadConnData(connData []byte, packageChan chan<- []byte) (remain []byte)
	//拆包函数
	Unpacket(data []byte) interface{}
	//封包函数
	Packet(msg interface{}) []byte
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

//服务器端运行golis
//netPro：运行协议参数，tcp/udp
//laddr ：程序监听ip和端口，如127.0.0.1:8080
func Run(netPro, laddr string) {
	log.Println("golis is listen port:", laddr)
	netLis, err := net.Listen(netPro, laddr)
	if err != nil {
		log.Fatalln(err)
	}
	defer netLis.Close()
	log.Println("waiting clients...")
	for {
		conn, err := netLis.Accept()
		if err != nil {
			continue
		}
		go connectHandle(conn)
	}
}

//客户端程序连接服务器
func Dial(netPro, laddr string) {
	conn, err := net.Dial(netPro, laddr)
	if err != nil {
		log.Fatalln(err)
	}
	go connectHandle(conn)
}

//处理新连接
func connectHandle(conn net.Conn) {
	defer conn.Close()
	//声明一个临时缓冲区，用来存储被截断的数据
	tmpBuffer := make([]byte, 0)
	buffer := make([]byte, 512)

	//声明一个管道用于接收解包的数据
	readerChannel := make(chan []byte, 16)
	//创建session
	session := Iosession{Connection: conn}
	//触发sessionCreated事件
	GolisHandler.SessionOpened(&session)

	exitChan := make(chan bool, 0)
	go waitData(&session, readerChannel, exitChan)

	flag := true
	for flag {

		n, err := conn.Read(buffer)
		//Log(time.Now().Unix(), "tmpBuffer.len:", len(tmpBuffer), "tmpBuffer.cap:", cap(tmpBuffer), "n:", n)

		if err == nil {
			tmpBuffer = GolisPackage.ReadConnData(append(tmpBuffer, buffer[:n]...), readerChannel)
		} else {
			//			Log("client is disconnected")
			//session关闭
			GolisHandler.SessionClosed(&session)
			flag = false
			exitChan <- true
			break
		}
	}

}

//等待数据包
func waitData(session *Iosession, readerChannel chan []byte, exitChan chan bool) {
	for {
		select {
		case data := <-readerChannel:
			readFromData(session, data)
		case <-exitChan:
			break
		}
	}
}

//从准备好的数据读取并拆包
func readFromData(session *Iosession, data []byte) {
	message := GolisPackage.Unpacket(data) //拆包
	//收到消息到达时触发事件
	GolisHandler.MessageReceived(session, message)
}
