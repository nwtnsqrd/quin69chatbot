package main

import (
	"log"

	"github.com/gempir/go-twitch-irc/v4"
)

func main() {

	client := twitch.NewAnonymousClient()

	client.OnPrivateMessage(func(message twitch.PrivateMessage) {
		log.Println(message.User.DisplayName, message.Message)
		saveChatMessage(message)
	})

	client.Join("quin69")

	if err := client.Connect(); err != nil {
		panic(err)
	}

}
