package roomM

import (
	"fmt"
	"net"
	"sync"

	"github.com/Icey-Glitch/Syncplay-G/mngr/event"
	playlistsM "github.com/Icey-Glitch/Syncplay-G/mngr/playlists"
	"github.com/Icey-Glitch/Syncplay-G/mngr/ready"
)

type Connection struct {
	Username   string
	State      interface{}
	Conn       net.Conn
	RoomName   string
	readyState ready.ReadyState

	LatencyCalculation float64
}

type Room struct {
	Name            string
	Users           []*Connection
	ReadyManager    *ready.ReadyManager
	PlaylistManager *playlistsM.PlaylistManager
	mutex           sync.RWMutex

	RoomState roomState

	stateEvent *event.TimedEvent
}

type roomState struct {
	IsPaused bool    `json:"isPaused"`
	Position float64 `json:"position"`
	SetBy    string
}

func NewRoom(name string) *Room {
	return &Room{
		Name:            name,
		Users:           make([]*Connection, 0),
		ReadyManager:    ready.NewReadyManager(),
		PlaylistManager: playlistsM.NewPlaylistManager(),
		stateEvent:      event.NewTimedEvent(1),
	}
}

func GetRoomByConnection(conn net.Conn, rooms map[string]*Room) *Room {
	for _, room := range rooms {
		for _, connection := range room.Users {
			if connection.Conn == conn {
				return room
			}
		}
	}
	return nil
}

func (r *Room) GetConnectionByUsername(username string) *Connection {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	for _, connection := range r.Users {
		if connection.Username == username {
			return connection
		}
	}
	return nil
}

func (r *Room) GetUsernameByConnection(conn net.Conn) string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	for _, connection := range r.Users {
		if connection.Conn == conn {
			return connection.Username
		}
	}
	return ""
}

func (r *Room) GetConnections() []*Connection {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	return r.Users
}

func (r *Room) AddConnection(connection *Connection) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.Users = append(r.Users, connection)

}

func (r *Room) RemoveConnection(conn net.Conn) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	for i, connection := range r.Users {
		if connection.Conn == conn {
			// delete ready state
			r.ReadyManager.RemoveUserReadyState(connection.Username)
			// delete playstate
			r.PlaylistManager.RemoveUserPlaystate(connection.Username)
			// remove connection
			r.Users = append(r.Users[:i], r.Users[i+1:]...)
			r.stateEvent.Publish(conn)
			break
		}
	}
}

func (r *Room) SubscribeToStateChanges() chan interface{} {
	return r.stateEvent.Subscribe()
}

/*
usage:
ch := room.SubscribeToStateChanges()
for {
	select {
	case state := <-ch:
		// do something with state
	}
}
*/

func (r *Room) UnsubscribeFromStateChanges(ch chan interface{}) {
	r.stateEvent.Unsubscribe(ch)
}

// list rooms
func ListRooms(rooms map[string]*Room) []string {
	roomNames := make([]string, 0)
	for roomName := range rooms {
		roomNames = append(roomNames, roomName)
	}
	return roomNames
}

// ready state
func (r *Room) SetUserReadyState(username string, isReady bool, manuallyInitiated bool) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.ReadyManager.SetUserReadyState(username, isReady, manuallyInitiated)
}

// print all ready states
func (r *Room) PrintReadyStates() {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	for _, connection := range r.Users {
		state, exists := r.ReadyManager.GetUserReadyState(connection.Username)
		if exists {
			fmt.Printf("Username: %s, IsReady: %t, ManuallyInitiated: %t\n", state.Username, state.IsReady, state.ManuallyInitiated)
		}
	}
}

// play state
func (r *Room) SetUserPlaystate(username string, position float64, paused bool, doSeek bool, setBy string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// check if playstate exists if not create it
	if _, exists := r.PlaylistManager.GetUserPlaystate(username); !exists {
		r.PlaylistManager.SetUserPlaystate(username, 0, true, false, "")
	}

	r.PlaylistManager.SetUserPlaystate(username, position, paused, doSeek, setBy)
}

// get play state
func (r *Room) GetUserPlaystate(username string) (interface{}, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	return r.PlaylistManager.GetUserPlaystate(username)
}

// room state
func (r *Room) GetRoomState() roomState {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	return r.RoomState
}

func (r *Room) SetRoomState(state roomState) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.RoomState = state

}

// set user latency calculation
func (r *Room) SetUserLatencyCalculation(username string, latencyCalculation float64) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	for _, connection := range r.Users {
		if connection.Username == username {
			connection.LatencyCalculation = latencyCalculation
		}
	}
}

// get user latency calculation
func (r *Room) GetLatencyCalculation(username string) float64 {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	for _, connection := range r.Users {
		if connection.Username == username {
			return connection.LatencyCalculation
		}
	}
	return 0
}
