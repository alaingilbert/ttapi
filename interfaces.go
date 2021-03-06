package ttapi

// IBot ...
type IBot interface {
	On(cmd string, clb func([]byte))
	OnNewSong(func(NewSongEvt))
	OnPmmed(func(PmmedEvt))
	OnRegistered(func(RegisteredEvt))
	OnRoomChanged(func(RoomInfoRes))
	OnSpeak(func(SpeakEvt))

	AddDj() error
	AddFavorite(roomID string) error
	AddModerator(userID string) error
	BecomeFan(userID string) error
	BootUser(userID, reason string) error
	Bop() error
	DirectoryGraph() (DirectoryGraphRes, error)
	GetAvatarIDs() ([]int, error)
	GetFavorites() (GetFavoritesRes, error)
	GetFanOf(userID string) (GetFanOfRes, error)
	GetFans() (GetFansRes, error)
	GetPresence(userID string) (GetPresenceRes, error)
	GetProfile(userID string) (GetProfileRes, error)
	GetUserID(name string) (string, error)
	SetStatus(status string) error
	modifyLaptop(laptop string) error
	ModifyName(newName string) error
	PlaylistAdd(songID, playlistName string, idx int) error
	PlaylistAll(playlistName string) (PlaylistAllRes, error)
	PlaylistCreate(playlistName string) error
	PlaylistDelete(playlistName string) error
	PlaylistListAll() (PlaylistListAllRes, error)
	PlaylistRemove(playlistName string, idx int) error
	PlaylistRename(oldPlaylistName, newPlaylistName string) error
	PlaylistReorder(playlistName string, idxFrom, idxTo int) error
	PlaylistSwitch(playlistName string) error
	PM(userID, msg string) error
	RemDj(userID string) error
	RemModerator(userID string) error
	RemFavorite(roomID string) error
	RemoveFan(userID string) error
	RoomDeregister() error
	RoomInfo() (RoomInfoRes, error)
	RoomRegister(roomID string) error
	SetAvatar(avatarID int) error
	Skip() error
	Snag() error
	Speak(msg string) error
	Speakf(format string, args ...interface{}) error
	Start()
	Stop()
	StopSong() error
	UserAvailableAvatars() (UserAvailableAvatarsRes, error)
	UserInfo() (UserInfoRes, error)
	UserModify(H) error
	VoteDown() error
	VoteUp() error
}

// Compile time checks to ensure type satisfies IBot interface
var _ IBot = &Bot{}
var _ IBot = (*Bot)(nil)
