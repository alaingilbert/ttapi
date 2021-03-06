package ttapi


const (
	wsOrigin   = "https://turntable.fm"
	wssURL     = "wss://chat1.turntable.fm:8080/socket.io/websocket"
	wsProtocol = ""

	endsong     = "endsong"
	newsong     = "newsong"
	nosong      = "nosong"
	pmmed       = "pmmed"
	registered  = "registered"
	roomChanged = "roomChanged"
	speak       = "speak"

	pmSend               = "pm.send"
	presenceUpdate       = "presence.update"
	playlistAdd          = "playlist.add"
	playlistAll          = "playlist.all"
	playlistCreate       = "playlist.create"
	playlistDelete       = "playlist.delete"
	playlistListAll      = "playlist.list_all"
	playlistRemove       = "playlist.remove"
	playlistRename       = "playlist.rename"
	playlistReorder      = "playlist.reorder"
	playlistSwitch       = "playlist.switch"
	presenceGet          = "presence.get"
	roomAddDj            = "room.add_dj"
	roomAddFavorite      = "room.add_favorite"
	roomAddModerator     = "room.add_moderator"
	roomDirectoryGraph   = "room.directory_graph"
	roomGetFavorites     = "room.get_favorites"
	roomRemFavorite      = "room.rem_favorite"
	roomRemModerator     = "room.rem_moderator"
	roomBootUser         = "room.boot_user"
	roomDeregister       = "room.deregister"
	roomInfo             = "room.info"
	roomRegister         = "room.register"
	roomRemDj            = "room.rem_dj"
	roomSpeak            = "room.speak"
	roomStopSong         = "room.stop_song"
	roomVote             = "room.vote"
	snagAdd              = "snag.add"
	userAvailableAvatars = "user.available_avatars"
	userBecomeFan        = "user.become_fan"
	userGetFanOf         = "user.get_fan_of"
	userGetFans          = "user.get_fans"
	userGetID            = "user.get_id"
	userGetProfileInfo   = "user.get_profile_info"
	userInfo             = "user.info"
	userModify           = "user.modify"
	userRemoveFan        = "user.remove_fan"
	userSetAvatar        = "user.set_avatar"

	// Valid statuses
	available   = "available"
	unavailable = "unavailable"
	away        = "away"

	// Valid laptops
	androidLaptop = "android"
	chromeLaptop  = "chrome"
	iphoneLaptop  = "iphone"
	linuxLaptop   = "linux"
	macLaptop     = "mac"
	pcLaptop      = "pc"

	// Valid clients
	webClient = "web"

	// Valid vote values
	down = "down"
	up   = "up"
)