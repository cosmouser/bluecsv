package config

import (
	"encoding/json"
	"flag"
	"os"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

// Config holds the data from the json config file
type Config struct {
	AdminEmail  string
	ExternalURL string
	PathPrefix  string
	LdapURL     string
	LdapBase    string
	LoginKeys   []string
}

// C is the exported config struct in other packages
var C Config

// ConfigPath is the path to the config json
var ConfigPath = flag.String("config", "./config.json", "path to config.json")

// TmplPath is the path to the template files
var TmplPath = flag.String("tmpl", "./templates/", "path to template files")
var port = flag.Int("port", 8100, "port to listen on")

// Port is the port that the server listens on, set via flag
var Port string

// Credentials is the map of access keys and user ids
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
