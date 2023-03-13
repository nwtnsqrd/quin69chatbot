package main

import (
	"log"
	"regexp"
	"strings"

	"github.com/gempir/go-twitch-irc/v4"
)

// countDuplicateMessages constructs a map[string]int from `msgs` where `v` from map[string]int
// is the amount of of times the same string occurs in `msgs`. The resulting map therefore
// contains all unique strings from `msgs` and assigns a count to them.
func countDuplicateMessages(msgs []string) map[string]int {

	dupFreq := make(map[string]int)

	for _, item := range msgs {
		i := strings.TrimSpace(item)

		// Check if key already exists in `dupFreq`
		// If it doesn't, set `v` for `k` to 1
		if _, exist := dupFreq[i]; !exist {
			dupFreq[i] = 1
			continue
		}

		// If it does exist, increment `v` by 1
		dupFreq[i] += 1
	}

	return dupFreq
}

// containsBlacklistedWord checks `msg` against words in the `blacklist` slice.
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

// repeatPopularMessages counts all duplicates in `msgCache`. If any item in `msgCache`
// reached or is above `MsgRepeatThreshold`, it is going to be sent to `Channel`.
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

			// Check `msg` agains entries in `blacklist`
			if containsBlacklistedWord(msg) {
				continue
			}

			log.Printf("%s: %s\n", BotUser, msg)
			client.Say(Channel, msg)
			lastMessageSent = msg
		}
	}
}

// matchesExpression compares `msg.Message` to the regex expression `expr`. If `ignorePrefixedMessage`
// is set to `true`, commands and replies are ignored and will not match `expr`.
func matchesExpression(msg twitch.PrivateMessage, expr string, ignorePrefixedMessage bool) bool {
	match, _ := regexp.Match(expr, []byte(strings.ToLower(msg.Message)))

	if ignorePrefixedMessage {
		return match && !strings.HasPrefix(msg.Message, "@") && !strings.HasPrefix(msg.Message, "!")
	}

	return match
}

// uniqueTokensFromMessage tokenzies `message.Message` and returns a slice with all unique tokens
// in `message.Message`.
func uniqueTokensFromMessage(message twitch.PrivateMessage) []string {

	// Split the message into seperate words
	allTokens := strings.Split(strings.ToLower(strings.TrimSpace(message.Message)), " ")
	keys := make(map[string]bool)
	list := []string{}

	for _, item := range allTokens {
		// Ignore tokens that are in `blacklist`
		if containsBlacklistedWord(item) {
			continue
		}
		// Check if key already exists in `keys`
		// If it doesn't, set `v` for `item` to true and append it to slice `list`
		if _, exist := keys[item]; !exist {
			keys[item] = true
			list = append(list, item)
		}
	}

	return list
}
