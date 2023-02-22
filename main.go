package main

import (
	"fmt"
	"log"
	"math/rand"
	"strings"

	"github.com/gempir/go-twitch-irc/v4"
)

const (
	MSG_CACHE_SIZE       = 9
	MSG_REPEAT_THRESHOLD = 3
	botUser              = "poenjoyer"
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
	"::D",     // this emote times you out for 1min
	"http",    // don't repeat linkers
	"borg",    // we dont borg
	"hahaa",   // this emote adds toxicity in quins toxicity system
	"sleeper", // do not repeat ResidentSleeper -> toxicity
	"#",       // do not repeat hashtags
	"kys",
	"kill",
	"die",
	"nambla",
	"@",    // do not repeat reply andies
	"boob", // set a good example and do not perpetuate the booba
}

var DUNNING_KRUGER_SLICE = []string{
	"donnie krangle",
	"danny cougar",
	"donnie pringles",
	"daniel kramer",
}

func dupCount(msgs []string) map[string]int {
	dupFreq := make(map[string]int)

	for _, item := range msgs {
		_, exist := dupFreq[item]

		if exist {
			dupFreq[item] += 1
			continue
		}

		dupFreq[item] = 1
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

		//log.Printf("%s -> %s\n", message.User.DisplayName, message.Message)

		// Only start once the cache has filled
		if c > MSG_CACHE_SIZE-1 || cacheWarmed {
			cacheWarmed = true

			// Do not process our own messages
			if message.User.DisplayName == botUser {
				return
			}

			// Count duplicates
			dupMsgs := dupCount(msgCache[:])

			for k, v := range dupMsgs {
				// Check the blacklist to avoid repeating certain messages
				if containsBlacklistedWord(k) {
					break
				}

				// When a certain message in the cache has reached the threshold, repeat it.
				// This is triggered way too often when there is intense spam in the chat.
				// It's not a big problem since the messages itself have a cooldown, but it
				// looks annoying in the log.
				// Find out why and / or try to rate limit the repeats.
				if v == MSG_REPEAT_THRESHOLD && len(k) < 200 {
					fmt.Printf("REPEATED %s\n", k)
					client.Say(channel, k)
				}
			}
		}

		saveChatMessage(message)

		// Replies to specific messages -----------------------------------------------------------

		if strings.Contains(strings.ToLower(message.Message), "danny") && !strings.HasPrefix(message.Message, "@") {
			idx := rand.Intn(len(DUNNING_KRUGER_SLICE))
			client.Say(channel, fmt.Sprintf("classic %s", DUNNING_KRUGER_SLICE[idx]))
			log.Println("Dunning Kruger'd")
		}

		if message.Message == "hi" || message.Message == "hi :)" {
			client.Say(channel, fmt.Sprintf("@%s hi :)", message.User.DisplayName))
			log.Printf("Said hi to %s", message.User.DisplayName)
		}
	})

	client.Join(channel)

	if err := client.Connect(); err != nil {
		panic(err)
	}

}
