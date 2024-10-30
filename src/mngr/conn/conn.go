package connM

import (
	"fmt"
	"net"
	"sync"

	"github.com/Icey-Glitch/Syncplay-G/mngr/event"
	roomM "github.com/Icey-Glitch/Syncplay-G/mngr/room"
)

/*
TODO:
- Clean up orphin connections and rooms
*/

type ConnectionManager struct {
	rooms           map[string]*roomM.Room
	mutex           sync.RWMutex
	connectionEvent *event.Event
}

func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		rooms:           make(map[string]*roomM.Room),
		connectionEvent: event.NewEvent(),
	}
}

func GetConnectionManager() *ConnectionManager {
	if connectionManager == nil {
		connectionManager = &ConnectionManager{
			rooms:           make(map[string]*roomM.Room),
			connectionEvent: event.NewEvent(),
		}
	}
	return connectionManager
}

var connectionManager *ConnectionManager

func (cm *ConnectionManager) AddConnection(username, roomName string, state interface{}, conn net.Conn) *roomM.Connection {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	connection := &roomM.Connection{
		Username: username,
		State:    state,
		Conn:     conn,
		RoomName: roomName,

		ClientLatencyCalculation: &roomM.ClientLatencyCalculation{
			ArivalTime: float64(0),
			ClientTime: float64(0),
			ClientRtt:  float64(0),
		},

		Owner: cm.rooms[roomName],
	}

	room := cm.rooms[roomName]
	err := room.AddConnection(connection)
	if err != nil {
		fmt.Errorf("Failed to add connection to room: %s", err.Error())
		return nil
	}

	cm.connectionEvent.Publish(connection)
	return connection
}

func (cm *ConnectionManager) RemoveConnection(conn net.Conn) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	for _, room := range cm.rooms {
		room.RemoveConnection(conn)
	}

	conn.Close()

	cm.connectionEvent.Publish(conn)
}

func (cm *ConnectionManager) CreateRoom(roomName string) *roomM.Room {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cm.rooms[roomName] = roomM.NewRoom(roomName)
	return cm.rooms[roomName]
}

func (cm *ConnectionManager) GetRoom(roomName string) *roomM.Room {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	return cm.rooms[roomName]
}

func (cm *ConnectionManager) GetRoomByConnection(conn net.Conn) *roomM.Room {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	for _, room := range cm.rooms {
		if roomM.GetRoomByConnection(conn, cm.rooms) != nil {
			return room
		}
	}
	return nil
}

func (cm *ConnectionManager) GetRoomByUsername(username string) *roomM.Room {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	for _, room := range cm.rooms {
		if room.GetConnectionByUsername(username) != nil {
			return room
		}
	}
	return nil
}

func (cm *ConnectionManager) SubscribeToConnections() chan interface{} {
	return cm.connectionEvent.Subscribe()
}

func (cm *ConnectionManager) UnsubscribeFromConnections(ch chan interface{}) {
	cm.connectionEvent.Unsubscribe(ch)
}
