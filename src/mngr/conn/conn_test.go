package connM

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConnectionManager(t *testing.T) {
	cm := NewConnectionManager()

	assert.NotNil(t, cm)
	assert.NotNil(t, cm.rooms)
	assert.NotNil(t, cm.mutex)
	assert.NotNil(t, cm.connectionEvent)
}

func TestGetConnectionManager(t *testing.T) {
	cm := GetConnectionManager()

	assert.NotNil(t, cm)
	assert.NotNil(t, cm.rooms)
	assert.NotNil(t, cm.mutex)
	assert.NotNil(t, cm.connectionEvent)
}

func TestAddConnection(t *testing.T) {
	cm := NewConnectionManager()
	username := "testUser"
	roomName := "testRoom"
	state := "testState"

	conn := &net.TCPConn{}

	room := cm.CreateRoom(roomName)
	cm.rooms[roomName] = room

	roomConn, err := cm.AddConnection(username, roomName, state, conn)
	assert.NoError(t, err)

	assert.NotNil(t, roomConn)
	assert.Equal(t, username, roomConn.Username)
	assert.Equal(t, state, roomConn.State)
	assert.Equal(t, conn, roomConn.Conn)
	assert.Equal(t, roomName, roomConn.RoomName)
}
