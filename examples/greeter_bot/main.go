package main

import (
	"os"
	"time"

	"github.com/alaingilbert/ttapi"
)

func main() {
	auth := os.Getenv("TTAPI_AUTH")
	userID := os.Getenv("TTAPI_USER_ID")
	roomID := os.Getenv("TTAPI_ROOM_ID")
	bot := ttapi.NewBot(auth, userID, roomID)
	bot.OnRegistered(func(evt ttapi.RegisteredEvt) {
		user := evt.User[0]
		// Do not greet self !
		if user.ID == userID {
			return
		}
		time.Sleep(5 * time.Second)
		_ = bot.Speakf("Welcome here @%s !", user.Name)
	})
	bot.Start()
}
