package main

import "net"

// User 结构体，表示一个在线用户
type User struct {
	Name    string      // 用户名（在这里初始时为用户的网络地址）
	Addr    string      // 用户地址（IP + 端口）
	Channel chan string // 用户的消息通道，用于接收服务器广播的消息
	conn    net.Conn    // 用户的网络连接对象，用于与客户端通信
}

// NewUser 创建一个新的用户对象
func NewUser(conn net.Conn) *User {
	userAddr := conn.RemoteAddr().String() // 获取用户的远程地址

	user := &User{
		Name:    userAddr,          // 初始用户名设置为用户的远程地址
		Addr:    userAddr,          // 用户地址设置为远程地址
		Channel: make(chan string), // 初始化用户的消息通道
		conn:    conn,              // 保存网络连接对象，用于后续通信
	}

	// 开启一个goroutine，监听用户的消息通道
	go user.ListenMessage()

	return user
}

// ListenMessage 监听用户的消息通道，将收到的消息发送到客户端
func (this *User) ListenMessage() {
	for {
		// 从用户的消息通道中读取消息
		msg := <-this.Channel

		// 将消息发送给用户，使用conn的Write方法写入客户端
		this.conn.Write([]byte(msg + "\n"))
	}
}
