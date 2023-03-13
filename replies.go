package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/gempir/go-twitch-irc/v4"
)

func checkMessageForReplies(message twitch.PrivateMessage) {
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

	if matchesExpression(message, `^cumfartt?$`, true) {
		//||	(message.User.DisplayName == "krnjombi" && matchesExpression(message, `^.*cock$`, true)) {
		replyRLimiter.Do(func() {
			go func() {
				cfCounter++
				client.Say(Channel, fmt.Sprintf("@%s cumfarted %d times in todays stream PogU", message.User.DisplayName, cfCounter))
				log.Println("CUMFART", cfCounter)
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
}
