package roomM

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRoom(t *testing.T) {
	roomName := "testRoom"
	room := NewRoom(roomName)

	assert.Equal(t, roomName, room.Name)
	assert.NotNil(t, room.Users)
	assert.NotNil(t, room.ReadyManager)
	assert.NotNil(t, room.PlaylistManager)
	assert.NotNil(t, room.stateEventManager)
	assert.NotNil(t, room.stateEventTicker)
}

func TestAddConnection(t *testing.T) {
	room := NewRoom("testRoom")
	conn := &Connection{
		Username: "testUser",
		Conn:     &net.TCPConn{},
	}

	err := room.AddConnection(conn)
	if err != nil {
		return
	}

	assert.Equal(t, 1, len(room.Users))
	assert.Equal(t, "testUser", room.Users[0].Username)
}

func TestRemoveConnection(t *testing.T) {
	room := NewRoom("testRoom")
	conn := &Connection{
		Username: "testUser",
		Conn:     &net.TCPConn{},
	}

	err := room.AddConnection(conn)
	if err != nil {
		return
	}
	room.RemoveConnection(conn.Conn)

	assert.Equal(t, 0, len(room.Users))
}

func TestGetConnectionByUsername(t *testing.T) {
	room := NewRoom("testRoom")
	conn := &Connection{
		Username: "testUser",
		Conn:     &net.TCPConn{},
	}

	err := room.AddConnection(conn)
	if err != nil {
		return
	}
	retrievedConn := room.GetConnectionByUsername("testUser")

	assert.NotNil(t, retrievedConn)
	assert.Equal(t, "testUser", retrievedConn.Username)
}

func TestGetUsernameByConnection(t *testing.T) {
	room := NewRoom("testRoom")
	conn := &Connection{
		Username: "testUser",
		Conn:     &net.TCPConn{},
	}

	err := room.AddConnection(conn)
	if err != nil {
		return
	}
	username := room.GetUsernameByConnection(conn.Conn)

	assert.Equal(t, "testUser", username)
}

func TestSetUserReadyState(t *testing.T) {
	room := NewRoom("testRoom")
	conn := &Connection{
		Username: "testUser",
		Conn:     &net.TCPConn{},
	}

	err := room.AddConnection(conn)
	if err != nil {
		return
	}
	room.SetUserReadyState("testUser", true, true)

	state, exists := room.ReadyManager.GetUserReadyState("testUser")
	assert.True(t, exists)
	assert.True(t, state.IsReady)
	assert.True(t, state.ManuallyInitiated)
}

func TestGetUsersLatencyCalculation_NilConnection(t *testing.T) {
	room := NewRoom("testRoom")
	_, err := room.GetUsersLatencyCalculation(nil)
	assert.NotNil(t, err)
	assert.Equal(t, "connection cannot be nil", err.Error())
}

func TestGetUsersLatencyCalculation_NilClientLatencyCalculation(t *testing.T) {
	room := NewRoom("testRoom")
	conn := &Connection{
		Username: "testUser",
		Conn:     &net.TCPConn{},
	}

	err := room.AddConnection(conn)
	if err != nil {
		return
	}
	_, err = room.GetUsersLatencyCalculation(conn)
	assert.NotNil(t, err)
	assert.Equal(t, "client latency calculation is nil", err.Error())
}

func TestGetUsersLatencyCalculation(t *testing.T) {
	room := NewRoom("testRoom")
	conn := &Connection{
		Username: "testUser",
		Conn:     &net.TCPConn{},
		ClientLatencyCalculation: &ClientLatencyCalculation{
			ArivalTime: 10.5,
			ClientTime: 20.5,
			ClientRtt:  30.5,
		},
	}

	err := room.AddConnection(conn)
	if err != nil {
		return
	}
	latencyCalculation, err := room.GetUsersLatencyCalculation(conn)
	assert.Nil(t, err)
	assert.Equal(t, 10.5, latencyCalculation.ArivalTime)
	assert.Equal(t, 20.5, latencyCalculation.ClientTime)
	assert.Equal(t, 30.5, latencyCalculation.ClientRtt)

}

func TestGetUserPlaystate_EmptyUsername(t *testing.T) {
	room := NewRoom("testRoom")
	_, _, err := room.GetUserPlaystate("")
	assert.NotNil(t, err)
	assert.Equal(t, "username cannot be empty", err.Error())
}

func TestGetUserPlaystate_NonExistentUsername(t *testing.T) {
	room := NewRoom("testRoom")
	_, _, err := room.GetUserPlaystate("nonExistentUser")
	assert.NotNil(t, err)
	assert.Equal(t, "user playstate not found for username: nonExistentUser", err.Error())
}

func TestSetUserLatencyCalculation_EmptyUsername(t *testing.T) {
	room := NewRoom("testRoom")
	err := room.SetUserLatencyCalculation(nil, 0, 0, 0, 0)
	assert.NotNil(t, err)
	assert.Equal(t, "connection cannot be nil", err.Error())
}

func TestSetUserLatencyCalculation_NilClientLatencyCalculation(t *testing.T) {
	room := NewRoom("testRoom")
	conn := &Connection{
		Username: "testUser",
		Conn:     &net.TCPConn{},
	}

	room.AddConnection(conn)
	err := room.SetUserLatencyCalculation(conn, 0, 0, 0, 0)
	assert.Nil(t, err)
	assert.NotNil(t, conn.ClientLatencyCalculation)
}

func TestSetUserLatencyCalculation(t *testing.T) {
	room := NewRoom("testRoom")
	conn := &Connection{
		Username: "testUser",
		Conn:     &net.TCPConn{},
	}

	room.AddConnection(conn)
	err := room.SetUserLatencyCalculation(conn, 0, 0, 0, 0)
	assert.Nil(t, err)
	assert.NotNil(t, conn.ClientLatencyCalculation)
	assert.Equal(t, 0.0, conn.ClientLatencyCalculation.ArivalTime)
	assert.Equal(t, 0.0, conn.ClientLatencyCalculation.ClientTime)
	assert.Equal(t, 0.0, conn.ClientLatencyCalculation.ClientRtt)
}
func TestGetRoomByConnection(t *testing.T) {
	rooms := make(map[string]*Room)

	room1 := NewRoom("room1")
	room2 := NewRoom("room2")

	conn1 := &Connection{
		Username: "user1",
		Conn:     &net.TCPConn{},
	}
	conn2 := &Connection{
		Username: "user2",
		Conn:     &net.TCPConn{},
	}

	err := room1.AddConnection(conn1)
	if err != nil {
		return
	}
	err = room2.AddConnection(conn2)
	if err != nil {
		return
	}

	rooms["room1"] = room1
	rooms["room2"] = room2

	// Test for existing connection in room1
	foundRoom := GetRoomByConnection(conn1.Conn, rooms)
	assert.NotNil(t, foundRoom)
	assert.Equal(t, "room1", foundRoom.Name)

	// Test for existing connection in room2
	foundRoom = GetRoomByConnection(conn2.Conn, rooms)
	assert.NotNil(t, foundRoom)
	assert.Equal(t, "room2", foundRoom.Name)

	// Test for non-existing connection
	nonExistentConn := &net.TCPConn{}
	foundRoom = GetRoomByConnection(nonExistentConn, rooms)
	assert.Nil(t, foundRoom)
}
func TestGetConnectionByConn(t *testing.T) {
	room := NewRoom("testRoom")
	conn := &Connection{
		Username: "testUser",
		Conn:     &net.TCPConn{},
	}

	err := room.AddConnection(conn)
	if err != nil {
		return
	}
	retrievedConn := room.GetConnectionByConn(conn.Conn)

	assert.NotNil(t, retrievedConn)
	assert.Equal(t, "testUser", retrievedConn.Username)

	// Test for non-existing connection
	nonExistentConn := &net.TCPConn{}
	retrievedConn = room.GetConnectionByConn(nonExistentConn)
	assert.Nil(t, retrievedConn)
}
