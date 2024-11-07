package roomM

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
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

// GetConnectionByConn
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

	retrievedConn, err := room.GetConnectionByConn(conn.Conn)
	assert.Nil(t, err)
	assert.NotNil(t, retrievedConn)
	assert.Equal(t, "testUser", retrievedConn.Username)

	// Test for non-existing connection
	nonExistentConn := &net.TCPConn{}
	retrievedConn, err = room.GetConnectionByConn(nonExistentConn)
	assert.NotNil(t, err)
	assert.Nil(t, retrievedConn)
}

// GetRoomByConnection
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

func TestListRooms(t *testing.T) {
	rooms := make(map[string]*Room)
	rooms["room1"] = NewRoom("room1")
	rooms["room2"] = NewRoom("room2")

	roomNames := ListRooms(rooms)
	assert.Contains(t, roomNames, "room1")
	assert.Contains(t, roomNames, "room2")
}

func TestPrintReadyStates(t *testing.T) {
	room := NewRoom("testRoom")
	conn := &Connection{
		Username: "testUser",
		Conn:     &net.TCPConn{},
		Owner:    room,
	}

	err := room.AddConnection(conn)
	if err != nil {
		t.Fatalf("Failed to add connection: %v", err)
	}
	room.SetUserReadyState("testUser", true, true)

	// Capture the output of PrintReadyStates
	output := captureOutput(func() {
		room.PrintReadyStates()
	})

	fmt.Println("Captured Output:", output)
	assert.Contains(t, output, "Username: testUser, IsReady: true, ManuallyInitiated: true")
}

// RemoveConnection
func TestRemoveConnection(t *testing.T) {
	room := NewRoom("testRoom")
	conn := &Connection{
		Username: "testUser",
		Conn:     &net.TCPConn{},
		Owner:    room,
	}

	err := room.AddConnection(conn)
	if err != nil {
		t.Fatalf("Failed to add connection: %v", err)
	}

	room.RemoveConnection(conn.Conn)
	assert.Equal(t, 0, len(room.Users))
}

func TestGetConnections(t *testing.T) {
	room := NewRoom("testRoom")
	conn1 := &Connection{
		Username: "testUser1",
		Conn:     &net.TCPConn{},
		Owner:    room,
	}
	conn2 := &Connection{
		Username: "testUser2",
		Conn:     &net.TCPConn{},
		Owner:    room,
	}

	err := room.AddConnection(conn1)
	if err != nil {
		t.Fatalf("Failed to add connection 1: %v", err)
	}
	err = room.AddConnection(conn2)
	if err != nil {
		t.Fatalf("Failed to add connection 2: %v", err)
	}

	connections := room.GetConnections()
	assert.Equal(t, 2, len(connections), "Expected 2 connections, got %d", len(connections))
	assert.Equal(t, "testUser1", connections[0].Username)
	assert.Equal(t, "testUser2", connections[1].Username)
}

func TestGetStateEventManager(t *testing.T) {
	room := NewRoom("testRoom")
	manager := room.GetStateEventManager()
	assert.NotNil(t, manager)
}

func TestGetStateEventTicker(t *testing.T) {
	room := NewRoom("testRoom")
	ticker := room.GetStateEventTicker()
	assert.NotNil(t, ticker)
}

// Helper function to capture output
func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}
