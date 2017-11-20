package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/voloshink/dggchat"
)

const (
	defaultConfigFile = "config.json"
	ferretEndpoint    = "https://polecat.me/api/ferret"
	githubURL         = "https://github.com/voloshink/FerretBot"
)

var (
	ferretCommands  = []string{"!polecat", "!ferret", "! FerretLOL"}
	lastMessage     = ""
	lastSent        = time.Now()
	lastPM          = time.Now()
	startTime       = time.Now()
	messageInterval = time.Minute
	pmInterval      = time.Second * 30
	configFile      string
	admins          []string
	whitelist       []string
	key             string
)

type (
	config struct {
		Key       string   `json:"login_key"`
		Admins    []string `json:"admins"`
		Whitelist []string `json:"whitelist"`
	}

	apiResp struct {
		URL string `json:"url"`
	}
)

func main() {
	var file string
	if len(os.Args) < 2 {
		file = defaultConfigFile
	} else {
		file = os.Args[1]
	}

	f, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatalln(err)
	}

	var c config
	err = json.Unmarshal(f, &c)
	if err != nil {
		log.Fatalln(err)
	}

	configFile = file

	if c.Key == "" {
		log.Fatalln("No login key provided")
	}

	key = c.Key

	if c.Admins == nil {
		c.Admins = make([]string, 0)
	}

	admins = c.Admins

	if c.Whitelist == nil {
		c.Whitelist = make([]string, 0)
	}

	whitelist = c.Whitelist

	go startBot(c.Key)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT)
	<-sc
}

func startBot(key string) {
	dgg, err := dggchat.New(key)
	if err != nil {
		log.Fatalln(err)
	}

	err = dgg.Open()
	if err != nil {
		log.Fatalln(err)
	}

	defer dgg.Close()

	messages := make(chan dggchat.Message)
	errors := make(chan string)

	dgg.AddMessageHandler(func(m dggchat.Message, s *dggchat.Session) {
		messages <- m
	})

	dgg.AddErrorHandler(func(e string, s *dggchat.Session) {
		errors <- e
	})

	for {
		select {
		case m := <-messages:
			if strings.HasPrefix(m.Message, "!") {
				handleCommand(m, dgg)
			}
		case e := <-errors:
			log.Printf("Error %s\n", e)
		}
	}

}
