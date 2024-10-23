package playlists

import (
	"sync"

	"github.com/Icey-Glitch/Syncplay-G/mngr/event"
)

type User struct {
	Username string
	Position float64
	Paused   bool
	DoSeek   bool
	SetBy    string
	Duration float64
	Name     string
	Size     float64
}

type Playlist struct {
	Files []interface{}
	Index interface{}
	User  struct {
		Username   string
		connection interface{}
	}

	Users map[string]User
}

type PlaylistManager struct {
	playlist   Playlist
	mutex      sync.RWMutex
	stateEvent *event.Event
}

func NewPlaylistManager() *PlaylistManager {
	return &PlaylistManager{
		playlist:   Playlist{Users: make(map[string]User)},
		stateEvent: event.NewEvent(),
	}
}

func (pm *PlaylistManager) SetUserPlaystate(username string, position float64, paused bool, doSeek bool, setBy string) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	pm.playlist.Users[username] = User{
		Username: username,
		Position: position,
		Paused:   paused,
		DoSeek:   doSeek,
		SetBy:    setBy,
	}

	pm.stateEvent.Publish(pm.playlist.Users[username])
}

// RemoveUserPlaystate removes the user from the playlist
func (pm *PlaylistManager) RemoveUserPlaystate(username string) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	delete(pm.playlist.Users, username)
	pm.stateEvent.Publish(username)
}

func (pm *PlaylistManager) SetUserFile(username string, duration float64, name string, size float64) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	user := pm.playlist.Users[username]
	user.Duration = duration
	user.Name = name
	user.Size = size
	pm.playlist.Users[username] = user

	pm.stateEvent.Publish(pm.playlist.Users[username])
}

func (pm *PlaylistManager) GetUserPlaystate(username string) (User, bool) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	state, exists := pm.playlist.Users[username]
	return state, exists
}

// GetUsers returns a map of all users in the playlist
func (pm *PlaylistManager) GetUsers() (map[string]User, bool) {
	if len(pm.playlist.Users) == 0 {
		return nil, false
	}

	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	return pm.playlist.Users, true

}

func (pm *PlaylistManager) GetPlaylist() Playlist {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	return pm.playlist
}

func (pm *PlaylistManager) SetPlaylist(playlist Playlist) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	pm.playlist = playlist
	pm.stateEvent.Publish(pm.playlist)
}

func (pm *PlaylistManager) GetUserObject(username string) (User, bool) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	user, exists := pm.playlist.Users[username]
	return user, exists
}

func (pm *PlaylistManager) SubscribeToStateChanges() chan interface{} {
	return pm.stateEvent.Subscribe()
}

func (pm *PlaylistManager) UnsubscribeFromStateChanges(ch chan interface{}) {
	pm.stateEvent.Unsubscribe(ch)
}
