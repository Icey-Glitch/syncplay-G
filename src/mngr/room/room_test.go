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

	room.AddConnection(conn)

	assert.Equal(t, 1, len(room.Users))
	assert.Equal(t, "testUser", room.Users[0].Username)
}

func TestRemoveConnection(t *testing.T) {
	room := NewRoom("testRoom")
	conn := &Connection{
		Username: "testUser",
		Conn:     &net.TCPConn{},
	}

	room.AddConnection(conn)
	room.RemoveConnection(conn.Conn)

	assert.Equal(t, 0, len(room.Users))
}

func TestGetConnectionByUsername(t *testing.T) {
	room := NewRoom("testRoom")
	conn := &Connection{
		Username: "testUser",
		Conn:     &net.TCPConn{},
	}

	room.AddConnection(conn)
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

	room.AddConnection(conn)
	username := room.GetUsernameByConnection(conn.Conn)

	assert.Equal(t, "testUser", username)
}

func TestSetUserReadyState(t *testing.T) {
	room := NewRoom("testRoom")
	conn := &Connection{
		Username: "testUser",
		Conn:     &net.TCPConn{},
	}

	room.AddConnection(conn)
	room.SetUserReadyState("testUser", true, true)

	state, exists := room.ReadyManager.GetUserReadyState("testUser")
	assert.True(t, exists)
	assert.True(t, state.IsReady)
	assert.True(t, state.ManuallyInitiated)
}

func TestSetRoomState(t *testing.T) {
	room := NewRoom("testRoom")
	state := roomState{
		IsPaused: true,
		Position: 10.5,
		SetBy:    "testUser",
	}

	err := room.SetRoomState(state)
	assert.Nil(t, err)

	retrievedState, err := room.GetRoomState()
	assert.Nil(t, err)
	assert.Equal(t, state, retrievedState)
}

func TestSetUserLatencyCalculation(t *testing.T) {
	room := NewRoom("testRoom")
	conn := &Connection{
		Username: "testUser",
		Conn:     &net.TCPConn{},
	}

	room.AddConnection(conn)
	err := room.SetUserLatencyCalculation("testUser", 1.0, 2.0, 3.0, 4.0)
	assert.Nil(t, err)

	latencyCalc, err := room.GetUsersLatencyCalculation(conn)
	assert.Nil(t, err)
	assert.Equal(t, 1.0, latencyCalc.ArivalTime)
	assert.Equal(t, 2.0, latencyCalc.ClientTime)
	assert.Equal(t, 3.0, latencyCalc.ClientRtt)
	assert.Equal(t, 4.0, latencyCalc.clientLatencyCalculation)
}

func TestSetUserLatencyCalculation_EmptyUsername(t *testing.T) {
	room := NewRoom("testRoom")
	err := room.SetUserLatencyCalculation("", 1.0, 2.0, 3.0, 4.0)
	assert.NotNil(t, err)
	assert.Equal(t, "username cannot be empty", err.Error())
}

func TestSetUserLatencyCalculation_NonExistentUsername(t *testing.T) {
	room := NewRoom("testRoom")
	err := room.SetUserLatencyCalculation("nonExistentUser", 1.0, 2.0, 3.0, 4.0)
	assert.NotNil(t, err)
	assert.Equal(t, "connection not found for username: nonExistentUser", err.Error())
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

	room.AddConnection(conn)
	_, err := room.GetUsersLatencyCalculation(conn)
	assert.NotNil(t, err)
	assert.Equal(t, "client latency calculation is nil", err.Error())
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

func TestSetRoomState_EmptyState(t *testing.T) {
	room := NewRoom("testRoom")
	err := room.SetRoomState(roomState{})
	assert.NotNil(t, err)
	assert.Equal(t, "state cannot be empty", err.Error())
}

func TestGetRoomState_EmptyRoomState(t *testing.T) {
	room := NewRoom("testRoom")
	_, err := room.GetRoomState()
	assert.NotNil(t, err)
	assert.Equal(t, "room state is empty", err.Error())
}
