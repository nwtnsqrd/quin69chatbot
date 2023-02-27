package main

import (
	"fmt"
	"log"
	"math/rand"
	"strings"

	"github.com/gempir/go-twitch-irc/v4"
)

const (
	MSG_CACHE_SIZE       = 9 // 10-1
	MSG_REPEAT_THRESHOLD = 3
	botUser              = "poenjoyer"
	channel              = "quin69"
)

type cachedMessage struct {
	message string
	sent    bool
}

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
	"@",        // do not repeat reply andies
	"boob",     // set a good example and do not perpetuate the booba
	"partyhat", // temporary
}

var DUNNING_KRUGER_SLICE = []string{
	"donnie krangle",
	"danny cougar",
	"donnie pringles",
	"daniel kramer",
	"david krangler",
	"danny cooper",
}

func dupCount(msgs []string) map[cachedMessage]int {
	dupFreq := make(map[cachedMessage]int)

	for _, item := range msgs {
		k := cachedMessage{message: item, sent: false}
		_, exist := dupFreq[k]

		if exist {
			dupFreq[k] += 1
			continue
		}

		dupFreq[k] = 1
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

func repeatPopularMessages(message twitch.PrivateMessage, client *twitch.Client) {
	// Do not process our own messages
	if message.User.DisplayName == botUser {
		return
	}

	// Count duplicates
	dupMsgs := dupCount(msgCache[:])

	for k, v := range dupMsgs {
		// Check the blacklist to avoid repeating certain messages
		if containsBlacklistedWord(k.message) {
			break
		}

		// When a certain message in the cache has reached the threshold, repeat it.
		// This is triggered way too often when there is intense spam in the chat.
		// It's not a big problem since the messages itself have a cooldown, but it
		// looks annoying in the log.
		// Find out why and / or try to rate limit the repeats.
		if v == MSG_REPEAT_THRESHOLD && (len(k.message) < 200 && len(k.message) > 0) && !k.sent {
			log.Printf("REPEATED %s\n", k.message)
			client.Say(channel, k.message)
			k.sent = true // this does not work
		}
	}
}

func main() {

	// Connect to Twitch IRC
	client := twitch.NewClient(botUser, "oauth:<OAUTHTOKEN>")

	// Init cache state
	cacheWarmed = false

	// Register chat message callback
	client.OnPrivateMessage(func(message twitch.PrivateMessage) {

		// save the message in Postgres
		saveChatMessage(message)

		// If msgCache is full, reset the counter
		if c == MSG_CACHE_SIZE {
			c = 0
		}

		// Append the current message to the cache
		msgCache[c] = message.Message
		c++

		// Replies to specific messages -----------------------------------------------------------

		if strings.Contains(strings.ToLower(message.Message), "danny") && !strings.HasPrefix(message.Message, "@") {
			idx := rand.Intn(len(DUNNING_KRUGER_SLICE))
			client.Say(channel, fmt.Sprintf("classic %s", DUNNING_KRUGER_SLICE[idx]))
			log.Println("Dunning Kruger'd")
		}

		// Only start once the cache has filled
		if !(c > MSG_CACHE_SIZE-1 || cacheWarmed) {
			return
		}

		cacheWarmed = true

		// Participate in chat spam
		repeatPopularMessages(message, client)

	})

	client.Join(channel)

	if err := client.Connect(); err != nil {
		panic(err)
	}

}
