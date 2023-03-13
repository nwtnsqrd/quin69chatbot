package main

import (
	"fmt"
	"os"
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
	cfCounter       int
	mode            *string

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
