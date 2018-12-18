package config

import (
	"encoding/json"
	"flag"
	log "github.com/sirupsen/logrus"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	AdminEmail  string
	ExternalUrl string
	PathPrefix  string
	LdapUrl     string
	LdapBase    string
	LoginKeys   []string
}

var C Config
var ConfigPath = flag.String("config", "./config.json", "path to config.json")
var TmplPath = flag.String("tmpl", "./templates/", "path to template files")
var port = flag.Int("port", 8100, "port to listen on")
var Port string
var Credentials map[string]string

func init() {
	flag.Parse()
	log.WithFields(log.Fields{
		"ConfigPath": *ConfigPath,
		"TmplPath":   *TmplPath,
		"Port":       *port,
	}).Info()
	Port = ":" + strconv.Itoa(*port)
	file, err := os.Open(*ConfigPath)
	if err != nil {
		log.Fatal(err)
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&C)
	if err != nil {
		log.Fatal(err)
	}
	file.Close()
	Credentials = make(map[string]string)
	for _, j := range C.LoginKeys {
		loginPair := strings.Split(j, ":")
		if len(loginPair[0])+len(loginPair[1]) != len(j)-1 {
			log.Fatal("No colons are allowed in access keys")
		}
		Credentials[loginPair[0]] = loginPair[1]
	}
}
