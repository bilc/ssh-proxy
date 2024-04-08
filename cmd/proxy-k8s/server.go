package main

import (
	"fmt"
	"log"
	"os/exec"
	"time"

	"golang.org/x/crypto/ssh"
)

type ServerManager struct {
}

func NewServerManager() *ServerManager {
	return &ServerManager{}
}

func (s *ServerManager) NewSshSession(user string) (*ssh.Session, error) {
	client := GlobalConfig.GetClient(user)
	if client == nil {
		return nil, fmt.Errorf("user %s not found", user)
	}
	log.Printf("client %v", client)
	cmd := exec.Command("/bin/bash", "-c", client.Install)
	err := cmd.Run()
	if err != nil {
		log.Fatalf("cmd  %s failed with err %s\n", client.Install, err)
	}
	log.Println("cmd run ok")
	session, err := newSshSession(client.SshHost, user, client.Pass)
	log.Println("try NewsSshSession  ", err)
	for i := 0; i < 10; i++ {
		if err == nil {
			break
		}
		log.Println("try NewsSshSession times ", i)
		time.Sleep(time.Second * 3)
		session, err = newSshSession(client.SshHost, user, client.Pass)
	}
	log.Println("try NewsSshSession  ret ", err)
	return session, err
}

func (s *ServerManager) CloseSshSession(user string) error {
	client := GlobalConfig.GetClient(user)
	if client == nil {
		return fmt.Errorf("user %s not found", user)
	}
	cmd := exec.Command("/bin/bash", "-c", client.Uninstall)
	output, err := cmd.Output()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
		return err
	}
	fmt.Printf("%s\n", output)
	return nil
}

func newSshSession(host string, user, pass string) (*ssh.Session, error) {
	// 建立SSH客户端连接
	client, err := ssh.Dial("tcp", host, &ssh.ClientConfig{
		User:            user,
		Auth:            []ssh.AuthMethod{ssh.Password(pass)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         time.Second * 10,
	})
	if err != nil {
		log.Printf("SSH dial error: %s| host %s user %s pass %s", err.Error(), host, user, pass)
		return nil, err
	}

	// 建立新会话
	session, err := client.NewSession()
	return session, err
}
