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
	assert.NotNil(t, &cm.mutex)
	assert.NotNil(t, cm.connectionEvent)
}

func TestGetConnectionManager(t *testing.T) {
	cm := GetConnectionManager()

	assert.NotNil(t, cm)
	assert.NotNil(t, cm.rooms)
	assert.NotNil(t, &cm.mutex)
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

func TestCreateRoom(t *testing.T) {
	cm := NewConnectionManager()
	roomName := "testRoom"

	room := cm.CreateRoom(roomName)

	assert.NotNil(t, room)
	assert.Equal(t, roomName, room.Name)
}

func TestGetRoom(t *testing.T) {
	cm := NewConnectionManager()
	roomName := "testRoom"

	room := cm.CreateRoom(roomName)

	assert.NotNil(t, room)
	assert.Equal(t, room, cm.GetRoom(roomName))

	roomName2 := "testRoom2"
	room2 := cm.CreateRoom(roomName2)

	assert.NotNil(t, room2)
	assert.Equal(t, room2, cm.GetRoom(roomName2))

	assert.Nil(t, cm.GetRoom("nonExistentRoom"))
}

func TestGetRoomByConnection(t *testing.T) {
	cm := NewConnectionManager()
	roomName := "testRoom"
	room := cm.CreateRoom(roomName)

	conn := &net.TCPConn{}
	roomConn, err := cm.AddConnection("testUser", roomName, "testState", conn)
	assert.NoError(t, err)

	assert.NotNil(t, roomConn)
	assert.Equal(t, room, cm.GetRoomByConnection(conn))

	roomName2 := "testRoom2"
	room2 := cm.CreateRoom(roomName2)

	conn2 := &net.TCPConn{}
	roomConn2, err := cm.AddConnection("testUser2", roomName2, "testState2", conn2)
	assert.NoError(t, err)

	assert.NotNil(t, roomConn2)
	assert.Equal(t, room2, cm.GetRoomByConnection(conn2))

	nonExistentConn := &net.TCPConn{}
	assert.Nil(t, cm.GetRoomByConnection(nonExistentConn))

	assert.Equal(t, room2, cm.GetRoomByConnection(conn2))
	assert.Equal(t, room, cm.GetRoomByConnection(conn))
}

func TestMoveConnection(t *testing.T) {
	cm := NewConnectionManager()
	roomName1 := "testRoom1"
	roomName2 := "testRoom2"

	cm.CreateRoom(roomName1)
	cm.CreateRoom(roomName2)

	conn := &net.TCPConn{}
	_, err := cm.AddConnection("testUser", roomName1, "testState", conn)
	assert.NoError(t, err)

	movedConn, err := cm.MoveConnection("testUser", roomName2, roomName1, conn)
	assert.NoError(t, err)

	assert.NotNil(t, movedConn)
	assert.Equal(t, roomName2, movedConn.RoomName)
}

func TestRemoveConnection(t *testing.T) {
	cm := NewConnectionManager()
	roomName := "testRoom"
	room := cm.CreateRoom(roomName)

	conn := &net.TCPConn{}
	_, err := cm.AddConnection("testUser", roomName, "testState", conn)
	assert.NoError(t, err)

	cm.RemoveConnection(conn)
	assert.Nil(t, room.GetConnectionByUsername("testUser"))
}

func TestGetRoomByUsername(t *testing.T) {
	cm := NewConnectionManager()
	roomName := "testRoom"
	room := cm.CreateRoom(roomName)

	conn := &net.TCPConn{}
	_, err := cm.AddConnection("testUser", roomName, "testState", conn)
	assert.NoError(t, err)

	assert.Equal(t, room, cm.GetRoomByUsername("testUser"))
	assert.Nil(t, cm.GetRoomByUsername("nonExistentUser"))
}

func TestSubscribeToConnections(t *testing.T) {
	cm := NewConnectionManager()
	ch := cm.SubscribeToConnections()

	assert.NotNil(t, ch)
}

func TestUnsubscribeFromConnections(t *testing.T) {
	cm := NewConnectionManager()
	ch := cm.SubscribeToConnections()

	assert.NotNil(t, ch)
	cm.UnsubscribeFromConnections(ch)
}
