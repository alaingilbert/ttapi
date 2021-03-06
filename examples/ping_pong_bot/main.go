package main

import (
	"os"

	"github.com/alaingilbert/ttapi"
)

func main() {
	auth := os.Getenv("TTAPI_AUTH")
	userID := os.Getenv("TTAPI_USER_ID")
	roomID := os.Getenv("TTAPI_ROOM_ID")
	bot := ttapi.NewBot(auth, userID, roomID)
	bot.OnSpeak(func(evt ttapi.SpeakEvt) {
		if evt.Text == "/ping" {
			_ = bot.Speak("pong")
		}
	})
	bot.Start()
}
