package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/gempir/go-twitch-irc/v4"
)

const (
	MSG_CACHE_SIZE       = 9
	MSG_REPEAT_THRESHOLD = 3
	botUser              = "tadpolesalad"
	channel              = "quin69"
)

var msgCache [MSG_CACHE_SIZE]string
var c int
var cacheWarmed bool

var BLACKLIST = []string{
	"nigg",
	"fag",
	"black",
	"kkk",
	"::D",
	"http",
}

func dupCount(msgs []string) map[string]int {
	dupFreq := make(map[string]int)

	for _, item := range msgs {
		_, exist := dupFreq[item]

		if exist {
			dupFreq[item] += 1
		} else {
			dupFreq[item] = 1
		}
	}

	return dupFreq
}

func containsBlacklistedWord(msg string) bool {
	// Transform message to lower case
	normalizedMsg := strings.ToLower(msg)

	// Check it against entries in the blacklist
	for i := range BLACKLIST {
		if strings.Contains(normalizedMsg, BLACKLIST[i]) {
			return true
		}
	}

	// finally, check if message is a command
	return strings.HasPrefix(normalizedMsg, "!")
}

func main() {

	// Connect to Twitch IRC
	client := twitch.NewClient(botUser, "oauth:<OAUTHTOKEN>")

	// Init cache state
	cacheWarmed = false

	// Register chat message callback
	client.OnPrivateMessage(func(message twitch.PrivateMessage) {

		// If msgCache is full, reset the counter
		if c == MSG_CACHE_SIZE {
			c = 0
		}

		// Append the current message to the cache
		msgCache[c] = message.Message
		c++

		log.Println(message.User.DisplayName, message.Message)

		// Only start once the cache has filled
		if c > MSG_CACHE_SIZE-1 || cacheWarmed {
			cacheWarmed = true

			// Count duplicates
			dupMsgs := dupCount(msgCache[:])

			for k, v := range dupMsgs {
				// Check the blacklist to avoid repeating certain messages
				if containsBlacklistedWord(k) {
					continue
				}

				// When a certain message in the cache has reached the threshold, repeat it
				if v == MSG_REPEAT_THRESHOLD {
					fmt.Printf("\nREPEATED %s\n", k)
					client.Say(channel, k)
				}
			}
		}

		saveChatMessage(message)

		if strings.Contains(strings.ToLower(message.Message), "danny") {
			client.Say(channel, "classic donnie krangle")
		}

		if message.Message == "hi" || message.Message == "hi :)" {
			client.Say(channel, fmt.Sprintf("@%s hi :)", message.User.DisplayName))
		}
	})

	client.Join(channel)

	if err := client.Connect(); err != nil {
		panic(err)
	}

}
