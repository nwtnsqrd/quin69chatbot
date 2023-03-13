package main

import (
	"flag"
	"log"

	"github.com/gempir/go-twitch-irc/v4"
)

func main() {
	// Register chat message callback
	client.OnPrivateMessage(func(message twitch.PrivateMessage) {

		// save the message in Postgres
		saveChatMessage(message)

		// If `msgCache` is full, reset the counter
		if c == MsgCacheSize {
			c = 0
		}

		// Append the current message to the cache
		msgCache[c] = message.Message
		c++

		// Check if we have a custom reply to the current message
		checkMessageForReplies(message)

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

func init() {
	mode = flag.String("mode", "offlinechat", "use `--mode live` or `--mode offlinechat`")
	flag.Parse()
}
