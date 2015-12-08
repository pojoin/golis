package golis

import (
	"errors"
	"log"
	"net"
	"sync"
	"time"
)

//系统变量定义
var (
	GolisHandler IoHandler            //事件处理
	GolisPackage Packager             //拆包封包处理接口
	w            WaitGroupWrapper     //等待退出
	runnable     bool                 //服务运行状态
	rwTimeout    time.Duration    = 0 //超时单位秒
)

//定义waitGroup
type WaitGroupWrapper struct {
	sync.WaitGroup
}

//定义session
type Iosession struct {
	SesionId   int32            //session唯一表示
	Connection net.Conn         //连接
	IsAuth     bool             //是否认证成功
	closed     bool             //是否已经关闭
	T          time.Time        //最后一次调用时间
	writeChan  chan interface{} //写入通道
}

//session写入数据
func (this *Iosession) Write(message interface{}) error {
	if runnable && !this.closed {
		this.writeChan <- message
		return nil
	} else {
		GolisHandler.MessageSendFail(this, message)
		return errors.New("session is closed")
	}
}

//关闭连接
func (this *Iosession) Close() {
	if !this.closed {
		GolisHandler.SessionClosed(this)
		this.closed = true
	}
	this.Connection.Close()
}

//启动channal发送数据
func (session *Iosession) writeDataToConn() {

	for runnable && !session.closed {
		select {
		case data := <-session.writeChan:
			{
				GolisHandler.MessageSent(session, data)
				_, err := session.Connection.Write(GolisPackage.Packet(data))
				if err != nil {
					GolisHandler.MessageSendFail(session, data)
				}
			}
		}
	}
}

//创建新session
func newIoSession(conn net.Conn) *Iosession {
	session := &Iosession{}
	session.Connection = conn
	session.T = time.Now()
	session.writeChan = make(chan interface{}, 16)
	go session.writeDataToConn()
	return session
}

//设置读写超时
func SetTimeout(timeoutSec time.Duration) {
	rwTimeout = timeoutSec
}

//拆包封包接口定义
type Packager interface {
	//读取连接数据
	//packageChan 准备好包后交给该chan
	ReadConnData(buffer *Buffer, packageChan chan<- *[]byte)
	//拆包函数
	Unpacket(data *[]byte) interface{}
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
	//消息发送失败触发
	MessageSendFail(session *Iosession, message interface{})
}

//服务器端运行golis
//netPro：运行协议参数，tcp/udp
//laddr ：程序监听ip和端口，如127.0.0.1:8080
func Run(netPro, laddr string) {
	runnable = true
	log.Println("golis is listen port:", laddr)

	netLis, err := net.Listen(netPro, laddr)
	if err != nil {
		log.Fatalln(err)
	}
	defer netLis.Close()
	log.Println("waiting clients...")
	for runnable {
		conn, err := netLis.Accept()
		if err != nil {
			continue
		}
		go connectHandle(conn)
	}

	w.WaitGroup.Wait()
	log.Println("golis is safe exit")
}

//停止服务
func Stop() {
	runnable = false
}

//客户端程序连接服务器
func Dial(netPro, laddr string) {
	runnable = true
	conn, err := net.Dial(netPro, laddr)
	if err != nil {
		log.Fatalln(err)
	}
	go connectHandle(conn)
}

//处理新连接
func connectHandle(conn net.Conn) {
	w.Add(1)
	//声明一个临时缓冲区，用来存储被截断的数据
	ioBuffer := NewBuffer()

	buffer := make([]byte, 512)

	//声明一个管道用于接收解包的数据
	readerChannel := make(chan *[]byte, 16)
	//创建session
	session := newIoSession(conn)
	//触发sessionCreated事件
	GolisHandler.SessionOpened(session)

	exitChan := make(chan bool)
	go waitData(session, readerChannel, exitChan)
	defer func() {
		conn.Close()
		ioBuffer = nil
		buffer = nil
		if err := recover(); err != nil {
			log.Println(err)
		}
		w.Done()
	}()

	for runnable && !session.closed {

		n, err := conn.Read(buffer)
		//设置超时
		//resetTimeout(conn)
		ioBuffer.PutBytes(buffer[:n])
		if err == nil {
			GolisPackage.ReadConnData(ioBuffer, readerChannel)
		} else {
			//session已经关闭
			session.Close()
			exitChan <- true
			return
		}
		if !runnable || session.closed {
			//session关闭
			session.Close()
			exitChan <- true
			return
		}

	}

}

//设置超时时间
func resetTimeout(conn net.Conn) {
	if rwTimeout == 0 {
		conn.SetDeadline(time.Time{})
	} else {
		conn.SetDeadline(time.Now().Add(time.Duration(rwTimeout) * time.Second))
	}
}

//等待数据包
func waitData(session *Iosession, readerChannel chan *[]byte, exitChan chan bool) {
	defer func() {
		close(exitChan)
		close(readerChannel)
		if err := recover(); err != nil {
			log.Println(err)
		}
	}()
	for runnable && !session.closed {
		select {
		case data := <-readerChannel:
			readFromData(session, data)
		case <-exitChan:
			return
		}
	}
}

//从准备好的数据读取并拆包
func readFromData(session *Iosession, data *[]byte) {
	message := GolisPackage.Unpacket(data) //拆包
	//收到消息到达时触发事件
	GolisHandler.MessageReceived(session, message)
}
