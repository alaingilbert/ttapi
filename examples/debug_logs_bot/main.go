package main

import (
	"os"

	"github.com/alaingilbert/ttapi"
	"github.com/sirupsen/logrus"
)

func main() {
	// This will display debug logs
	logrus.SetLevel(logrus.DebugLevel)

	auth := os.Getenv("TTAPI_AUTH")
	userID := os.Getenv("TTAPI_USER_ID")
	roomID := os.Getenv("TTAPI_ROOM_ID")
	bot := ttapi.NewBot(auth, userID, roomID)
	bot.Start()
}
