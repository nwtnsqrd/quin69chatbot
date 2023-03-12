package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"regexp"
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
)

var (
	// Twitch user that is authenticated to use this bot
	BotUser = os.Getenv("QUINBOT_USER")

	// Twitch channel that BotUser should connect to
	Channel = os.Getenv("QUINBOT_CHANNEL")

	msgCache        [MsgCacheSize]string
	c               int
	client          = twitch.NewClient(BotUser, fmt.Sprintf("oauth:%s", os.Getenv("QUINBOT_OAUTH")))
	mainRLimiter    = rate.Sometimes{First: 1, Interval: MainCooldownSeconds * time.Second}
	replyRLimiter   = rate.Sometimes{First: 1, Interval: ReplyCooldownSeconds * time.Second}
	cacheWarmed     = false
	lastMessageSent string

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

func countDuplicateMessages(msgs []string) map[string]int {
	dupFreq := make(map[string]int)

	for _, item := range msgs {
		i := strings.TrimSpace(item)

		if _, exist := dupFreq[i]; !exist {
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
	dupMsgs := countDuplicateMessages(msgCache[:])

	for msg, ct := range dupMsgs {
		// When a certain message in the cache has reached the threshold, repeat it.
		// Do not repeat the same message twice in a row
		if ct >= MsgRepeatThreshold && msg != lastMessageSent {

			// Check the blacklist to avoid repeating certain messages
			if containsBlacklistedWord(msg) {
				continue
			}

			log.Printf("%s: %s\n", BotUser, msg)
			client.Say(Channel, msg)
			lastMessageSent = msg
		}
	}
}

func matchesExpression(msg twitch.PrivateMessage, expr string, ignorePrefixedMessage bool) bool {
	match, _ := regexp.Match(expr, []byte(strings.ToLower(msg.Message)))

	if ignorePrefixedMessage {
		return match && !strings.HasPrefix(msg.Message, "@") && !strings.HasPrefix(msg.Message, "!")
	}

	return match
}

func uniqueTokensFromMessage(message twitch.PrivateMessage) []string {
	allTokens := strings.Split(strings.ToLower(strings.TrimSpace(message.Message)), " ")
	keys := make(map[string]bool)
	list := []string{}

	for _, item := range allTokens {
		if containsBlacklistedWord(item) {
			continue
		}
		if _, exist := keys[item]; !exist {
			keys[item] = true
			list = append(list, item)
		}
	}

	return list
}

func main() {
	mode := flag.String("mode", "offlinechat", "use `--mode live` or `--mode offlinechat`")
	flag.Parse()

	// Register chat message callback
	client.OnPrivateMessage(func(message twitch.PrivateMessage) {

		// save the message in Postgres
		saveChatMessage(message)

		// If msgCache is full, reset the counter
		if c == MsgCacheSize {
			c = 0
		}

		// Append the current message to the cache
		msgCache[c] = message.Message
		c++

		// ---------------------------- Replies to specific messages ------------------------------
		if matchesExpression(message, `^.*danny\s.*$`, true) {
			replyRLimiter.Do(func() {
				go func() {
					time.Sleep(ReplyDelaySeconds * time.Second)
					idx := rand.Intn(len(dunning_kruger_slice))
					client.Say(Channel, fmt.Sprintf("classic %s", dunning_kruger_slice[idx]))
					log.Println("Dunning Kruger'd")
				}()
			})
		}

		if matchesExpression(message, `^hi\s.*$`, true) {
			replyRLimiter.Do(func() {
				go func() {
					time.Sleep(ReplyDelaySeconds * time.Second)
					client.Say(Channel, fmt.Sprintf("@%s hi :)", message.User.DisplayName))
					log.Println("Said hi to", message.User.DisplayName)
				}()
			})
		}

		if matchesExpression(message, `^siro\sperma.*$`, true) {
			replyRLimiter.Do(func() {
				go func() {
					time.Sleep(ReplyDelaySeconds * time.Second)
					client.Say(Channel, "SirO Tssk PERMANENT BANISHMENT")
					log.Println("PERMANENT BANISHMENT")
				}()
			})
		}

		if *mode == "offlinechat" {
			if matchesExpression(message, fmt.Sprintf("^.*%s.*$", BotUser), false) {
				replyRLimiter.Do(func() {
					go func() {
						client.Say(Channel, fmt.Sprintf("@%s I am currently in unmanned BOT mode. ttyl peepoCute", message.User.DisplayName))
						log.Println("Told", message.User.DisplayName, "that I'm away")
					}()
				})
			}
		}
		// ----------------------------------------------------------------------------------------

		// Only start once the cache has warmed
		if !(c == MsgCacheSize || cacheWarmed) {
			return
		}

		// Print when the cache has warmed and the bot is ready to spam
		if !cacheWarmed {
			log.Println("Cache is warmed up")
			cacheWarmed = !cacheWarmed
		}

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
