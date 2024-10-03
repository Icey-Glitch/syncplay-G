package messages

type Features struct {
	IsolateRooms         bool `json:"isolateRooms"`
	Readiness            bool `json:"readiness"`
	ManagedRooms         bool `json:"managedRooms"`
	PersistentRooms      bool `json:"persistentRooms"`
	Chat                 bool `json:"chat"`
	MaxChatMessageLength int  `json:"maxChatMessageLength"`
	MaxUsernameLength    int  `json:"maxUsernameLength"`
	MaxRoomNameLength    int  `json:"maxRoomNameLength"`
	MaxFilenameLength    int  `json:"maxFilenameLength"`
}
