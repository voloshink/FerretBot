package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/voloshink/dggchat"
)

func handleCommand(m dggchat.Message, s *dggchat.Session) {
	if isFerretCommand(m.Message) {
		handleFerretCommand(m, s)
		return
	}

	if strings.HasPrefix(m.Message, "!fwhitelist") {
		handleWhitelist(m, s)
	}

	if strings.HasPrefix(m.Message, "!fblacklist") {
		handleBlacklist(m, s)
	}

	if strings.HasPrefix(m.Message, "!fsource") {
		handleSource(s)
	}

	if strings.HasPrefix(m.Message, "!fuptime") {
		handleUptime(m, s)
	}

	if strings.HasPrefix(m.Message, "!ping") {
		if strings.EqualFold("polecat", m.Sender.Nick) {
			_ = s.SendMessage("FerretLOL")
		}
	}
}

func isFerretCommand(s string) bool {
	for _, c := range ferretCommands {
		if strings.HasPrefix(s, c) {
			return true
		}
	}
	return false
}

func getFerret() (string, error) {
	resp, err := http.Get(ferretEndpoint)
	if err != nil {
		log.Println(err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Api endpoint returned status %d", resp.StatusCode)
		return "", errors.New("api status code")
	}

	var ar apiResp
	json.NewDecoder(resp.Body).Decode(&ar)

	if ar.URL == "" {
		log.Println("Api endpoint did not respond with a url")
		return "", errors.New("bad api response")
	}

	return ar.URL, nil
}

func isWhitelisted(s string) bool {
	return isInList(s, whitelist)
}

func isAdmin(s string) bool {
	return isInList(s, admins)
}

func isInList(s string, list []string) bool {
	for _, v := range list {
		if strings.EqualFold(s, v) {
			return true
		}
	}

	return false
}

func handleFerretCommand(m dggchat.Message, s *dggchat.Session) {
	if isWhitelisted(m.Sender.Nick) {
		timeElapsed := time.Since(lastSent)
		if timeElapsed < messageInterval {
			return
		}

		url, err := getFerret()
		if err != nil {
			return
		}

		if url == lastMessage {
			return
		}

		err = s.SendMessage(fmt.Sprintf("FerretLOL %s FerretLOL", url))
		if err != nil {
			return
		}

		lastSent = time.Now()
		lastMessage = url

	} else {
		timeElapsed := time.Since(lastPM)
		if timeElapsed < pmInterval {
			return
		}

		url, err := getFerret()
		if err != nil {
			return
		}

		err = s.SendPrivateMessage(m.Sender.Nick,
			fmt.Sprintf("Look like you're not whitelisted, have a pm ferret: %s", url))
		if err != nil {
			return
		}

		lastPM = time.Now()
	}
}

func handleWhitelist(m dggchat.Message, s *dggchat.Session) {
	if !isAdmin(m.Sender.Nick) {
		return
	}

	mSlice := strings.Split(m.Message, " ")
	if len(mSlice) != 2 {
		return
	}

	target := mSlice[1]

	var response string
	if isWhitelisted(target) {
		response = fmt.Sprintf("%s already whitelisted FerretLOL", target)
	} else {
		whitelist = append(whitelist, target)
		response = fmt.Sprintf("%s whitelisted FerretLOL", target)
		saveConfig()
	}

	timeElapsed := time.Since(lastSent)

	if timeElapsed < messageInterval {
		return
	}

	if response == lastMessage {
		return
	}

	err := s.SendMessage(response)
	if err != nil {
		return
	}

	lastMessage = response
	lastSent = time.Now()
}

func handleBlacklist(m dggchat.Message, s *dggchat.Session) {
	if !isAdmin(m.Sender.Nick) {
		return
	}

	mSlice := strings.Split(m.Message, " ")
	if len(mSlice) != 2 {
		return
	}

	target := mSlice[1]

	var response string
	if !isWhitelisted(target) {
		response = fmt.Sprintf("%s not whitelisted FerretLOL", target)
	} else {
		for i, v := range whitelist {
			if strings.EqualFold(v, target) {
				whitelist = append(whitelist[:i], whitelist[i+1:]...)
			}
		}
		response = fmt.Sprintf("%s removed from whitelist FerretLOL", target)
		saveConfig()
	}

	timeElapsed := time.Since(lastSent)

	if timeElapsed < messageInterval {
		return
	}

	if response == lastMessage {
		return
	}

	err := s.SendMessage(response)
	if err != nil {
		return
	}

	lastMessage = response
	lastSent = time.Now()
}

func saveConfig() {
	c := config{
		Key:       key,
		Admins:    admins,
		Whitelist: whitelist,
	}

	encoded, err := json.Marshal(&c)
	if err != nil {
		log.Println(err)
		return
	}

	err = ioutil.WriteFile(configFile, encoded, 0644)
	if err != nil {
		log.Println(err)
	}
}

func handleSource(s *dggchat.Session) {
	timeElapsed := time.Since(lastSent)

	if timeElapsed < messageInterval {
		return
	}

	response := githubURL

	if response == lastMessage {
		return
	}

	err := s.SendMessage(fmt.Sprintf("FerretLOL %s FerretLOL", response))
	if err != nil {
		return
	}

	lastMessage = response
	lastSent = time.Now()
}

func handleUptime(m dggchat.Message, s *dggchat.Session) {
	if !isAdmin(m.Sender.Nick) {
		return
	}

	response := fmt.Sprintf("FerretLOL Uptime: %s", time.Since(startTime))

	if response == lastMessage {
		return
	}

	err := s.SendMessage(response)
	if err != nil {
		return
	}

	lastMessage = response
}
