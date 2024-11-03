package Features

type Features struct {
	IsolateRooms         bool `json:"isolateRooms"`
	Readiness            bool `json:"readiness"`
	ManagedRooms         bool `json:"managedRooms"`
	PersistentRooms      bool `json:"persistentRooms"`
	Chat                 bool `json:"chat"`
	SharedPlaylists      bool `json:"sharedPlaylists"`
	MaxChatMessageLength int  `json:"maxChatMessageLength"`
	MaxUsernameLength    int  `json:"maxUsernameLength"`
	MaxRoomNameLength    int  `json:"maxRoomNameLength"`
	MaxFilenameLength    int  `json:"maxFilenameLength"`
}

// GlobalFeatures is a global variable that holds the features of the server
var GlobalFeatures Features

// SetGlobalFeatures sets the global features of the server
func SetGlobalFeatures(features Features) {
	GlobalFeatures = features
}

// GetGlobalFeatures returns the global features of the server
func GetGlobalFeatures() Features {
	return GlobalFeatures
}

// NewFeatures returns a new Features struct
func NewFeatures() *Features {
	return &Features{
		IsolateRooms:         true,
		Readiness:            true,
		ManagedRooms:         false,
		PersistentRooms:      false,
		Chat:                 true,
		SharedPlaylists:      true,
		MaxChatMessageLength: 1000,
		MaxUsernameLength:    20,
		MaxRoomNameLength:    20,
		MaxFilenameLength:    50,
	}
}
