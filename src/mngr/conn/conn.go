package connM

import (
	"net"
	"sync"

	roomM "github.com/Icey-Glitch/Syncplay-G/mngr/room"
)

type ConnectionManager struct {
	rooms map[string]*roomM.Room
	mutex sync.RWMutex
}

func (cm *ConnectionManager) AddConnection(username, roomName string, state interface{}, conn net.Conn) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	connection := &roomM.Connection{
		Username: username,
		State:    state,
		Conn:     conn,
		RoomName: roomName,
	}

	room := cm.rooms[roomName]

	room.AddConnection(connection)
}

func (cm *ConnectionManager) RemoveConnection(conn net.Conn) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	for _, room := range cm.rooms {
		room.RemoveConnection(conn)
	}
}

func (cm *ConnectionManager) CreateRoom(roomName string) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cm.rooms[roomName] = roomM.NewRoom(roomName)
}

func (cm *ConnectionManager) GetRoom(roomName string) *roomM.Room {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	return cm.rooms[roomName]
}

// get room by connection
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

// get room by username
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

func GetConnectionManager() *ConnectionManager {
	if connectionManager == nil {
		connectionManager = &ConnectionManager{
			rooms: make(map[string]*roomM.Room),
		}
	}
	return connectionManager
}

var connectionManager *ConnectionManager
