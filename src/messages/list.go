package messages

import (
	"fmt"
	"net"
	"sync"

	roomM "github.com/Icey-Glitch/Syncplay-G/mngr/room"
)

// FileInfo represents the file information.
type FileInfo struct {
	Duration float64 `json:"duration"`
	Name     string  `json:"name"`
	Size     float64 `json:"size"`
}

// PlayerInfo represents the player information.
type PlayerInfo struct {
	Position *float64 `json:"position,omitempty"`
	File     FileInfo `json:"file"`
}

// RoomInfo represents the room information.
type RoomInfo map[string]PlayerInfo

// ListResponse represents the list response.
type ListResponse struct {
	List map[string]RoomInfo `json:"List"`
}

// handleListRequest handles the "List" request and returns the response.
func HandleListRequest(conn net.Conn, room *roomM.Room) {
	fmt.Println("List request received")

	// Initialize the list map
	list := make(map[string]RoomInfo)

	// Retrieve user states from the PlaylistManager
	users, valid := room.PlaylistManager.GetUsers()
	if !valid {
		fmt.Println("Error: Failed to retrieve user states from the PlaylistManager")
		return
	}

	// Use a mutex to ensure thread-safe access to shared resources
	var mutex sync.RWMutex

	// Iterate over the users and construct the room info
	for _, user := range users {
		fileInfo := FileInfo{
			Duration: user.Duration,
			Name:     user.Name,
			Size:     user.Size,
		}
		playerInfo := PlayerInfo{
			File: fileInfo,
		}
		if user.Position != 0 {
			playerInfo.Position = &user.Position
		}

		// Lock the mutex for writing
		mutex.Lock()
		// Check if the room already exists in the list
		if _, exists := list[room.Name]; !exists {
			list[room.Name] = make(RoomInfo)
		}

		// Add the player info to the room info
		list[room.Name][user.Username] = playerInfo
		mutex.Unlock()
	}

	// Construct the list response
	response := ListResponse{
		List: list,
	}

	fmt.Println(response)
	// Send the response back to the client
	/*
		err := utils.SendJSONMessage(conn, responseBytes, room.PlaylistManager, room.GetUsernameByConnection(conn))
		if err != nil {
			fmt.Println("Error: Failed to send list response to", room.GetUsernameByConnection(conn), ":", err)
			return
		}

		fmt.Println("List response successfully sent to", room.GetUsernameByConnection(conn))
	*/

}
