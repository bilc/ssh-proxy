package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/crypto/ssh"
)

func privateDiffie() (b ssh.Signer, err error) {
	private, err := ioutil.ReadFile("/home/blc/.ssh/id_rsa")
	if err != nil {
		return
	}
	b, err = ssh.ParsePrivateKey(private)
	return
}

// 开启goroutine, 处理连接的Channel
func handleChannels(chans <-chan ssh.NewChannel) {
	for newChannel := range chans {
		go handleChannel(newChannel)
	}
}

// parseDims extracts two uint32s from the provided buffer.
func parseDims(b []byte) (uint32, uint32) {
	w := binary.BigEndian.Uint32(b)
	h := binary.BigEndian.Uint32(b[4:])
	return w, h
}

// Winsize stores the Height and Width of a terminal.
type Winsize struct {
	Height uint16
	Width  uint16
	x      uint16 // unused
	y      uint16 // unused
}

// SetWinsize sets the size of the given pty.
func SetWinsize(fd uintptr, w, h uint32) {
	log.Printf("window resize %dx%d", w, h)
	ws := &Winsize{Width: uint16(w), Height: uint16(h)}
	syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(syscall.TIOCSWINSZ), uintptr(unsafe.Pointer(ws)))
}

func handleChannel(ch ssh.NewChannel) {
	// 仅处理session类型的channel（交互式tty服务端）
	if t := ch.ChannelType(); t != "session" {
		ch.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", t))
		return
	}
	// 返回两个队列，connection用于数据交换，requests用户控制指令交互
	connection, requests, err := ch.Accept()
	if err != nil {
		log.Printf("Could not accept channel (%s)", err.Error())
		return
	}
	connection.Write([]byte("welcom"))
	buf := make([]byte, 100)
	n, err := connection.Read(buf)
	log.Println("accept ", string(buf), n, err)

	servSession, _ := NewSshSession()
	// result, err := servSession.Output("ls -al")
	// if err != nil {
	// 	log.Printf("Failed to run command, Err:%s", err.Error())
	// }
	// fmt.Println(string(result))

	// bufio.newBuffer()
	// connection.Read()

	servSession.Stdin, _ = os.CreateTemp("/tmp", "sshtmp")
	servSession.Stdout, _ = os.CreateTemp("/tmp", "sshtmp")
	servSession.Stderr, _ = os.CreateTemp("/tmp", "sshtmp")

	// 为session启动一个bash
	// 关闭连接和session
	close := func() {
		connection.Close()
		log.Println("Session closed")
	}

	// 管道session到bash和visa-versa
	// 使用 sync.Once 确保close只调用一次
	var once sync.Once
	go func() {
		for {
			n, e := io.Copy(servSession.Stdout, connection)
			log.Println("Session closed 1 ", n, e)
			time.Sleep(time.Second)
		}

		once.Do(close)
	}()
	go func() {
		for {
			n, e := io.Copy(connection, servSession.Stdin)
			log.Println("Session closed 2", n, e)
			time.Sleep(time.Second)
		}

		once.Do(close)
	}()

	// session out-of-band请求有"shell"、"pty-req"、"env"等几种
	go func() {
		for req := range requests {
			switch req.Type {
			case "shell":
				if len(req.Payload) == 0 {
					req.Reply(true, nil)
				}
			case "pty-req":
				//	termLen := req.Payload[3]
				// w, h := parseDims(req.Payload[termLen+4:])
				// SetWinsize(tty.Fd(), w, h)
				req.Reply(true, nil)
			case "window-change":
				// w, h := parseDims(req.Payload)
				// // SetWinsize(tty.Fd(), w, h)
			}
		}
	}()
	time.Sleep(time.Hour)
}

func main() {
	config := &ssh.ServerConfig{
		// 密码验证回调函数
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			if c.User() == "demo" && string(pass) == "123456" {
				return nil, nil
			}
			return nil, fmt.Errorf("password rejected for %q", c.User())
		},
		// NoClientAuth: true, // 客户端不验证，即任何客户端都可以连接
		// ServerVersion: "SSH-2.0-OWN-SERVER", // "SSH-2.0-"，SSH版本
	}
	// 秘钥用于SSH交互双方进行 Diffie-hellman 秘钥交换验证
	if b, err := privateDiffie(); err != nil {
		log.Printf("private diffie host key error: %s", err.Error())
	} else {
		config.AddHostKey(b)
	}

	// 监听地址和端口
	listener, err := net.Listen("tcp", "0.0.0.0:19022")
	if err != nil {
		log.Fatalf("Failed to listen on 1022 (%s)", err.Error())
	}
	log.Println("listen to 0.0.0.0:19022")

	// 接受所有连接
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept incoming connection (%s)", err)
			continue
		}
		// 使用前，必须传入连接进行握手 net.Conn
		sshConn, chans, reqs, err := ssh.NewServerConn(conn, config)
		if err != nil {
			log.Printf("Failed to handshake (%s)", err)
			continue
		}
		log.Printf("New SSH connection from %s (%s)", sshConn.RemoteAddr(), sshConn.ClientVersion())
		go ssh.DiscardRequests(reqs)
		// 接收所有channels
		go handleChannels(chans)
	}
}
