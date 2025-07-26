package main

import (
	"net"
	"strings"
)

type User struct {
	Name   string
	Addr   string
	C      chan string
	conn   net.Conn
	server *Server
}

// Construct new User
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}

	// Go to Listen Goroutine
	go user.ListenMessage()

	return user
}

func (this *User) Online() {
	//用户上线，将当前用户加入到onlineMap中
	this.server.MapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.MapLock.Unlock()

	//广播用户上线消息
	this.server.BroadCast(this, "已上线")
}

// 用户下线
func (this *User) Offline() {
	this.server.MapLock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.MapLock.Unlock()

	//广播用户下线消息
	this.server.BroadCast(this, "已下线")
}

func (this *User) Sendmsg(msg string) {
	this.conn.Write([]byte(msg))
}

// 用户处理消息
func (this *User) DoMessage(msg string) {
	// 查询当前用户都有哪些
	if msg == "who" {
		this.server.MapLock.Lock()
		for _, user := range this.server.OnlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + ": 在线...\n"
			this.Sendmsg(onlineMsg)
		}
		this.server.MapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename-" {
		newName := strings.Split(msg, "-")[1]
		this.server.MapLock.Lock()
		_, contain := this.server.OnlineMap[newName]

		// 检查newname是否存在
		if contain {
			this.server.MapLock.Unlock()
			this.Sendmsg("当前用户名被使用\n")
		} else {
			delete(this.server.OnlineMap, this.Name)
			this.server.OnlineMap[newName] = this
			this.Name = newName

			this.server.MapLock.Unlock()
			this.Sendmsg("你已经更新用户名" + this.Name + "\n")
		}
	} else if len(msg) > 4 && msg[:3] == "to-" {
		// 1. 获取用户名
		username := strings.Split(msg, "-")[1]
		if username == "" {
			this.Sendmsg("消息格式不正确，请使用\"to-张伟-你好\"格式\n")
			return
		}

		// 2. 根据用户名，得到User对象
		this.server.MapLock.Lock()
		remoteuser, ok := this.server.OnlineMap[username]
		this.server.MapLock.Unlock()
		if !ok {
			this.Sendmsg("该用户名不存在\n")
			return
		}

		// 3. 获取消息内容，向User对象发送消息
		content := strings.Split(msg, "-")[2]
		if content == "" {
			this.Sendmsg("无消息内容！")
			return
		}
		remoteuser.Sendmsg(this.Name + "对您说：" + content + "\n")
	} else {
		this.server.BroadCast(this, msg)
	}
}

// Listen Goroutine
func (this *User) ListenMessage() {
	for {
		msg := <-this.C

		this.conn.Write([]byte(msg + "\n"))
	}
}
