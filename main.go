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
	// Amount of messages that are being cached to find the most popular ones
	MsgCacheSize = 9 // 10-1

	// Threshold of how many times the same message has to be in msgCache in order to get repeated
	MsgRepeatThreshold = 3

	// Cooldown to ratelimit spamming chat messages
	MainCooldownSeconds = 3

	// Cooldown to ratelimit sending replies
	ReplyCooldownSeconds = 8

	// Time before a reply is actually sent
	ReplyDelaySeconds = 2

	// Twitch user that is authenticated to use this bot
	BotUser = "poenjoyer"

	// Twitch channel that BOT_USER should connect to
	Channel = "quin69"
)

var (
	msgCache        [MsgCacheSize]string
	c               int
	client          = twitch.NewClient(BotUser, "oauth:uk443gieu4w9tk333q4v1pvljswrd9")
	mainRLimiter    = rate.Sometimes{First: 1, Interval: MainCooldownSeconds * time.Second}
	replyRLimiter   = rate.Sometimes{First: 1, Interval: ReplyCooldownSeconds * time.Second}
	cacheWarmed     = false
	lastMessageSent string

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
		"@",    // do not repeat reply andies
		"boob", // set a good example and do not perpetuate the booba
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
		i := strings.TrimSpace(item)

		_, exist := dupFreq[i]

		if !exist {
			dupFreq[i] = 1
			continue
		}

		dupFreq[i] += 1
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
	if message.User.DisplayName == BotUser {
		return
	}

	// Count duplicates
	dupMsgs := dupCount(msgCache[:])

	for k, v := range dupMsgs {
		// When a certain message in the cache has reached the threshold, repeat it.
		// Do not repeat the same message twice in a row
		if v >= MsgRepeatThreshold && k != lastMessageSent {

			// Check the blacklist to avoid repeating certain messages
			if containsBlacklistedWord(k) {
				continue
			}

			log.Printf("%s: %s\n", BotUser, k)
			client.Say(Channel, k)
			lastMessageSent = k
		}
	}
}

func containsKeyword(msg twitch.PrivateMessage, keyword string, ignorePrefixedMessage bool) bool {
	normalizedMsg := strings.ToLower(msg.Message)

	if ignorePrefixedMessage {
		return strings.Contains(normalizedMsg, keyword) && !strings.HasPrefix(msg.Message, "@") && !strings.HasPrefix(msg.Message, "!")
	}

	return strings.Contains(normalizedMsg, keyword)
}

func main() {
	// Register chat message callback
	client.OnPrivateMessage(func(message twitch.PrivateMessage) {

		// save the message in Postgres
		// saveChatMessage(message)

		// If msgCache is full, reset the counter
		if c == MsgCacheSize {
			c = 0
		}

		// Append the current message to the cache
		msgCache[c] = message.Message
		c++

		// ---------------------------- Replies to specific messages ------------------------------
		if containsKeyword(message, "danny ", true) {
			replyRLimiter.Do(func() {
				go func() {
					time.Sleep(ReplyDelaySeconds * time.Second)
					idx := rand.Intn(len(dunning_kruger_slice))
					client.Say(Channel, fmt.Sprintf("classic %s", dunning_kruger_slice[idx]))
					log.Println("Dunning Kruger'd")
				}()
			})
		}

		if containsKeyword(message, "hi ", true) {
			replyRLimiter.Do(func() {
				go func() {
					time.Sleep(ReplyDelaySeconds * time.Second)
					client.Say(Channel, fmt.Sprintf("@%s hi :)", message.User.DisplayName))
					log.Println("Said hi to", message.User.DisplayName)
				}()
			})
		}

		if containsKeyword(message, "SirO perma", true) {
			replyRLimiter.Do(func() {
				go func() {
					time.Sleep(ReplyDelaySeconds * time.Second)
					client.Say(Channel, "SirO PERMANENT BANISHMENT")
					log.Println("PERMANENT BANISHMENT")
				}()
			})
		}
		// ----------------------------------------------------------------------------------------

		// Only start once the cache has warmed
		if !(c == MsgCacheSize || cacheWarmed) {
			return
		}

		cacheWarmed = true

		// Participate in chat spam
		mainRLimiter.Do(func() {
			repeatPopularMessages(message)
		})

	})

	client.Join(Channel)

	if err := client.Connect(); err != nil {
		panic(err)
	}

}
