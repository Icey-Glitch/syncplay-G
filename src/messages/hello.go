package messages

type HelloResponseMessage struct {
	Hello struct {
		Username string `json:"username"`
		Room     struct {
			Name string `json:"name"`
		} `json:"room"`
		Version     string   `json:"version"`
		RealVersion string   `json:"realversion"`
		Features    Features `json:"features"`
		MOTD        string   `json:"motd"`
	} `json:"Hello"`
}

func CreateHelloResponse(username, version, roomName string) HelloResponseMessage {
	return HelloResponseMessage{
		Hello: struct {
			Username string `json:"username"`
			Room     struct {
				Name string `json:"name"`
			} `json:"room"`
			Version     string   `json:"version"`
			RealVersion string   `json:"realversion"`
			Features    Features `json:"features"`
			MOTD        string   `json:"motd"`
		}{
			Username: username,
			Room: struct {
				Name string `json:"name"`
			}{
				Name: roomName,
			},
			Version:     version,
			RealVersion: "1.7.3",
			Features: Features{
				IsolateRooms:         false,
				Readiness:            true,
				ManagedRooms:         true,
				PersistentRooms:      false,
				Chat:                 true,
				MaxChatMessageLength: 150,
				MaxUsernameLength:    16,
				MaxRoomNameLength:    35,
				MaxFilenameLength:    250,
			},
			MOTD: "",
		},
	}
}
