package main

import (
	"flag"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

type Server struct {
	Host string
	Port string
	User string
	Pass string
}

type Client struct {
	User string
	Pass string
	Serv Server
}

type Config struct {
	Listen  string
	KeyPath string
	Clients []Client
}

func (c *Config) Auth(user, pass string) bool {
	for _, client := range c.Clients {
		if user == client.User && pass == client.Pass {
			return true
		}
	}
	return false
}
func (c *Config) GetServer(user string) *Server {
	for _, client := range c.Clients {
		if user == client.User {
			return &client.Serv
		}
	}
	return nil
}

var GlobalConfig Config

func init() {
	log.Println("init config")
	var configFile string
	flag.StringVar(&configFile, "config", "./static.yaml", "path to config file")
	flag.Parse()

	b, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatal(err)
	}

	if err := yaml.Unmarshal(b, &GlobalConfig); err != nil {
		log.Fatal(err)
	}

	log.Println("config loaded ", GlobalConfig)
}
