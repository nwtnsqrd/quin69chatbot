package main

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/gempir/go-twitch-irc/v4"
	"golang.org/x/time/rate"
)

const (
	MSG_CACHE_SIZE       = 9 // 10-1
	MSG_REPEAT_THRESHOLD = 3
	COOLDOWN_SECONDS     = 3
	BOT_USER             = "poenjoyer"
	CHANNEL              = "quin69"
)

var (
	msgCache    [MSG_CACHE_SIZE]string
	c           int
	client      = twitch.NewClient(BOT_USER, "oauth:uk443gieu4w9tk333q4v1pvljswrd9")
	rLimiter    = rate.Sometimes{First: 1, Interval: COOLDOWN_SECONDS * time.Second}
	cacheWarmed = false

	// Make sure that every ascii letter is lowercase
	blacklist = []string{
		"nigg",
		"fag",
		"black",
		"kkk",
		"::d",     // this emote times you out for 1min
		"http",    // don't repeat linkers
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

	dunning_kruger_slice = []string{
		"donnie krangle",
		"danny cougar",
		"donnie pringles",
		"daniel kramer",
		"david krangler",
		"danny cooper",
	}
)

func dupCount(msgs []string) map[string]int {
	dupFreq := make(map[string]int)

	for _, item := range msgs {
		_, exist := dupFreq[item]

		if !exist {
			dupFreq[item] = 1
			continue
		}

		dupFreq[item] += 1
	}

	return dupFreq
}

func containsBlacklistedWord(msg string) bool {
	// Transform message to lower case
	normalizedMsg := strings.ToLower(msg)

	// Check it against entries in the blacklist
	for i := range blacklist {
		if strings.Contains(normalizedMsg, blacklist[i]) {
			return true
		}
	}

	// Finally, check if message is a command
	return strings.HasPrefix(normalizedMsg, "!")
}

func repeatPopularMessages(message twitch.PrivateMessage) {
	// Do not process our own messages
	if message.User.DisplayName == BOT_USER {
		return
	}

	// Count duplicates
	dupMsgs := dupCount(msgCache[:])

	for k, v := range dupMsgs {
		// When a certain message in the cache has reached the threshold, repeat it.
		if v >= MSG_REPEAT_THRESHOLD && (len(k) < 200 && len(k) > 0) {

			// Check the blacklist to avoid repeating certain messages
			if containsBlacklistedWord(k) {
				continue
			}

			log.Printf("REPEATED %s\n", k)
			client.Say(CHANNEL, k)
		}
	}
}

func containsKeyword(msg twitch.PrivateMessage, keyword string, ignorePrefixedMessage bool) bool {
	if ignorePrefixedMessage {
		return strings.Contains(strings.ToLower(msg.Message), keyword) && !strings.HasPrefix(msg.Message, "@") && !strings.HasPrefix(msg.Message, "!")
	}

	return strings.Contains(strings.ToLower(msg.Message), keyword)
}

func main() {
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

		// ---------------------------- Replies to specific messages ------------------------------
		if containsKeyword(message, "danny", true) {
			idx := rand.Intn(len(dunning_kruger_slice))
			client.Say(CHANNEL, fmt.Sprintf("classic %s", dunning_kruger_slice[idx]))
			log.Println("Dunning Kruger'd")
		}
		// ----------------------------------------------------------------------------------------

		// Only start once the cache has warmed
		if !(c == MSG_CACHE_SIZE || cacheWarmed) {
			return
		}

		cacheWarmed = true

		// Participate in chat spam
		rLimiter.Do(func() {
			repeatPopularMessages(message)
		})

	})

	client.Join(CHANNEL)

	if err := client.Connect(); err != nil {
		panic(err)
	}

}
