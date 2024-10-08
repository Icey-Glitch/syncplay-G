package messages

import (
	"encoding/json"
	"fmt"
	"net"

	roomM "github.com/Icey-Glitch/Syncplay-G/mngr/room"
	"github.com/Icey-Glitch/Syncplay-G/utils"
)

// FileInfo represents the file information.
type FileInfo struct {
	Duration float64 `json:"duration"`
	Name     string  `json:"name"`
	Size     float64 `json:"size"`
}

// PlayerInfo represents the player information.
type PlayerInfo struct {
	Position *int     `json:"position,omitempty"`
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
	// Retrieve user states from the PlaylistManager
	userStates := room.PlaylistManager.GetUserStates()

	// Construct the players map
	players := make(map[string]PlayerInfo)
	for username, userState := range userStates {
		fileInfo := FileInfo{
			Duration: userState.Duration,
			Name:     userState.Name,
			Size:     userState.Size,
		}

		playerInfo := PlayerInfo{
			File: fileInfo,
		}

		// Add position if it exists
		if userState.Position != 0 {
			playerInfo.Position = &userState.Position
		}

		players[username] = playerInfo
	}

	// Retrieve the room name
	roomName := room.Name

	// Construct the room info
	roomInfo := RoomInfo(players)

	// Construct the list response
	listResponse := ListResponse{
		List: map[string]RoomInfo{
			roomName: roomInfo,
		},
	}

	// Convert the response to JSON
	response, err := json.Marshal(listResponse)
	if err != nil {
		fmt.Println("Error marshalling list response:", err)
		return
	}

	// Send the response
	utils.SendJSONMessage(conn, listResponse)

	fmt.Println("List response sent" + string(response))
}
