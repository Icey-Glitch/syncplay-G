package playlists

import (
	"fmt"
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

	LastMessageAge float64
}

type Playlist struct {
	Files  []interface{}
	Index  interface{}
	Paused bool
	DoSeek bool
	User   struct {
		Username   string
		connection interface{}
	}

	Users map[string]User

	doSeekTime float64
}

type PlaylistManager struct {
	Playlist   Playlist
	mutex      sync.RWMutex
	stateEvent *event.Event
}

func NewPlaylistManager() *PlaylistManager {
	return &PlaylistManager{
		Playlist:   Playlist{Users: make(map[string]User), Paused: true, DoSeek: true},
		stateEvent: event.NewEvent(),
	}
}

// CreateUserPlaystate creates a new user in the playlist
func (pm *PlaylistManager) CreateUserPlaystate(username string) error {
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	_, exists := pm.Playlist.Users[username]
	if exists {
		return fmt.Errorf("user %s already exists in the playlist", username)
	}

	pm.Playlist.Users[username] = User{
		Username: username,
		Position: 0,
		Paused:   true,
		DoSeek:   false,
	}

	pm.stateEvent.Publish(pm.Playlist.Users[username])
	return nil
}

func (pm *PlaylistManager) SetUserPlaystate(username string, position float64, paused bool, doSeek bool, setBy string, messageAge float64) error {
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	user, exists := pm.Playlist.Users[username]
	if !exists {
		return fmt.Errorf("user %s does not exist in the playlist", username)
	}

	if doSeek != user.DoSeek {
		err := pm.SetUsersDoSeek(doSeek, messageAge)
		if err != nil {
			return err
		}
	}

	if paused != pm.Playlist.Paused {
		pm.SetUsersPaused(paused)

		pm.SetUsersPosition(position, messageAge)
	}

	// TODO: update room paused state if one user unpauses or pause all users if one user pauses

	pm.Playlist.Users[username] = User{
		Username: username,
		Position: position,
		Paused:   paused,
		DoSeek:   doSeek,
		SetBy:    setBy,
	}

	pm.stateEvent.Publish(pm.Playlist.Users[username])
	return nil
}

// RemoveUserPlaystate removes the user from the playlist
func (pm *PlaylistManager) RemoveUserPlaystate(username string) error {
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	_, exists := pm.Playlist.Users[username]
	if !exists {
		return fmt.Errorf("user %s does not exist in the playlist", username)
	}

	delete(pm.Playlist.Users, username)
	pm.stateEvent.Publish(username)
	return nil
}

func (pm *PlaylistManager) SetUserFile(username string, duration float64, name string, size float64) error {
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	user, exists := pm.Playlist.Users[username]
	if !exists {
		return fmt.Errorf("user %s does not exist in the playlist", username)
	}

	user.Duration = duration
	user.Name = name
	user.Size = size

	pm.Playlist.Users[username] = user
	pm.stateEvent.Publish(user)
	return nil
}

func (pm *PlaylistManager) GetUserPlaystate(username string) (User, bool) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	state, exists := pm.Playlist.Users[username]
	return state, exists
}

// SetUsersDoSeek sets all users in the playlist to doSeek
func (pm *PlaylistManager) SetUsersDoSeek(doSeek bool, age float64) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	if age > pm.Playlist.doSeekTime { // only update if the new age is greater
		pm.Playlist.doSeekTime = age
		pm.Playlist.Paused = true
		pm.Playlist.DoSeek = doSeek
	}
	return nil
}

// SetUsersPaused sets all users in the playlist to paused
func (pm *PlaylistManager) SetUsersPaused(paused bool) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	pm.Playlist.Paused = paused
}

// SetUsersPosition sets the position of all users in the playlist
func (pm *PlaylistManager) SetUsersPosition(position float64, age float64) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	for username := range pm.Playlist.Users {
		user := pm.Playlist.Users[username]
		if age > user.LastMessageAge { // only update if the new age is greater
			user.Position = position
			pm.Playlist.Users[username] = user
		}
	}
	return nil
}

// GetUserPauseState returns the pause state of the playlist
func (pm *PlaylistManager) GetUserPauseState() bool {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	return pm.Playlist.Paused
}

// GetUsers returns a list of users in the playlist
func (pm *PlaylistManager) GetUsers() (map[string]User, bool) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	return pm.Playlist.Users, len(pm.Playlist.Users) > 0

}

// SetLastMessageAge sets the last message age for the user
func (pm *PlaylistManager) SetLastMessageAge(username string, age float64) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	user := pm.Playlist.Users[username]
	user.LastMessageAge = age
	pm.Playlist.Users[username] = user
}

func (pm *PlaylistManager) GetPlaylist() Playlist {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	return pm.Playlist
}

func (pm *PlaylistManager) SetPlaylist(playlist Playlist) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	pm.Playlist = playlist
	pm.stateEvent.Publish(pm.Playlist)
}

func (pm *PlaylistManager) GetUserObject(username string) (User, bool) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	user, exists := pm.Playlist.Users[username]
	return user, exists
}

func (pm *PlaylistManager) SubscribeToStateChanges() chan interface{} {
	return pm.stateEvent.Subscribe()
}

func (pm *PlaylistManager) UnsubscribeFromStateChanges(ch chan interface{}) {
	pm.stateEvent.Unsubscribe(ch)
}
