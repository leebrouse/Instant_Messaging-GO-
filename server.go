package main

import (
	"fmt"
	"io"
	"net"
	"sync"
)

// Server 结构体，表示服务器
type Server struct {
	IP   string // 服务器的IP地址
	Port int    // 服务器的端口号

	OnlineMap map[string]*User // 在线用户列表，键为用户名，值为用户指针
	mapLock   sync.RWMutex     // 读写锁，用于保护OnlineMap的并发安全

	// 广播消息的通道
	Message chan string
}

// NewServer 创建一个服务器对象
func NewServer(ip string, port int) *Server {
	server := &Server{
		IP:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User), // 初始化用户列表
		Message:   make(chan string),      // 初始化消息通道
	}

	return server
}

// ListenMessager 监听并转发广播消息
func (this *Server) ListenMessager() {
	for {
		// 从Message通道中接收消息
		msg := <-this.Message

		// 给所有在线用户发送该消息
		this.mapLock.Lock() // 加锁，确保在线用户列表的安全访问
		for _, cli := range this.OnlineMap {
			cli.Channel <- msg // 向每个用户的消息通道发送广播消息
		}
		this.mapLock.Unlock() // 解锁
	}
}

// BroadCast 广播消息给所有用户
func (this *Server) BroadCast(user *User, msg string) {
	// 构建发送的消息内容，格式为：[用户地址]用户名:消息内容
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg

	// 将构建好的消息发送到Message通道中
	this.Message <- sendMsg
}

// Handler 处理用户连接
func (this *Server) Handler(conn net.Conn) {
	// 新用户连接后，创建一个新的User对象
	user := NewUser(conn)

	// 用户上线，将用户加入在线用户列表
	this.mapLock.Lock()
	this.OnlineMap[user.Name] = user
	this.mapLock.Unlock()

	// 广播用户上线的消息
	this.BroadCast(user, "已上线")

	//接受客户端发送的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				this.BroadCast(user, "下线")
				return
			}

			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err:", err)
				return
			}

			//提取用户的消息(去除‘\n’)
			msg := string(buf[:n-1])

			//将得到的消息进行广播
			this.BroadCast(user, msg)
		}
	}()

	// 阻塞当前协程，保持用户在线状态
	select {}
}

// Start 启动服务器
func (this *Server) Start() {
	// 开启TCP监听
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.IP, this.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}
	defer listener.Close() // 当程序退出时关闭监听器

	// 启动监听广播消息的协程
	go this.ListenMessager()

	// 不断接收客户端的连接请求
	for {
		// 接收客户端连接
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept err:", err)
			continue
		}

		// 开启一个新的协程处理连接
		go this.Handler(conn)
	}
}
