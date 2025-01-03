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
	connToRoom      map[net.Conn]*roomM.Room // Map to store connection to room mapping
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
			connToRoom:      make(map[net.Conn]*roomM.Room),
		}
	}
	return connectionManager
}

var connectionManager *ConnectionManager

func (cm *ConnectionManager) AddConnection(username, roomName string, state interface{}, conn net.Conn) (*roomM.Connection, error) {
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
	cm.connToRoom[conn] = room // Update the map

	err := room.AddConnection(connection)
	if err != nil {
		err1 := fmt.Errorf("failed to add connection to room: %s", err.Error())
		return nil, err1
	}

	cm.connectionEvent.Publish(connection)
	return connection, nil
}

func (cm *ConnectionManager) MoveConnection(username string, newRoomName string, oldRoomName string, conn net.Conn) (*roomM.Connection, error) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	oldRoom := cm.rooms[oldRoomName]
	newRoom := cm.rooms[newRoomName]

	if oldRoom == nil {
		return nil, fmt.Errorf("old room does not exist")
	}

	if newRoom == nil {
		return nil, fmt.Errorf("new room does not exist")
	}

	connection := oldRoom.GetConnectionByUsername(username)
	if connection == nil {
		return nil, fmt.Errorf("user does not exist in the old room")
	}

	oldRoom.RemoveConnection(conn)

	// Avoid deadlock by not calling AddConnection directly
	connectionobj := &roomM.Connection{
		Username: username,
		State:    connection.State,
		Conn:     conn,
		RoomName: newRoomName,

		ClientLatencyCalculation: &roomM.ClientLatencyCalculation{
			ArivalTime: float64(0),
			ClientTime: float64(0),
			ClientRtt:  float64(0),
		},

		Owner: cm.rooms[newRoomName],
	}

	err := newRoom.AddConnection(connection)
	if err != nil {
		err1 := fmt.Errorf("failed to add connection to room: %s", err.Error())
		return nil, err1
	}

	cm.connectionEvent.Publish(connectionobj)
	return connectionobj, nil
}

// TODO: maybe deprecate, very expensive to iterate over all rooms. if you have the room, or conn you can do it directly on the room.
func (cm *ConnectionManager) RemoveConnection(conn net.Conn) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	for _, room := range cm.rooms {
		room.RemoveConnection(conn)
	}

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

	if cm.rooms[roomName] == nil {
		return nil
	}

	return cm.rooms[roomName]
}

func (cm *ConnectionManager) GetRoomByConnection(conn net.Conn) *roomM.Room {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	return cm.connToRoom[conn]
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
