package ttapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/net/websocket"
)

var heartbeatRgx = regexp.MustCompile(`~m~[0-9]+~m~(~h~[0-9]+)`)
var lenRgx = regexp.MustCompile(`^~m~([0-9]+)~m~`)

// Bot is a thread safe client for turntable \o/
// To get the auth, user id and room id, you can use the following bookmarklet
// http://alaingilbert.github.io/Turntable-API/bookmarklet.html
type Bot struct {
	auth            string                   // auth id, can be retrieved using bookmarklet
	userID          string                   // user id, can be retrieved using bookmarklet
	roomID          string                   // room id, can be retrieved using bookmarklet
	client          string                   // web
	laptop          string                   // mac
	logWs           bool                     // either or not to log websocket messages
	msgID           int                      // keep track of message id used to communicate with ws server
	clientID        string                   // random string
	currentStatus   string                   // available/unavailable/away used for the chat
	lastHeartbeat   time.Time                // keep track of last heartbeat timestamp
	lastActivity    time.Time                // keep track of last received message timestamp
	ws              *websocket.Conn          // websocket connection to turntable
	unackMsgs       []UnackMsg               // list of messages sent that are not acknowledged by the ws server
	currentSearches []Search                 // Current searches in progress
	callbacks       map[string][]interface{} // user defined callbacks set for each events
	ctx             context.Context          // bot context
	cancel          context.CancelFunc       // cancel function to stop bot
	CurrentSongID   string                   // cached current song id
	CurrentDjID     string                   // cached current dj id
	tmpSong         H                        // cached song fake message, used to emit our own fake event (endsong)
	txCh            chan TxMsg               // messages to transmit to turntable
	rxCh            chan RxMsg               // messages received from turntable
}

// NewBot creates a new bot
func NewBot(auth, userID, roomID string) *Bot {
	b := new(Bot)
	b.auth = auth
	b.userID = userID
	b.roomID = roomID
	b.currentStatus = available
	b.client = webClient
	b.laptop = macLaptop
	b.lastHeartbeat = time.Now()
	b.lastActivity = time.Now()
	b.clientID = strconv.FormatInt(time.Now().Unix(), 10) + "-" + strconv.FormatFloat(rand.Float64(), 'f', 17, 64)
	b.callbacks = make(map[string][]interface{})
	b.ctx, b.cancel = context.WithCancel(context.Background())
	b.txCh = make(chan TxMsg, 10)
	b.rxCh = make(chan RxMsg, 10)
	return b
}

func (b *Bot) startWS() {
	defer b.cancel()
	var err error
	b.ws, err = websocket.Dial(wssURL, wsProtocol, wsOrigin)
	if err != nil {
		logrus.Error("failed to dial websocket:", err)
		return
	}
	b.readWS()
}

func (b *Bot) readWS() {
	var msg []byte
	var msgLen, msgRead int
LOOP:
	for {
		select {
		case <-b.ctx.Done():
			break LOOP
		default:
		}
		var buf = make([]byte, 1024*1024)
		if err := b.ws.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
			logrus.Error("failed to set read deadline:", err)
		}
		n, err := b.ws.Read(buf)
		if err != nil {
			if err == io.EOF {
				logrus.Error("socket eof:", err)
				break
			} else if strings.HasSuffix(err.Error(), "use of closed network connection") {
				break
			} else if strings.HasSuffix(err.Error(), "i/o timeout") {
				continue
			} else {
				logrus.Error("socket unexpected error", err)
				// connection reset by peer
				break
			}
		}
		packet := buf[0:n]
		if bytes.HasPrefix(packet, []byte("~m~")) {
			m := lenRgx.FindSubmatch(packet)
			msgLen, _ = strconv.Atoi(string(m[1]))
			msgRead = -6 - len(m[1])
			msg = make([]byte, 0)
		}
		msg = append(msg, packet...)
		msgRead += n
		if msgRead == msgLen {
			b.rx(msg)
		}
	}
}

func (b *Bot) processMessage(msg []byte) {
	logrus.Debug("> " + string(msg))
	if isHeartbeat(msg) {
		b.processHeartbeat(msg)
		return
	}
	if bytes.Equal(msg, []byte("~m~10~m~no_session")) {
		b.emit(ready, nil)
		SGo(func() {
			_ = b.updatePresence()
			_ = b.userModify(H{"laptop": b.laptop})
			if b.roomID != "" {
				if err := b.roomRegister(b.roomID); err != nil {
					logrus.Error(err)
					b.roomID = ""
				}
			}
		})
		return
	}
	b.lastActivity = time.Now()

	rawJson, err := b.extractMessageJson(msg)
	if err != nil {
		logrus.Error(err)
		return
	}

	var jsonHashMap map[string]interface{}
	if err := json.Unmarshal(rawJson, &jsonHashMap); err != nil {
		logrus.Error(err)
		return
	}

	b.executeCallback(rawJson, jsonHashMap)
	b.processCommand(rawJson, jsonHashMap)
}

func isHeartbeat(msg []byte) bool {
	return heartbeatRgx.Match(msg)
}

func getHeartbeatID(data []byte) (string, error) {
	matches := heartbeatRgx.FindSubmatch(data)
	if len(matches) != 2 {
		return "", errors.New("invalid heartbeat : " + truncStr(string(data), 30, "..."))
	}
	return string(matches[1]), nil
}

func (b *Bot) processHeartbeat(data []byte) {
	heartbeatID, err := getHeartbeatID(data)
	if err != nil {
		logrus.Error(err)
		return
	}
	payload := `~m~` + strconv.Itoa(len(heartbeatID)) + `~m~` + heartbeatID
	_, _ = b.ws.Write([]byte(payload))
	b.lastHeartbeat = time.Now()
	SGo(func() { _ = b.updatePresence() })
}

func (b *Bot) processCommand(rawJson []byte, jsonHashMap map[string]interface{}) {
	if command, ok := jsonHashMap["command"].(string); ok {
		switch command {
		case remDJ:
			if modID, ok := jsonHashMap["modid"].(string); ok {
				if modID != "" {
					b.emit(escort, rawJson)
				}
			}
		case nosong:
			b.CurrentDjID = ""
			b.CurrentSongID = ""
			b.emit(endsong, b.tmpSong)
		case newsong:
			if b.CurrentSongID != "" {
				b.emit(endsong, b.tmpSong)
			}
			b.CurrentDjID = castStr(safeMapPath(jsonHashMap, "room.metadata.current_dj"))
			b.CurrentSongID = castStr(safeMapPath(jsonHashMap, "room.metadata.current_song._id"))
			if m, ok := jsonHashMap["room"].(map[string]interface{}); ok {
				b.setTmpSong(m)
			}
		case updateVotes:
			if b.tmpSong != nil {
				ups := safeMapPath(jsonHashMap, "room.metadata.upvotes")
				downs := safeMapPath(jsonHashMap, "room.metadata.downvotes")
				ls := safeMapPath(jsonHashMap, "room.metadata.listeners")
				if _, ok := b.tmpSong["room"].(map[string]interface{}); ok {
					if _, ok := b.tmpSong["room"].(map[string]interface{})["metadata"].(map[string]interface{}); ok {
						b.tmpSong["room"].(map[string]interface{})["metadata"].(map[string]interface{})["upvotes"] = ups
						b.tmpSong["room"].(map[string]interface{})["metadata"].(map[string]interface{})["downvotes"] = downs
						b.tmpSong["room"].(map[string]interface{})["metadata"].(map[string]interface{})["listeners"] = ls
					} else {
						b.tmpSong["room"].(map[string]interface{})["metadata"] = map[string]interface{}{"upvotes": ups, "downvotes": downs, "listeners": ls}
					}
				} else {
					b.tmpSong["room"] = map[string]interface{}{"metadata": map[string]interface{}{"upvotes": ups, "downvotes": downs, "listeners": ls}}
				}
			}
		case searchComplete:
			b.executeSearchCallbacks(rawJson, jsonHashMap)
		}
		b.emit(command, rawJson)
	}
}

// Forward search results to the requested callback when we find a match.  We
// look to see if the query string in the results matches the query string with
// our in-progress search slice.
func (b *Bot) executeSearchCallbacks(rawJson []byte, jsonHashMap map[string]interface{}) {
	for idx, search := range b.currentSearches {
		if resultQuery, ok := jsonHashMap["query"].(string); ok {
			if search.Query == resultQuery {
				if search.Callback != nil {
					SGo(func() { search.Callback(rawJson) })
				}
				b.currentSearches = append(b.currentSearches[:idx], b.currentSearches[idx+1:]...) // remove element at idx
				break
			}
		}
	}
}

// Check all unack messages, if we received a response message from the socket server,
// we will execute the callback if any was provided.
func (b *Bot) executeCallback(rawJson []byte, jsonHashMap map[string]interface{}) {
	for idx, unackMsg := range b.unackMsgs {
		if jid, ok := jsonHashMap["msgid"].(float64); ok {
			if unackMsg.MsgID == int(jid) {

				isSearch := false

				// Extra logic for specific events
				switch unackMsg.Payload["api"].(string) {
				case roomRegister:
					if success, ok := jsonHashMap["success"].(bool); ok && success {
						b.roomID = castStr(unackMsg.Payload["roomid"])
						SGo(func() {
							roomInfoRaw := b.roomInfoRaw()
							var roomInfoHash H
							var roomInfo RoomInfoRes
							if err := json.Unmarshal(roomInfoRaw, &roomInfoHash); err != nil {
								logrus.Error(err)
							}
							if err := json.Unmarshal(roomInfoRaw, &roomInfo); err != nil {
								logrus.Error(err)
							}
							b.setTmpSong(roomInfoHash)
							b.emit(roomChanged, roomInfo)
						})
					}
				case roomInfo:
					b.CurrentDjID = castStr(safeMapPath(jsonHashMap, "room.metadata.current_dj"))
					b.CurrentSongID = castStr(safeMapPath(jsonHashMap, "room.metadata.current_song._id"))
				case fileSearch:
					if success, ok := jsonHashMap["success"].(bool); ok && success {
						if query, ok := unackMsg.Payload["query"].(string); ok {
							isSearch = true
							b.currentSearches = append(b.currentSearches, Search{Query: query, Callback: unackMsg.Callback})
						}
					}
				}

				// Execute callback if provided
				if !isSearch && unackMsg.Callback != nil {
					unackMsg.Callback(rawJson)
				}

				// Remove element from slice, since we received the response for that message
				b.unackMsgs = append(b.unackMsgs[:idx], b.unackMsgs[idx+1:]...) // remove element at idx
				break
			}
		}
	}
}

func (b *Bot) setTmpSong(room map[string]interface{}) {
	b.tmpSong = H{"command": endsong, "room": room, "success": true}
}

// Extract message length. eg: `~m~457~m~` -> 457
func getMessageLen(msg []byte) (int, error) {
	matches := lenRgx.FindSubmatch(msg)
	if len(matches) != 2 {
		return 0, errors.New("failed to find message length : " + truncStr(string(msg), 15, "..."))
	}
	msgLen := doParseInt(string(matches[1]))
	return msgLen, nil
}

// Extract the json part of a websocket message
func (b *Bot) extractMessageJson(msg []byte) ([]byte, error) {
	msgLen, err := getMessageLen(msg)
	if err != nil {
		return []byte(""), err
	}
	startIdx := bytes.Index(msg, []byte("{"))
	rawJson := msg[startIdx : startIdx+msgLen]
	return rawJson, nil
}

// UnackMsg store information about a message we sent that is not ack yet
type UnackMsg struct {
	MsgID    int                  // Message ID sent to socket server
	Payload  H                    // Payload that we sent to socket server
	Callback func(rawJson []byte) // Callback to receive answer from socket server
}

// Search store any searches currently in progress
type Search struct {
	Query    string               // Query that was sent to search
	Callback func(rawJson []byte) // Callback to receive search results
}

// Send a payload to the WS server
func (b *Bot) send(payload H, callback func([]byte)) {
	msgID := b.msgID
	payload["msgid"] = msgID
	payload["clientid"] = b.clientID
	payload["userid"] = b.userID
	payload["userauth"] = b.auth
	payload["client"] = b.client
	by, err := json.Marshal(payload)
	if err != nil {
		logrus.Error(err)
		return
	}

	logrus.Debug("< " + string(by))

	if _, err := b.ws.Write([]byte(`~m~` + strconv.Itoa(len(by)) + `~m~` + string(by))); err != nil {
		logrus.Error(err)
		return
	}
	b.unackMsgs = append(b.unackMsgs, UnackMsg{MsgID: msgID, Payload: payload, Callback: callback})
	b.msgID++
}

// TxMsg ...
type TxMsg struct {
	Payload  H
	Callback func([]byte)
}

// RxMsg ...
type RxMsg struct {
	Msg []byte
}

func (b *Bot) tx(payload H, res interface{}) {
	ctx, cancel := context.WithCancel(context.Background())
	clb := func(rawJson []byte) {
		switch v := res.(type) {
		case *[]byte:
			*v = rawJson
		default:
			if err := json.Unmarshal(rawJson, &res); err != nil {
				logrus.Error(err)
			}
		}
		cancel()
	}
	b.txCh <- TxMsg{Payload: payload, Callback: clb}
	select {
	case <-ctx.Done():
	case <-b.ctx.Done():
	}
}

func (b *Bot) rx(msg []byte) {
	b.rxCh <- RxMsg{Msg: msg}
}

// Stop the bot
func (b *Bot) Stop() {
	b.cancel()
}

// Start the bot
func (b *Bot) Start() {
	SGo(b.startWS)
	for {
		select {
		case rxMsg := <-b.rxCh:
			b.processMessage(rxMsg.Msg)
		case txMsg := <-b.txCh:
			b.send(txMsg.Payload, txMsg.Callback)
		case <-b.ctx.Done():
			return
		}
	}
}

// emit events to bot listeners
func (b *Bot) emit(cmd string, data interface{}) {
	for _, clb := range b.callbacks[cmd] {
		func(clb interface{}) {
			if dataBy, ok := data.([]byte); ok {
				if clbBy, ok := clb.(func([]byte)); ok {
					SGo(func() { clbBy(dataBy) })
				} else {
					if cmd == registered {
						var payload RegisteredEvt
						_ = json.Unmarshal(dataBy, &payload)
						SGo(func() { clb.(func(RegisteredEvt))(payload) })
					} else if cmd == pmmed {
						var payload PmmedEvt
						_ = json.Unmarshal(dataBy, &payload)
						SGo(func() { clb.(func(PmmedEvt))(payload) })
					} else if cmd == newsong {
						var payload NewSongEvt
						_ = json.Unmarshal(dataBy, &payload)
						SGo(func() { clb.(func(NewSongEvt))(payload) })
					} else if cmd == nosong {
						var payload NoSongEvt
						_ = json.Unmarshal(dataBy, &payload)
						SGo(func() { clb.(func(NoSongEvt))(payload) })
					} else if cmd == snagged {
						var payload SnaggedEvt
						_ = json.Unmarshal(dataBy, &payload)
						SGo(func() { clb.(func(SnaggedEvt))(payload) })
					} else if cmd == bootedUser {
						var payload BootedUserEvt
						_ = json.Unmarshal(dataBy, &payload)
						SGo(func() { clb.(func(BootedUserEvt))(payload) })
					} else if cmd == updateVotes {
						var payload UpdateVotesEvt
						_ = json.Unmarshal(dataBy, &payload)
						SGo(func() { clb.(func(UpdateVotesEvt))(payload) })
					} else if cmd == deregistered {
						var payload DeregisteredEvt
						_ = json.Unmarshal(dataBy, &payload)
						SGo(func() { clb.(func(DeregisteredEvt))(payload) })
					} else if cmd == addDJ {
						var payload AddDJEvt
						_ = json.Unmarshal(dataBy, &payload)
						SGo(func() { clb.(func(AddDJEvt))(payload) })
					} else if cmd == remDJ {
						var payload RemDJEvt
						_ = json.Unmarshal(dataBy, &payload)
						SGo(func() { clb.(func(RemDJEvt))(payload) })
					} else if cmd == escort {
						var payload EscortEvt
						_ = json.Unmarshal(dataBy, &payload)
						SGo(func() { clb.(func(EscortEvt))(payload) })
					} else if cmd == newModerator {
						var payload NewModeratorEvt
						_ = json.Unmarshal(dataBy, &payload)
						SGo(func() { clb.(func(NewModeratorEvt))(payload) })
					} else if cmd == remModerator {
						var payload RemModeratorEvt
						_ = json.Unmarshal(dataBy, &payload)
						SGo(func() { clb.(func(RemModeratorEvt))(payload) })
					} else if cmd == speak {
						var payload SpeakEvt
						_ = json.Unmarshal(dataBy, &payload)
						if payload.UserID == b.userID {
							return
						}
						SGo(func() { clb.(func(SpeakEvt))(payload) })
					}
				}
			} else if roomInfo, ok := data.(RoomInfoRes); ok {
				if cmd == roomChanged {
					SGo(func() { clb.(func(RoomInfoRes))(roomInfo) })
				}
			} else if m, ok := data.(H); ok {
				if cmd == endsong {
					SGo(func() { clb.(func(H))(m) })
				}
			} else {
				if cmd == ready {
					SGo(func() { clb.(func())() })
				}
			}
		}(clb)
	}
}

func (b *Bot) addCallback(cmd string, clb interface{}) {
	b.callbacks[cmd] = append(b.callbacks[cmd], clb)
}

func baseErr(p BaseRes) error {
	if !p.Success {
		return errors.New(p.Err)
	}
	return nil
}

func txBaseErr(b *Bot, h H) error {
	var res BaseRes
	b.tx(h, &res)
	return baseErr(res)
}

//-----------------------------------------------------------------------------

func (b *Bot) search(query string) (out SearchRes, err error) {
	b.tx(H{"api": fileSearch, "query": query}, &out)
	return out, baseErr(out.BaseRes)
}

func (b *Bot) speak(msg string) error {
	return txBaseErr(b, H{"api": roomSpeak, "roomid": b.roomID, "text": msg})
}

func (b *Bot) speakf(format string, args ...interface{}) error {
	return b.speak(fmt.Sprintf(format, args...))
}

func (b *Bot) pm(userID, msg string) error {
	return txBaseErr(b, H{"api": pmSend, "receiverid": userID, "text": msg})
}

func (b *Bot) modifyName(newName string) error {
	return txBaseErr(b, H{"api": userModify, "name": newName})
}

func (b *Bot) modifyLaptop(laptop string) error {
	if !isValidLaptop(laptop) {
		return errors.New("invalid laptop : " + truncStr(laptop, 15, "..."))
	}
	return txBaseErr(b, H{"api": userModify, "laptop": laptop})
}

func (b *Bot) setAvatar(avatarID int) error {
	return txBaseErr(b, H{"api": userSetAvatar, "avatarid": avatarID})
}

func (b *Bot) userAvailableAvatars() (out UserAvailableAvatarsRes, err error) {
	b.tx(H{"api": userAvailableAvatars}, &out)
	return out, baseErr(out.BaseRes)
}

func (b *Bot) getAvatarIDs() (out []int, err error) {
	userInfo, err := b.userInfo()
	if err != nil {
		return out, err
	}
	availableAvatars, err := b.userAvailableAvatars()
	if err != nil {
		return out, err
	}
	for _, avatar := range availableAvatars.Avatars {
		if userInfo.Points >= avatar.Min {
			if userInfo.ACL < avatar.ACL {
				continue
			}
			for _, avatarID := range avatar.Avatarids {
				out = append(out, avatarID)
			}
		}
	}
	return out, nil
}

func (b *Bot) updatePresence() error {
	return txBaseErr(b, H{"api": presenceUpdate, "status": b.currentStatus})
}

func (b *Bot) roomRegister(roomID string) error {
	if roomID == "" {
		roomID = b.roomID
	}
	return txBaseErr(b, H{"api": roomRegister, "roomid": roomID})
}

func (b *Bot) userModify(h H) error {
	p := H{"api": userModify}
	for k, v := range h {
		p[k] = v
	}
	return txBaseErr(b, p)
}

func (b *Bot) setStatus(status string) error {
	if !isValidStatus(status) {
		return errors.New("invalid status : " + truncStr(status, 15, "..."))
	}
	b.currentStatus = status
	return b.updatePresence()
}

func (b *Bot) bootUser(userID, reason string) error {
	return txBaseErr(b, H{"api": roomBootUser, "roomid": b.roomID, "target_userid": userID, "reason": reason})
}

func (b *Bot) roomDeregister() error {
	return txBaseErr(b, H{"api": roomDeregister, "roomid": b.roomID})
}

func (b *Bot) playlistCreate(playlistName string) error {
	return txBaseErr(b, H{"api": playlistCreate, "playlist_name": playlistName})
}

func (b *Bot) playlistDelete(playlistName string) error {
	return txBaseErr(b, H{"api": playlistDelete, "playlist_name": playlistName})
}

func (b *Bot) playlistListAll() (out PlaylistListAllRes, err error) {
	b.tx(H{"api": playlistListAll}, &out)
	return out, baseErr(out.BaseRes)
}

func (b *Bot) playlistAll(playlistName string) (out PlaylistAllRes, err error) {
	if playlistName == "" {
		playlistName = "default"
	}
	b.tx(H{"api": playlistAll, "playlist_name": playlistName}, &out)
	return out, baseErr(out.BaseRes)
}

func (b *Bot) playlistAdd(songID, playlistName string, idx int) error {
	if playlistName == "" {
		playlistName = "default"
	}
	if songID == "" {
		songID = b.CurrentSongID
	}
	return txBaseErr(b, H{"api": playlistAdd, "playlist_name": playlistName, "song_dict": H{"fileid": songID}, "index": idx})
}

func (b *Bot) playlistRemove(playlistName string, idx int) error {
	if playlistName == "" {
		playlistName = "default"
	}
	return txBaseErr(b, H{"api": playlistRemove, "playlist_name": playlistName, "index": idx})
}

func (b *Bot) playlistReorder(playlistName string, idxFrom, idxTo int) error {
	if playlistName == "" {
		playlistName = "default"
	}
	return txBaseErr(b, H{"api": playlistReorder, "playlist_name": playlistName, "index_from": idxFrom, "index_to": idxTo})
}

func (b *Bot) playlistSwitch(playlistName string) error {
	if playlistName == "" {
		playlistName = "default"
	}
	return txBaseErr(b, H{"api": playlistSwitch, "playlist_name": playlistName})
}

func (b *Bot) playlistRename(oldPlaylistName, newPlaylistName string) error {
	return txBaseErr(b, H{"api": playlistRename, "old_playlist_name": oldPlaylistName, "new_playlist_name": newPlaylistName})
}

func (b *Bot) addModerator(userID string) error {
	return txBaseErr(b, H{"api": roomAddModerator, "roomid": b.roomID, "target_userid": userID})
}

func (b *Bot) remModerator(userID string) error {
	return txBaseErr(b, H{"api": roomRemModerator, "roomid": b.roomID, "target_userid": userID})
}

func (b *Bot) snag() error {
	sh := Sha1([]byte(GenerateToken()))
	fh := Sha1([]byte(GenerateToken()))
	i := []string{b.userID, b.CurrentDjID, b.CurrentSongID, b.roomID, "queue", "board", "false", "false", sh}
	vh := Sha1([]byte(strings.Join(i, "/")))
	h := H{"api": snagAdd, "roomid": b.roomID, "djid": b.CurrentDjID, "songid": b.CurrentSongID, "site": "queue",
		"location": "board", "in_queue": "false", "blocked": "false", "vh": vh, "sh": sh, "fh": fh}
	return txBaseErr(b, h)
}

func (b *Bot) vote(val string) error {
	if val != up && val != down {
		return errors.New("invalid vote value " + truncStr(val, 10, "..."))
	}
	vh := Sha1([]byte(b.roomID + val + b.CurrentSongID))
	th := Sha1([]byte(GenerateToken()))
	ph := Sha1([]byte(GenerateToken()))
	return txBaseErr(b, H{"api": roomVote, "roomid": b.roomID, "val": val, "vh": vh, "th": th, "ph": ph})
}

func (b *Bot) voteUp() error {
	return b.vote(up)
}

func (b *Bot) bop() error {
	return b.voteUp()
}

func (b *Bot) voteDown() error {
	return b.vote(down)
}

func (b *Bot) addFavorite(roomID string) error {
	if roomID == "" {
		roomID = b.roomID
	}
	return txBaseErr(b, H{"api": roomAddFavorite, "roomid": roomID})
}

func (b *Bot) remFavorite(roomID string) error {
	if roomID == "" {
		roomID = b.roomID
	}
	return txBaseErr(b, H{"api": roomRemFavorite, "roomid": roomID})
}

func (b *Bot) getFavorites() (out GetFavoritesRes, err error) {
	b.tx(H{"api": roomGetFavorites}, &out)
	return out, baseErr(out.BaseRes)
}

func (b *Bot) directoryGraph() (out DirectoryGraphRes, err error) {
	b.tx(H{"api": roomDirectoryGraph}, &out)
	return out, baseErr(out.BaseRes)
}

func (b *Bot) addDj() error {
	return txBaseErr(b, H{"api": roomAddDj, "roomid": b.roomID})
}

func (b *Bot) remDj(userID string) error {
	return txBaseErr(b, H{"api": roomRemDj, "roomid": b.roomID, "djid": userID})
}

func (b *Bot) getPresence(userID string) (out GetPresenceRes, err error) {
	if userID == "" {
		userID = b.userID
	}
	return out, baseErr(out.BaseRes)
}

func (b *Bot) stopSong() error {
	return txBaseErr(b, H{"api": roomStopSong, "roomid": b.roomID})
}

func (b *Bot) skip() error {
	return b.stopSong()
}

func (b *Bot) userInfo() (out UserInfoRes, err error) {
	b.tx(H{"api": userInfo}, &out)
	return out, baseErr(out.BaseRes)
}

func (b *Bot) getFanOf(userID string) (out GetFanOfRes, err error) {
	b.tx(H{"api": userGetFanOf, "userid": userID}, &out)
	return out, baseErr(out.BaseRes)
}

func (b *Bot) getFans() (out GetFansRes, err error) {
	b.tx(H{"api": userGetFans}, &out)
	return out, baseErr(out.BaseRes)
}

func (b *Bot) becomeFan(userID string) error {
	return txBaseErr(b, H{"api": userBecomeFan, "djid": userID})
}

func (b *Bot) removeFan(userID string) error {
	return txBaseErr(b, H{"api": userRemoveFan, "djid": userID})
}

func (b *Bot) roomInfoHashMap() (out H, err error) {
	b.tx(H{"api": roomInfo, "roomid": b.roomID}, &out)
	if !out["success"].(bool) {
		return out, errors.New(out["err"].(string))
	}
	return out, nil
}

func (b *Bot) roomInfoRaw() (out []byte) {
	b.tx(H{"api": roomInfo, "roomid": b.roomID}, &out)
	return out
}

func (b *Bot) roomInfo() (out RoomInfoRes, err error) {
	b.tx(H{"api": roomInfo, "roomid": b.roomID}, &out)
	return out, baseErr(out.BaseRes)
}

func (b *Bot) getUserID(name string) (id string, err error) {
	var res GetUserIDRes
	b.tx(H{"api": userGetID, "name": name}, &res)
	return res.UserID, baseErr(res.BaseRes)
}

func (b *Bot) getProfile(userID string) (out GetProfileRes, err error) {
	b.tx(H{"api": userGetProfileInfo, "profileid": userID}, &out)
	return out, baseErr(out.BaseRes)
}

//-----------------------------------------------------------------------------

// OnReady triggered when the bot is connected
func (b *Bot) OnReady(clb func()) {
	b.addCallback(ready, clb)
}

// OnSpeak triggered when a message is received in the public chat
func (b *Bot) OnSpeak(clb func(SpeakEvt)) {
	b.addCallback(speak, clb)
}

// OnAddDJ triggered when a user takes a dj spot
func (b *Bot) OnAddDJ(clb func(AddDJEvt)) {
	b.addCallback(addDJ, clb)
}

// OnRemDJ triggered when a user leaves a dj spot
func (b *Bot) OnRemDJ(clb func(RemDJEvt)) {
	b.addCallback(remDJ, clb)
}

// OnEscort triggered when a user is escorted off the stage
func (b *Bot) OnEscort(clb func(EscortEvt)) {
	b.addCallback(remDJ, clb)
}

// OnNewModerator triggered when a user is promoted to a moderator
func (b *Bot) OnNewModerator(clb func(NewModeratorEvt)) {
	b.addCallback(newModerator, clb)
}

// OnRemModerator triggered when a user loses their moderator title
func (b *Bot) OnRemModerator(clb func(RemModeratorEvt)) {
	b.addCallback(remModerator, clb)
}

// OnUpdateVotes triggered when a user vote
// Note: the userid is provided only if the user votes up, or later changes their mind and votes down
func (b *Bot) OnUpdateVotes(clb func(UpdateVotesEvt)) {
	b.addCallback(updateVotes, clb)
}

// OnUpdateUser triggered when a user updates their name/profile
func (b *Bot) OnUpdateUser(clb func([]byte)) {
	b.addCallback(updateUser, clb)
}

// OnRegistered triggered when someone enter the room
func (b *Bot) OnRegistered(clb func(RegisteredEvt)) {
	b.addCallback(registered, clb)
}

// OnDeregistered triggered when a user leaves the room
func (b *Bot) OnDeregistered(clb func(DeregisteredEvt)) {
	b.addCallback(deregistered, clb)
}

// OnRoomChanged triggered when the bot enter a room
func (b *Bot) OnRoomChanged(clb func(RoomInfoRes)) {
	b.addCallback(roomChanged, clb)
}

// OnNewSong triggered when a new song starts
func (b *Bot) OnNewSong(clb func(NewSongEvt)) {
	b.addCallback(newsong, clb)
}

// OnEndSong triggered at the end of the song. (Just before the newsong/nosong event)
// The data returned by this event contains information about the song that has just ended.
func (b *Bot) OnEndSong(clb func(H)) {
	b.addCallback(endsong, clb)
}

// OnNoSong triggered when there is no song
func (b *Bot) OnNoSong(clb func(evt NoSongEvt)) {
	b.addCallback(endsong, clb)
}

// OnBootedUse triggered when a user gets booted
func (b *Bot) OnBootedUser(clb func(BootedUserEvt)) {
	b.addCallback(bootedUser, clb)
}

// OnSnagged triggered when a user snag the currently playing song
func (b *Bot) OnSnagged(clb func(SnaggedEvt)) {
	b.addCallback(snagged, clb)
}

// OnPmmed triggered when a private message is received
func (b *Bot) OnPmmed(clb func(PmmedEvt)) {
	b.addCallback(pmmed, clb)
}

// On triggered when "cmd" is received
func (b *Bot) On(cmd string, clb func([]byte)) {
	b.addCallback(cmd, clb)
}

//-----------------------------------------------------------------------------

// RoomRegister register in a room
func (b *Bot) RoomRegister(roomID string) error {
	return b.roomRegister(roomID)
}

// Search for a song
func (b *Bot) Search(query string) (SearchRes, error) {
	return b.search(query)
}

// Speak send a message in the public chat
func (b *Bot) Speak(msg string) error {
	return b.speak(msg)
}

// Speakf alias to Speak with formatted arguments
func (b *Bot) Speakf(format string, args ...interface{}) error {
	return b.speakf(format, args...)
}

// PM sends a private message
func (b *Bot) PM(userID, msg string) error {
	return b.pm(userID, msg)
}

// ModifyName changes your name
func (b *Bot) ModifyName(newName string) error {
	return b.modifyName(newName)
}

// ModifyLaptop set your laptop
func (b *Bot) ModifyLaptop(laptop string) error {
	return b.modifyLaptop(laptop)
}

// SetAvatar set your avatar
func (b *Bot) SetAvatar(avatarID int) error {
	return b.setAvatar(avatarID)
}

// UserAvailableAvatars get all available avatars
func (b *Bot) UserAvailableAvatars() (UserAvailableAvatarsRes, error) {
	return b.userAvailableAvatars()
}

// GetAvatarIds get the avatar ids that you can currently use
func (b *Bot) GetAvatarIDs() ([]int, error) {
	return b.getAvatarIDs()
}

// UserModify ...
func (b *Bot) UserModify(h H) error {
	return b.userModify(h)
}

// SetStatus sets your current status
func (b *Bot) SetStatus(status string) error {
	return b.setStatus(status)
}

// BootUser kick a user out of the room
func (b *Bot) BootUser(userID, reason string) error {
	return b.bootUser(userID, reason)
}

// RoomDeregister exit the current room
func (b *Bot) RoomDeregister() error {
	return b.roomDeregister()
}

// PlaylistCreate creates a new playlist
func (b *Bot) PlaylistCreate(playlistName string) error {
	return b.playlistCreate(playlistName)
}

// PlaylistDelete deletes a playlist
func (b *Bot) PlaylistDelete(playlistName string) error {
	return b.playlistDelete(playlistName)
}

// PlaylistAll list all your playlists
func (b *Bot) PlaylistListAll() (PlaylistListAllRes, error) {
	return b.playlistListAll()
}

// PlaylistAll get all information about a playlist
func (b *Bot) PlaylistAll(playlistName string) (PlaylistAllRes, error) {
	return b.playlistAll(playlistName)
}

// PlaylistAdd adds a song to a playlist
// songID will default to the current song id
// playlistName will default to "default"
// idx will default to 0
func (b *Bot) PlaylistAdd(songID, playlistName string, idx int) error {
	return b.playlistAdd(songID, playlistName, idx)
}

// PlaylistRemove remove a song from a playlist
func (b *Bot) PlaylistRemove(playlistName string, idx int) error {
	return b.playlistRemove(playlistName, idx)
}

// PlaylistReorder reorder a playlist. Take the song at index idxFrom and move it to index idxTo.
func (b *Bot) PlaylistReorder(playlistName string, idxFrom, idxTo int) error {
	return b.playlistReorder(playlistName, idxFrom, idxTo)
}

// PlaylistSwitch switch to another playlist
func (b *Bot) PlaylistSwitch(playlistName string) error {
	return b.playlistSwitch(playlistName)
}

// PlaylistRename rename a playlist
func (b *Bot) PlaylistRename(oldPlaylistName, newPlaylistName string) error {
	return b.playlistRename(oldPlaylistName, newPlaylistName)
}

// AddModerator adds a moderator
func (b *Bot) AddModerator(userID string) error {
	return b.addModerator(userID)
}

// RemModerator remove a moderator
func (b *Bot) RemModerator(userID string) error {
	return b.remModerator(userID)
}

// Snag trigger the heart animation used to show that you've snagged the currently playing song.
//
// Warning
// This function will not add the song into the queue. Use PlaylistAdd to queue the song, and if successful, then use Snag to trigger the animation.
func (b *Bot) Snag() error {
	return b.snag()
}

// VoteUp vote up the ongoing song
func (b *Bot) VoteUp() error {
	return b.voteUp()
}

// Bop alis to VoteUp
func (b *Bot) Bop() error {
	return b.bop()
}

// VoteDown vote down the ongoing song
func (b *Bot) VoteDown() error {
	return b.voteDown()
}

// AddFavorite add a room to your favorite rooms
func (b *Bot) AddFavorite(roomID string) error {
	return b.addFavorite(roomID)
}

// RemFavorite remove a room from your favorite rooms
func (b *Bot) RemFavorite(roomID string) error {
	return b.remFavorite(roomID)
}

// GetFavorites get your favorite rooms
func (b *Bot) GetFavorites() (GetFavoritesRes, error) {
	return b.getFavorites()
}

// DirectoryGraph get the location of your friends/idols
func (b *Bot) DirectoryGraph() (DirectoryGraphRes, error) {
	return b.directoryGraph()
}

// AddDj step up as a DJ
func (b *Bot) AddDj() error {
	return b.addDj()
}

// RemDj remove userID from DJ spot | or yourself if userID is empty
func (b *Bot) RemDj(userID string) error {
	return b.remDj(userID)
}

// GetPresence get presence for the specified user, or your presence if a userID is not specified
func (b *Bot) GetPresence(userID string) (GetPresenceRes, error) {
	return b.getPresence(userID)
}

// StopSong skip the song you are currently playing
func (b *Bot) StopSong() error {
	return b.stopSong()
}

// Skip is an alias to StopSong
func (b *Bot) Skip() error {
	return b.skip()
}

// UserInfo returns the information about the user
func (b *Bot) UserInfo() (UserInfoRes, error) {
	return b.userInfo()
}

// GetFanOf gets the list of everyone the specified userID is a fan of, or the list of everyone you are a fan of if a userID is not specified
func (b *Bot) GetFanOf(userID string) (GetFanOfRes, error) {
	return b.getFanOf(userID)
}

// GetFans get the list of everyone who is a fan of you
func (b *Bot) GetFans() (GetFansRes, error) {
	return b.getFans()
}

// BecomeFan fan someone
func (b *Bot) BecomeFan(userID string) error {
	return b.becomeFan(userID)
}

// RemoveFan unfan someone
func (b *Bot) RemoveFan(userID string) error {
	return b.removeFan(userID)
}

// RoomInfo gets information about the current room
func (b *Bot) RoomInfoHashMap() (H, error) {
	return b.roomInfoHashMap()
}

// RoomInfo gets information about the current room
func (b *Bot) RoomInfo() (RoomInfoRes, error) {
	return b.roomInfo()
}

// GetUserID gets a user's ID by their name
func (b *Bot) GetUserID(name string) (id string, err error) {
	return b.getUserID(name)
}

// GetProfile given a UserID, gets a user profile
func (b *Bot) GetProfile(userID string) (GetProfileRes, error) {
	return b.getProfile(userID)
}
