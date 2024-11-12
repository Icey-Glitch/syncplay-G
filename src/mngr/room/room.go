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
	Owner      *Room
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
	Mutex           sync.RWMutex

	stateEventManager *event.EventManager
	stateEventTicker  *event.Ticker
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

// GetConnectionByConn get connection by conn
func (r *Room) GetConnectionByConn(conn net.Conn) (user *Connection, err error) {
	r.Mutex.RLock()
	defer r.Mutex.RUnlock()

	for _, connection := range r.Users {
		if connection.Conn == conn {
			return connection, nil
		}
	}
	return nil, fmt.Errorf("connection not found")
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
	r.Mutex.RLock()
	defer r.Mutex.RUnlock()

	for _, connection := range r.Users {
		if connection.Conn == conn {
			return connection.Username
		}
	}
	return ""
}

func (r *Room) GetConnections() []*Connection {
	r.Mutex.RLock()
	defer r.Mutex.RUnlock()

	return r.Users
}

func (r *Room) AddConnection(connection *Connection) error {
	if connection.Owner == nil {
		return fmt.Errorf("room is nil")
	}

	r.Mutex.Lock()
	defer r.Mutex.Unlock()

	if r.connectionExists(connection) {
		return fmt.Errorf("connection or username already exists")
	}

	r.Users = append(r.Users, connection)
	if err := r.PlaylistManager.CreateUserPlaystate(connection.Username); err != nil {
		fmt.Println("Failed to create user playstate " + err.Error())
		return err
	}
	return nil
}

func (r *Room) connectionExists(connection *Connection) bool {
	for _, conn := range r.Users {
		if conn.Conn == connection.Conn || conn.Username == connection.Username {
			return true
		}
	}
	return false
}

func (r *Room) RemoveConnection(conn net.Conn) {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()

	for i, connection := range r.Users {
		if connection.Conn == conn {
			r.removeUserStates(connection)
			r.Users = append(r.Users[:i], r.Users[i+1:]...)
			if connection.StateEvent != nil {
				connection.StateEvent.Stop()
			}
			break
		}
	}
}

// RemoveConnectionByUsername remove connection by username
func (r *Room) removeUserStates(connection *Connection) {
	r.ReadyManager.RemoveUserReadyState(connection.Username)
	if err := r.PlaylistManager.RemoveUserPlaystate(connection.Username); err != nil {
		fmt.Println("failed to remove UserPlaystate " + err.Error())
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

// ListRooms list rooms
func ListRooms(rooms map[string]*Room) []string {
	roomNames := make([]string, 0)
	for roomName := range rooms {
		roomNames = append(roomNames, roomName)
	}
	return roomNames
}

// SetUserReadyState ready state
func (r *Room) SetUserReadyState(username string, isReady bool, manuallyInitiated bool) {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()

	r.ReadyManager.SetUserReadyState(username, isReady, manuallyInitiated)
}

// PrintReadyStates print all ready states
func (r *Room) PrintReadyStates() {
	r.Mutex.RLock()
	defer r.Mutex.RUnlock()

	for _, connection := range r.Users {
		state, exists := r.ReadyManager.GetUserReadyState(connection.Username)
		if exists {
			fmt.Printf("Username: %s, IsReady: %t, ManuallyInitiated: %t\n", state.Username, state.IsReady, state.ManuallyInitiated)
		}
	}
}

// GetUserPlaystate get play state
func (r *Room) GetUserPlaystate(username string) (interface{}, bool, error) {
	if username == "" {
		return nil, false, fmt.Errorf("username cannot be empty")
	}

	r.Mutex.RLock()
	defer r.Mutex.RUnlock()

	playstate, ok := r.PlaylistManager.GetUserPlaystate(username)
	if !ok {
		return nil, false, fmt.Errorf("user playstate not found for username: %s", username)
	}

	return playstate, true, nil
}

// SetUserLatencyCalculation sets the client latency calculation struct
func (r *Room) SetUserLatencyCalculation(connection *Connection, arivalTime float64, clientTime float64, clientRtt float64, clientLatencyCalculation float64) error {
	if connection == nil {
		return fmt.Errorf("connection cannot be nil")
	}

	r.Mutex.Lock()
	defer r.Mutex.Unlock()

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

	r.Mutex.RLock()
	defer r.Mutex.RUnlock()

	if connection.ClientLatencyCalculation == nil {
		return ClientLatencyCalculation{}, fmt.Errorf("client latency calculation is nil")
	}

	return *connection.ClientLatencyCalculation, nil
}
