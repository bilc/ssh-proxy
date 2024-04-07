package main

import (
	"fmt"
	"log"
	"os/exec"
	"time"

	"golang.org/x/crypto/ssh"
)

type ServerManager struct {
	conf *Config
}

func NewServerManager(conf *Config) *ServerManager {
	return &ServerManager{conf: conf}
}

func (s *ServerManager) NewSshSession(user string) (*ssh.Session, error) {
	client := s.conf.GetClient(user)
	if client == nil {
		return nil, fmt.Errorf("user %s not found", user)
	}
	cmd := exec.Command("/bin/bash", "-c", client.ServerConfig.Install)
	output, err := cmd.Output()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}
	fmt.Printf("%s\n", output)

	session, err := NewSshSession(client.ServerConfig.SshHost, user, client.Pass)
	for i := 0; i < 10; i++ {
		if err == nil {
			break
		}
		time.Sleep(time.Second * 3)
		session, err = NewSshSession(client.ServerConfig.SshHost, user, client.Pass)
	}
	return session, err
}

func (s *ServerManager) CloseSshSession(user string) error {
	client := s.conf.GetClient(user)
	if client == nil {
		return fmt.Errorf("user %s not found", user)
	}
	cmd := exec.Command("/bin/bash", "-c", client.ServerConfig.Uninstall)
	output, err := cmd.Output()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
		return err
	}
	fmt.Printf("%s\n", output)
	return nil
}

func NewSshSession(host string, user, pass string) (*ssh.Session, error) {
	// 建立SSH客户端连接
	client, err := ssh.Dial("tcp", host, &ssh.ClientConfig{
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
