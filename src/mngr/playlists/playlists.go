package playlistsM

import (
	"sync"
)

type PlaylistManager struct {
	playlist RoomPlaylistState
	mutex    sync.RWMutex
}

type RoomPlaylistState struct {
	Files []interface{}
	Index interface{}
	User  struct {
		Username   string
		connection interface{}
	}

	Users map[string]user
}

type user struct {
	Position int
	Paused   bool
	SetBy    string
	DoSeek   bool

	Duration float64
	Name     string
	Size     float64
}

func NewPlaylistManager() *PlaylistManager {
	return &PlaylistManager{
		playlist: RoomPlaylistState{
			Users: make(map[string]user),
		},
	}
}

func (pm *PlaylistManager) SetPlaylist(playlist RoomPlaylistState) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	pm.playlist = playlist
}

func (pm *PlaylistManager) GetPlaylist() RoomPlaylistState {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	return pm.playlist
}

func (pm *PlaylistManager) SetUserPlaystate(username string, position int, paused bool, doSeek bool, setBy string) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	// check if playstate exists if not create it
	if _, exists := pm.playlist.Users[username]; !exists {
		pm.playlist.Users[username] = user{}
	}

	pm.playlist.Users[username] = user{
		Position: position,
		Paused:   paused,
		DoSeek:   doSeek,
		SetBy:    setBy,
	}
}

// SetUserFile
func (pm *PlaylistManager) SetUserFile(username string, duration float64, name string, size float64) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	pm.playlist.User.Username = username
	pm.playlist.User.connection = nil

	pm.playlist.Users[username] = user{
		Duration: duration,
		Name:     name,
		Size:     size,
	}

}

func (pm *PlaylistManager) GetUserPlaystate(username string) (user, bool) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	state, exists := pm.playlist.Users[username]
	return state, exists
}

// get User object
func (pm *PlaylistManager) GetUserObject(username string) (user, bool) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	state, exists := pm.playlist.Users[username]
	return state, exists
}

func (pm *PlaylistManager) RemoveUserPlaystate(username string) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	delete(pm.playlist.Users, username)
}

// User state list for list message
func (pm *PlaylistManager) GetUserStates() map[string]user {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	return pm.playlist.Users
}
