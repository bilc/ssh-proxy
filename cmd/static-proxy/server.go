package main

import (
	"log"
	"time"

	"golang.org/x/crypto/ssh"
)

func NewSshSession(host, port string, user, pass string) (*ssh.Session, error) {
	// 建立SSH客户端连接
	client, err := ssh.Dial("tcp", host+":"+port, &ssh.ClientConfig{
		User:            user,
		Auth:            []ssh.AuthMethod{ssh.Password(pass)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         time.Second * 10,
	})
	if err != nil {
		log.Fatalf("SSH dial error: %s", err.Error())
	}

	// 建立新会话
	session, err := client.NewSession()
	return session, err
}

type ChannelDemocrator struct {
	channel ssh.Channel
}

func NewChannelDemocrator(channel ssh.Channel) *ChannelDemocrator {
	return &ChannelDemocrator{channel: channel}
}

func (c *ChannelDemocrator) Read(p []byte) (int, error) {
	log.Printf("read: %s", string(p))
	return c.channel.Read(p)
}
func (c *ChannelDemocrator) Write(p []byte) (int, error) {
	log.Printf("write: %s", string(p))
	return c.channel.Write(p)
}
func (c *ChannelDemocrator) Close() error {
	return c.channel.Close()
}
