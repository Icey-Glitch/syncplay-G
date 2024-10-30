package roomM

import (
	"fmt"
	"net"
	"sync"

	"github.com/Icey-Glitch/Syncplay-G/mngr/event"
	playlistsM "github.com/Icey-Glitch/Syncplay-G/mngr/playlists"
	"github.com/Icey-Glitch/Syncplay-G/mngr/ready"
)

/*
TODO:
- Clean up orphin connections and rooms
*/

type Connection struct {
	Username   string
	State      interface{}
	Conn       net.Conn
	RoomName   string
	readyState ready.ReadyState

	// client latency calculation struct
	ClientLatencyCalculation *ClientLatencyCalculation

	StateEvent *event.ManagedEvent
}

type ClientLatencyCalculation struct {
	ArivalTime               float64
	ClientTime               float64
	ClientRtt                float64
	clientLatencyCalculation float64
}

type Room struct {
	Name            string
	Users           []*Connection
	ReadyManager    *ready.ReadyManager
	PlaylistManager *playlistsM.PlaylistManager
	mutex           sync.RWMutex

	RoomState roomState

	stateEventManager *event.EventManager
	stateEventTicker  *event.Ticker
}

type roomState struct {
	IsPaused bool    `json:"isPaused"`
	Position float64 `json:"position"`
	SetBy    string
}

func NewRoom(name string) *Room {
	return &Room{
		Name:              name,
		Users:             make([]*Connection, 0),
		ReadyManager:      ready.NewReadyManager(),
		PlaylistManager:   playlistsM.NewPlaylistManager(),
		stateEventManager: event.NewEventManager(),
		stateEventTicker:  event.NewTicker(1, true),
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
	//r.mutex.RLock()
	//defer r.mutex.RUnlock()

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

	// add playstate
	err := r.PlaylistManager.CreateUserPlaystate(connection.Username)
	if err != nil {
		fmt.Println("Failed to create user playstate " + err.Error())
		return
	}
}

func (r *Room) RemoveConnection(conn net.Conn) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	for i, connection := range r.Users {
		if connection.Conn == conn {

			// delete ready state
			r.ReadyManager.RemoveUserReadyState(connection.Username)
			// delete playstate
			err := r.PlaylistManager.RemoveUserPlaystate(connection.Username)
			if err != nil {
				fmt.Println("failed to remove UserPlaystate " + err.Error())
				return
			}
			// remove connection
			r.Users = append(r.Users[:i], r.Users[i+1:]...)

			if connection.StateEvent != nil {
				connection.StateEvent.Stop()
			}

			break
		}
	}
}

// GetStateEventManager returns the state event manager
func (r *Room) GetStateEventManager() *event.EventManager {
	return r.stateEventManager
}

// GetStateEventTicker returns the state event ticker
func (r *Room) GetStateEventTicker() *event.Ticker {
	return r.stateEventTicker
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

// get play state
func (r *Room) GetUserPlaystate(username string) (interface{}, bool, error) {
	if username == "" {
		return nil, false, fmt.Errorf("username cannot be empty")
	}

	r.mutex.RLock()
	defer r.mutex.RUnlock()

	playstate, ok := r.PlaylistManager.GetUserPlaystate(username)
	if !ok {
		return nil, false, fmt.Errorf("user playstate not found for username: %s", username)
	}

	return playstate, true, nil
}

// room state
func (r *Room) GetRoomState() (roomState, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if (roomState{}) == r.RoomState {
		return roomState{}, fmt.Errorf("room state is empty")
	}

	return r.RoomState, nil
}

func (r *Room) SetRoomState(state roomState) error {
	if (roomState{}) == state {
		return fmt.Errorf("state cannot be empty")
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.RoomState = state
	return nil
}

// SetUserLatencyCalculation sets the client latency calculation struct
func (r *Room) SetUserLatencyCalculation(username string, arivalTime float64, clientTime float64, clientRtt float64, clientLatencyCalculation float64) error {
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	connection := r.GetConnectionByUsername(username)
	if connection == nil {
		return fmt.Errorf("connection not found for username: %s", username)
	}

	if connection.ClientLatencyCalculation == nil {
		connection.ClientLatencyCalculation = &ClientLatencyCalculation{}
	}

	connection.ClientLatencyCalculation.ArivalTime = arivalTime
	connection.ClientLatencyCalculation.ClientTime = clientTime
	connection.ClientLatencyCalculation.ClientRtt = clientRtt
	connection.ClientLatencyCalculation.clientLatencyCalculation = clientLatencyCalculation

	return nil
}

// GetUsersLatencyCalculation returns the client latency calculation struct
func (r *Room) GetUsersLatencyCalculation(connection *Connection) (ClientLatencyCalculation, error) {
	if connection == nil {
		return ClientLatencyCalculation{}, fmt.Errorf("connection cannot be nil")
	}

	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if connection.ClientLatencyCalculation == nil {
		return ClientLatencyCalculation{}, fmt.Errorf("client latency calculation is nil")
	}

	return *connection.ClientLatencyCalculation, nil
}
