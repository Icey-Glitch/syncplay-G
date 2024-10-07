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
	position int
	paused   bool
	setBy    string
	doSeek   bool
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
		position: position,
		paused:   paused,
		doSeek:   doSeek,
		setBy:    setBy,
	}
}

func (pm *PlaylistManager) GetUserPlaystate(username string) (user, bool) {
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
