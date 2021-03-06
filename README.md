# Turntable API

-----

Join us on Discord https://discord.gg/4AA2DqWpVc

-----

A simple go wrapper for the turntable API.  
You'll need to find your `AUTH`, `USERID` and `ROOMID` information with [this bookmarklet](http://alaingilbert.github.com/Turntable-API/bookmarklet.html).

## Installation

```bash
go get github.com/alaingilbert/ttapi
```

## Examples

### Chat bot

This bot responds to anybody who writes "/hello" in the chat.

```go
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
        if evt.Text == "/hello" {
            _ = bot.Speakf("Hey! How are you @%s ?", evt.Name)
        }
    })
    bot.Start()
}
```

More examples here -> https://github.com/alaingilbert/ttapi/tree/master/examples


# Debugging

Add the following line in your main function

```go
logrus.SetLevel(logrus.DebugLevel)
```

That will print on the terminal all the data that you get and all the data that you send.
