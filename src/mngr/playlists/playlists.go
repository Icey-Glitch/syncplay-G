package playlists

import (
	"fmt"
	"math"
	"sync"

	Features "github.com/Icey-Glitch/Syncplay-G/features"
	"github.com/Icey-Glitch/Syncplay-G/mngr/event"
)

type User struct {
	Username string
	Position float64
	Paused   bool
	DoSeek   bool

	LastMessageAge float64

	File        *File
	UsrPlaylist []File
}

type Playlist struct {
	Files []File
	Index interface{}

	SetBy        string
	Paused       bool
	DoSeek       bool
	Position     float64
	PositionTime float64
	Ignore       float64

	User struct {
		Username   string
		connection interface{}
	}

	Users map[string]User

	doSeekTime float64
}

type File struct {
	Size     float64
	Name     string
	Duration float64
}

type PlaylistManager struct {
	Playlist   Playlist
	mutex      sync.RWMutex
	stateEvent *event.Event
}

func NewPlaylistManager() *PlaylistManager {
	return &PlaylistManager{
		Playlist:   Playlist{Users: make(map[string]User), Paused: true, DoSeek: false, PositionTime: 0},
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

		LastMessageAge: 0,

		File: nil,
	}

	pm.stateEvent.Publish(pm.Playlist.Users[username])
	return nil
}

func (pm *PlaylistManager) SetUserPlaystate(username string, position float64, paused bool, doSeek bool, setBy string, messageAge float64, Ignore bool) error {
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	if !Ignore {
		// check if the possiton is within acceptable range else update the position and set by to inform other users if there is a desync
		err := pm.SetUsersPosition(position, messageAge)
		if err != nil {
			return err
		}

		if math.Abs(pm.Playlist.Position-position) > Features.GlobalConfig.DesyncRange {
			pm.Playlist.Position = position
			pm.Playlist.PositionTime = messageAge
			pm.Playlist.SetBy = setBy
		}

		if doSeek != pm.Playlist.DoSeek {
			err = pm.SetUsersDoSeek(doSeek, messageAge)
			if err != nil {
				return err
			}
		}

		if paused != pm.Playlist.Paused {
			pm.SetUsersPaused(paused)

			err = pm.SetUsersPosition(position, messageAge)
			if err != nil {
				return err
			}
		}
	} else {
		pm.Playlist.Paused = paused
		pm.Playlist.DoSeek = doSeek
		pm.Playlist.Position = position
		pm.Playlist.PositionTime = messageAge
		pm.Playlist.SetBy = setBy

	}

	// TODO: update room paused state if one user unpauses or pause all users if one user pauses
	// check if user exists
	_, exists := pm.Playlist.Users[username]
	if !exists {
		return fmt.Errorf("user %s does not exist in the playlist", username)
	}

	pm.Playlist.Users[username] = User{
		Username: username,
		Position: position,
		Paused:   paused,
		DoSeek:   doSeek,
	}

	pm.Playlist.SetBy = setBy

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

// AddFile adds a file to the playlist
func (pm *PlaylistManager) AddFile(duration float64, name string, size float64, User string) (File, error) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	// check if file already exists
	for _, file := range pm.Playlist.Files {
		if file.Name == name {
			return file, nil
		}
	}

	// check if shared playlist is enabled
	if Features.GlobalFeatures.SharedPlaylists {
		pm.Playlist.Files = append(pm.Playlist.Files, File{
			Size:     size,
			Name:     name,
			Duration: duration,
		})

	} else {
		// add to the user's playlist in their user object
		user, exists := pm.Playlist.Users[User]
		if !exists {
			return File{}, fmt.Errorf("user %s does not exist in the playlist", User)
		}

		user.UsrPlaylist = append(user.UsrPlaylist, File{
			Size:     size,
			Name:     name,
			Duration: duration,
		})
		pm.Playlist.Users[User] = user

		return user.UsrPlaylist[len(user.UsrPlaylist)-1], nil

	}

	// return the file, and nil error
	// check if the Files array is empty
	if len(pm.Playlist.Files) == 0 {
		return File{}, fmt.Errorf("files array is empty")
	}
	return pm.Playlist.Files[len(pm.Playlist.Files)-1], nil
}

// SetLastMessageAge sets the last message age for the user
func (pm *PlaylistManager) SetLastMessageAge(username string, age float64) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	user := pm.Playlist.Users[username]
	user.LastMessageAge = age
	pm.Playlist.Users[username] = user
}

// GetLastMessageAge gets the last message age for the user
func (pm *PlaylistManager) GetLastMessageAge(username string) float64 {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	return pm.Playlist.Users[username].LastMessageAge
}

// SetIgnoreInt
func (pm *PlaylistManager) SetIgnoreInt(ignoreInt float64) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	pm.Playlist.Ignore = ignoreInt
}

// helper func
func (pm *PlaylistManager) AddFiles(files []File, User string) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	// check if shared playlist is enabled
	if Features.GlobalFeatures.SharedPlaylists {
		// if file doesnt exist in new array then remove it from the playlist, and add new files
		for i := 0; i < len(pm.Playlist.Files); i++ {
			found := false
			for _, file := range files {
				if file.Name == pm.Playlist.Files[i].Name {
					found = true
					break
				}
			}

			if !found {
				pm.Playlist.Files = append(pm.Playlist.Files[:i], pm.Playlist.Files[i+1:]...)
				i--
			}

		}

		for _, file := range files {
			pm.Playlist.Files = append(pm.Playlist.Files, file)
		}

	} else {
		// Retrieve the User struct from the map
		userPlaylist := pm.Playlist.Users[User]

		// if file doesnt exist in new array then remove it from the playlist
		for i := 0; i < len(userPlaylist.UsrPlaylist); i++ {
			found := false
			for _, file := range files {
				if file.Name == userPlaylist.UsrPlaylist[i].Name {
					found = true
					break
				}
			}

			if !found {
				// Remove file from user's playlist
				userPlaylist.UsrPlaylist = append(userPlaylist.UsrPlaylist[:i], userPlaylist.UsrPlaylist[i+1:]...)
				i--
			}
		}

		// Add new files to the user's playlist
		for _, file := range files {
			userPlaylist.UsrPlaylist = append(userPlaylist.UsrPlaylist, file)
		}

		// Store the updated User struct back in the map
		pm.Playlist.Users[User] = userPlaylist
	}
}

// SetUserFile sets the file for the user
func (pm *PlaylistManager) SetUserFile(username string, file File) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	user, exists := pm.Playlist.Users[username]
	if !exists {
		return fmt.Errorf("user %s does not exist in the playlist", username)
	}

	user.File = &file
	pm.Playlist.Users[username] = user
	return nil
}

// CalculatePosition calculates the position of the playlist and the delta time between the last message and the current message
func (pm *PlaylistManager) CalculatePosition(messageAge float64) (position float64, delta float64) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	if pm.Playlist.Paused {
		return pm.Playlist.Position, 0
	}

	Delta := messageAge - pm.Playlist.PositionTime
	Position := pm.Playlist.Position + Delta
	return Position, Delta
}

func (pm *PlaylistManager) GetUserPlaystate(username string) (User, bool) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	state, exists := pm.Playlist.Users[username]
	// if paused, set position to the room position
	if pm.Playlist.Paused {
		state.Position = pm.Playlist.Position
	}
	return state, exists
}

// SetUsersDoSeek sets all users in the playlist to doSeek
func (pm *PlaylistManager) SetUsersDoSeek(doSeek bool, age float64) error {
	if age > pm.Playlist.doSeekTime { // only update if the new age is greater
		pm.Playlist.doSeekTime = age
		pm.Playlist.Paused = true
		pm.Playlist.DoSeek = doSeek
	}
	return nil
}

// SetUsersPaused sets all users in the playlist to paused
func (pm *PlaylistManager) SetUsersPaused(paused bool) {
	pm.Playlist.Paused = paused
}

// SetUsersPosition sets the position of all users in the playlist
func (pm *PlaylistManager) SetUsersPosition(position float64, age float64) error {
	if age > pm.Playlist.PositionTime { // only update if the new age is greater
		pm.Playlist.PositionTime = age
		pm.Playlist.Position = position
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
