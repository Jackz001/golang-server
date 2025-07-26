package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	flag       int
}

// 创建客户端实例并连接服务器
func NewClient(serverIp string, serverPort int) *Client {
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		flag:       999,
	}

	// 建立TCP连接
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("net.Dial error:", err)
		return nil
	}

	client.conn = conn
	return client
}

// 处理服务器响应的go程（持续输出到标准输出）
func (this *Client) DealResponese() {
	io.Copy(os.Stdout, this.conn)
}

// 显示用户菜单界面
func (this *Client) menu() bool {
	var flag int

	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户名")
	fmt.Println("0.退出")

	fmt.Scanln(&flag)

	if flag >= 0 && flag <= 3 {
		this.flag = flag
		return true
	} else {
		fmt.Println(">>>>请输入合法范围内的数字<<<<")
		return false
	}
}

// 查询当前在线用户
func (this *Client) SelectUsers() {
	sendMsg := "who\n"
	_, err := this.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn Write error:", err)
		return
	}
}

// 进入私聊模式
func (this *Client) PrivateChat() {
	var remoteName string
	var remoteMsg string

	this.SelectUsers()
	fmt.Println("请输入聊天对象[用户名], exit退出:")
	fmt.Scanln(&remoteName)

	for remoteName != "exit" {
		fmt.Println(">>>>请输入消息内容, exit退出:")
		fmt.Scanln(&remoteMsg)

		for remoteMsg != "exit" {
			if len(remoteMsg) != 0 && len(remoteName) != 0 {
				sendMsg := "to-" + remoteName + "-" + remoteMsg + "\n"
				_, err := this.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("conn Write err,", err)
					break
				}
			}

			remoteMsg = ""
			fmt.Println(">>>>请输入消息内容, exit退出:")
			fmt.Scanln(&remoteMsg)
		}

		this.SelectUsers()
		fmt.Println("请输入聊天对象[用户名], exit退出:")
		fmt.Scanln(&remoteName)
	}
}

// 进入公聊模式
func (this *Client) PublicChat() {
	var chatMsg string

	fmt.Println(">>>>请输入聊天内容, exit退出")
	fmt.Scanln(&chatMsg)

	for chatMsg != "exit" {
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_, err := this.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn Write err,", err)
				break
			}
		}

		chatMsg = ""
		fmt.Println(">>>>请输入聊天内容, exit退出")
		fmt.Scanln(&chatMsg)
	}
}

// 更新当前用户名
func (this *Client) UpdateName() bool {
	fmt.Println(">>>>请输入用户名:")
	fmt.Scanln(&this.Name)

	sendMsg := "rename-" + this.Name + "\n"
	_, err := this.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write err:", err)
		return false
	}

	return true
}

// 运行客户端主循环
func (this *Client) Run() {

	for this.flag != 0 {
		for this.menu() != true {
		}

		// 根据选择执行对应操作
		switch this.flag {
		case 1:
			this.PublicChat()
			break
		case 2:
			this.PrivateChat()
			break
		case 3:
			this.UpdateName()
			break
		}
	}
}

var serverIp string
var serverPort int

func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置服务器ip地址(默认是127.0.0.1)")
	flag.IntVar(&serverPort, "port", 8888, "设置服务器端口(默认是8888)")
}

func main() {
	flag.Parse()

	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println(">>>>>>>连接服务器失败")
		return
	}

	go client.DealResponese()
	fmt.Println(">>>>>>>连接服务器成功")
	client.Run()
}
