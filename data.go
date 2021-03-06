package ttapi

// SpeakEvt struct received when someone speak in the public chat
type SpeakEvt struct {
	Command string
	UserID  string
	Name    string
	Text    string
}

// RegisteredEvt ...
type RegisteredEvt struct {
	Command string `json:"command"`
	Roomid  string `json:"roomid"`
	User    []struct {
		Fanofs  int     `json:"fanofs"`
		Name    string  `json:"name"`
		Created float64 `json:"created"`
		Laptop  string  `json:"laptop"`
		Userid  string  `json:"userid"`
		ACL     float64 `json:"acl"`
		Fans    int     `json:"fans"`
		Points  int     `json:"points"`
		Images  struct {
			Fullfront string `json:"fullfront"`
			Headfront string `json:"headfront"`
		} `json:"images"`
		ID         string  `json:"_id"`
		Avatarid   int     `json:"avatarid"`
		Registered float64 `json:"registered"`
	} `json:"user"`
	Success bool `json:"success"`
}

// BaseRes ...
type BaseRes struct {
	Msgid   int    `json:"msgid"`
	Success bool   `json:"success"`
	Err     string `json:"err"`
}

// UserInfoRes ...
type UserInfoRes struct {
	BaseRes
	Fanofs          int     `json:"fanofs"`
	Name            string  `json:"name"`
	Created         float64 `json:"created"`
	UnverifiedEmail string  `json:"unverified_email"`
	Laptop          string  `json:"laptop"`
	Userid          string  `json:"userid"`
	ACL             float64 `json:"acl"`
	Email           string  `json:"email"`
	Fans            int     `json:"fans"`
	Points          int     `json:"points"`
	Images          struct {
		Fullfront string `json:"fullfront"`
		Headfront string `json:"headfront"`
	} `json:"images"`
	ID            string  `json:"_id"`
	Avatarid      int     `json:"avatarid"`
	Registered    float64 `json:"registered"`
	HasTtPassword bool    `json:"has_tt_password"`
}

// GetFanOfRes ...
type GetFanOfRes struct {
	BaseRes
	FanOf []string `json:"fanof"`
}

// GetUserIDRes ...
type GetUserIDRes struct {
	BaseRes
	UserID string `json:"userid"`
}

// GetProfileRes ...
type GetProfileRes struct {
	BaseRes
	Name       string  `json:"name"`
	Created    float64 `json:"created"`
	Laptop     string  `json:"laptop"`
	Userid     string  `json:"userid"`
	Registered float64 `json:"registered"`
	ACL        float64 `json:"acl"`
	Fans       int     `json:"fans"`
	Points     int     `json:"points"`
	Images     struct {
		Fullfront string `json:"fullfront"`
		Headfront string `json:"headfront"`
	} `json:"images"`
	ID       string `json:"_id"`
	Avatarid int    `json:"avatarid"`
	Fanofs   int    `json:"fanofs"`
}

// PmmedEvt ...
type PmmedEvt struct {
	Text     string  `json:"text"`
	Userid   string  `json:"userid"`
	SenderID string  `json:"senderid"`
	Command  string  `json:"command"`
	Time     float64 `json:"time"`
	Roomobj  struct {
		Chatserver []interface{} `json:"chatserver"`
		Name       string        `json:"name"`
		Created    float64       `json:"created"`
		Shortcut   string        `json:"shortcut"`
		Roomid     string        `json:"roomid"`
		Metadata   struct {
			DjFull               bool          `json:"dj_full"`
			Djs                  []interface{} `json:"djs"`
			ScreenUploadsAllowed bool          `json:"screen_uploads_allowed"`
			CurrentSong          interface{}   `json:"current_song"`
			Privacy              string        `json:"privacy"`
			MaxDjs               int           `json:"max_djs"`
			Downvotes            int           `json:"downvotes"`
			Userid               string        `json:"userid"`
			Listeners            int           `json:"listeners"`
			Featured             bool          `json:"featured"`
			Djcount              int           `json:"djcount"`
			CurrentDj            interface{}   `json:"current_dj"`
			Djthreshold          int           `json:"djthreshold"`
			ModeratorID          []string      `json:"moderator_id"`
			Upvotes              int           `json:"upvotes"`
			MaxSize              int           `json:"max_size"`
			Votelog              []interface{} `json:"votelog"`
		} `json:"metadata"`
	} `json:"roomobj"`
}

// RoomInfoRes ...
type RoomInfoRes struct {
	BaseRes
	Room struct {
		Chatserver []interface{} `json:"chatserver"`
		Name       string        `json:"name"`
		Created    float64       `json:"created"`
		Shortcut   string        `json:"shortcut"`
		Roomid     string        `json:"roomid"`
		Metadata   struct {
			Songlog []struct {
				Source   string  `json:"source"`
				Sourceid string  `json:"sourceid"`
				Created  float64 `json:"created"`
				Djid     string  `json:"djid"`
				Score    float64 `json:"score,omitempty"`
				Djname   string  `json:"djname"`
				ID       string  `json:"_id"`
				Metadata struct {
					Coverart string `json:"coverart"`
					Length   int    `json:"length"`
					Artist   string `json:"artist"`
					Song     string `json:"song"`
				} `json:"metadata"`
			} `json:"songlog"`
			DjFull               bool     `json:"dj_full"`
			Djs                  []string `json:"djs"`
			ScreenUploadsAllowed bool     `json:"screen_uploads_allowed"`
			CurrentSong          struct {
				Playlist  string  `json:"playlist"`
				Created   float64 `json:"created"`
				Sourceid  string  `json:"sourceid"`
				Source    string  `json:"source"`
				Djname    string  `json:"djname"`
				Starttime float64 `json:"starttime"`
				ID        string  `json:"_id"`
				Djid      string  `json:"djid"`
				Metadata  struct {
					Coverart string `json:"coverart"`
					Length   int    `json:"length"`
					Artist   string `json:"artist"`
					Song     string `json:"song"`
				} `json:"metadata"`
			} `json:"current_song"`
			Privacy   string `json:"privacy"`
			MaxDjs    int    `json:"max_djs"`
			Downvotes int    `json:"downvotes"`
			Creator   struct {
				Fanofs  int     `json:"fanofs"`
				Name    string  `json:"name"`
				Created float64 `json:"created"`
				Laptop  string  `json:"laptop"`
				Userid  string  `json:"userid"`
				ACL     float64 `json:"acl"`
				Fans    int     `json:"fans"`
				Points  int     `json:"points"`
				Images  struct {
					Fullfront string `json:"fullfront"`
					Headfront string `json:"headfront"`
				} `json:"images"`
				ID         string  `json:"_id"`
				Avatarid   int     `json:"avatarid"`
				Registered float64 `json:"registered"`
			} `json:"creator"`
			Userid            string `json:"userid"`
			Listeners         int    `json:"listeners"`
			StickerPlacements struct {
				Six04130Fb3F4Bfc001809D428 []interface{} `json:"604130fb3f4bfc001809d428"`
			} `json:"sticker_placements"`
			Screens struct {
				Curtain interface{} `json:"curtain"`
				Right   interface{} `json:"right"`
				Left    interface{} `json:"left"`
			} `json:"screens"`
			Featured    bool       `json:"featured"`
			Djcount     int        `json:"djcount"`
			CurrentDj   string     `json:"current_dj"`
			Djthreshold int        `json:"djthreshold"`
			ModeratorID []string   `json:"moderator_id"`
			Upvotes     int        `json:"upvotes"`
			MaxSize     int        `json:"max_size"`
			Votelog     [][]string `json:"votelog"`
		} `json:"metadata"`
	} `json:"room"`
	Djids       []string `json:"djids"`
	Listenerids []string `json:"listenerids"`
	Now         float64  `json:"now"`
	Users       []struct {
		Fanofs  int     `json:"fanofs"`
		Name    string  `json:"name"`
		Created float64 `json:"created"`
		Laptop  string  `json:"laptop"`
		Userid  string  `json:"userid"`
		ACL     float64 `json:"acl"`
		Fans    int     `json:"fans"`
		Points  int     `json:"points"`
		Images  struct {
			Fullfront string `json:"fullfront"`
			Headfront string `json:"headfront"`
		} `json:"images"`
		ID         string  `json:"_id"`
		Avatarid   int     `json:"avatarid"`
		Registered float64 `json:"registered"`
	} `json:"users"`
}

// NewSongEvt ...
type NewSongEvt struct {
	Command string  `json:"command"`
	Now     float64 `json:"now"`
	Roomid  string  `json:"roomid"`
	Room    struct {
		Chatserver []interface{} `json:"chatserver"`
		Name       string        `json:"name"`
		Created    float64       `json:"created"`
		Shortcut   string        `json:"shortcut"`
		Roomid     string        `json:"roomid"`
		Metadata   struct {
			Songlog []struct {
				Source   string  `json:"source"`
				Sourceid string  `json:"sourceid"`
				Created  float64 `json:"created"`
				Djid     string  `json:"djid"`
				Score    float64 `json:"score,omitempty"`
				Djname   string  `json:"djname"`
				ID       string  `json:"_id"`
				Metadata struct {
					Coverart string `json:"coverart"`
					Length   int    `json:"length"`
					Artist   string `json:"artist"`
					Song     string `json:"song"`
				} `json:"metadata"`
			} `json:"songlog"`
			DjFull               bool     `json:"dj_full"`
			Djs                  []string `json:"djs"`
			ScreenUploadsAllowed bool     `json:"screen_uploads_allowed"`
			CurrentSong          struct {
				Playlist  string  `json:"playlist"`
				Created   float64 `json:"created"`
				Sourceid  string  `json:"sourceid"`
				Source    string  `json:"source"`
				Djname    string  `json:"djname"`
				Starttime float64 `json:"starttime"`
				ID        string  `json:"_id"`
				Djid      string  `json:"djid"`
				Metadata  struct {
					Coverart string `json:"coverart"`
					Length   int    `json:"length"`
					Song     string `json:"song"`
					Artist   string `json:"artist"`
				} `json:"metadata"`
			} `json:"current_song"`
			Privacy     string        `json:"privacy"`
			MaxDjs      int           `json:"max_djs"`
			Downvotes   int           `json:"downvotes"`
			Userid      string        `json:"userid"`
			Listeners   int           `json:"listeners"`
			Featured    bool          `json:"featured"`
			Djcount     int           `json:"djcount"`
			CurrentDj   string        `json:"current_dj"`
			Djthreshold int           `json:"djthreshold"`
			ModeratorID []string      `json:"moderator_id"`
			Upvotes     int           `json:"upvotes"`
			MaxSize     int           `json:"max_size"`
			Votelog     []interface{} `json:"votelog"`
		} `json:"metadata"`
	} `json:"room"`
	Success bool `json:"success"`
}

// UserAvailableAvatarsRes ...
type UserAvailableAvatarsRes struct {
	BaseRes
	Avatars []struct {
		Avatarids []int   `json:"avatarids"`
		Min       int     `json:"min"`
		ACL       float64 `json:"acl,omitempty"`
		Pro       []struct {
			Colwidth  int    `json:"colwidth"`
			Name      string `json:"name"`
			Avatarids []int  `json:"avatarids"`
		} `json:"pro,omitempty"`
	} `json:"avatars"`
}

// GetPresenceRes ...
type GetPresenceRes struct {
	BaseRes
	Presence struct {
		Status string `json:"status"`
		UserID string `json:"userid"`
	} `json:"presence"`
}

// PlaylistListAllRes ...
type PlaylistListAllRes struct {
	BaseRes
	List []struct {
		Active bool   `json:"active"`
		Name   string `json:"name"`
	} `json:"list"`
}

// PlaylistAllRes ...
type PlaylistAllRes struct {
	BaseRes
	List []struct {
		Sourceid string `json:"sourceid"`
		Source   string `json:"source"`
		ID       string `json:"_id"`
		Metadata struct {
			Coverart string `json:"coverart"`
			Length   int    `json:"length"`
			Song     string `json:"song"`
			Artist   string `json:"artist"`
		} `json:"metadata"`
		Created float64 `json:"created"`
	} `json:"list"`
}

// GetFavoritesRes ...
type GetFavoritesRes struct {
	BaseRes
	List []string `json:"list"`
}

/**
{
    "msgid": 5,
    "rooms": [
        [
            {
                "chatserver": ["chat1.turntable.fm", 8080],
                "description": "||HOUSE||ELECTRO||DnB||DUBSTEP||TECHNO||MASHUPS||TRANCE||",
                "created": 1614831836.871862,
                "shortcut": "the_party_bus",
                "name": "The Party Bus!",
                "roomid": "604060dc3f4bfc001be4c459",
                "metadata": {
                    "dj_full": false,
                    "djs": ["6040509f3f4bfc001be4c056", "604127333f4bfc0018163042", "604086243f4bfc001be4ccaf"],
                    "screen_uploads_allowed": true,
                    "current_song": {
                        "playlist": "default",
                        "created": 1614898766.480138,
                        "sourceid": "zMPyTjm0lzQ",
                        "source": "yt",
                        "djname": "aQuanaut",
                        "starttime": 1614953381.782457,
                        "_id": "6041664ec2dbd9001be749a1",
                        "djid": "604127333f4bfc0018163042",
                        "metadata": { "coverart": "https://i.ytimg.com/vi/zMPyTjm0lzQ/hqdefault.jpg", "length": 315, "artist": "Icarus", "song": "Fade Away" }
                    },
                    "privacy": "public",
                    "max_djs": 5,
                    "downvotes": 0,
                    "userid": "6040509f3f4bfc001be4c056",
                    "listeners": 8,
                    "featured": false,
                    "djcount": 3,
                    "current_dj": "604127333f4bfc0018163042",
                    "djthreshold": 0,
                    "moderator_id": ["6040509f3f4bfc001be4c056", "604061bd3f4bfc001be4c4aa", "604127333f4bfc0018163042", "604062e23f4bfc001be4c507"],
                    "upvotes": 2,
                    "max_size": 200,
                    "votelog": []
                }
            },
            [
                {
                    "status": "away",
                    "fanofs": 3,
                    "name": "agilbert",
                    "created": 1614885115.13613,
                    "laptop": "mac",
                    "laptop_version": null,
                    "userid": "604130fb3f4bfc001809d428",
                    "acl": 0,
                    "fans": 2,
                    "points": 76,
                    "images": { "fullfront": "/roommanager_assets/avatars/8/fullfront.png", "headfront": "/roommanager_assets/avatars/8/headfront.png" },
                    "_id": "604130fb3f4bfc001809d428",
                    "avatarid": 8,
                    "registered": 1614885141.040958
                }
            ]
        ]
    ],
    "success": true
}
*/
// DirectoryGraphRes ...
type DirectoryGraphRes struct {
	BaseRes
	Rooms [][]interface{} `json:"rooms"`
}

// GetFansRes ...
type GetFansRes struct {
	BaseRes
	Fans []string `json:"fans"`
}