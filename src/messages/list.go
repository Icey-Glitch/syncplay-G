package messages

import (
	"fmt"
	"sync"

	Features "github.com/Icey-Glitch/Syncplay-G/features"
	"github.com/Icey-Glitch/Syncplay-G/utils"

	roomM "github.com/Icey-Glitch/Syncplay-G/mngr/room"
)

// client {"List": null}
// ListRequest represents the list request message.
type ListRequest struct {
	List interface{} `json:"List"`
}

// FileInfo represents the file information.
type FileInfo struct {
	Duration float64 `json:"duration"`
	Name     string  `json:"name"`
	Size     float64 `json:"size"`
}

// PlayerInfo represents the player information.
/*
"car1": {
        "position": 0,
        "file": {},
        "controller": false,
        "isReady": false,
        "features": {
          "sharedPlaylists": true,
          "chat": true,
          "uiMode": "GUI",
          "featureList": true,
          "readiness": true,
          "managedRooms": true,
          "persistentRooms": true
        }
      }
*/
type PlayerInfo struct {
	Position   *float64    `json:"position,omitempty"`
	File       interface{} `json:"file"`
	Controller bool        `json:"controller"`
	IsReady    bool        `json:"isReady"`
	Features   FeaturesList
}

type FeaturesList struct {
	SharedPlaylists bool   `json:"sharedPlaylists"`
	Chat            bool   `json:"chat"`
	UIMode          string `json:"uiMode"`
	FeatureList     bool   `json:"featureList"`
	Readiness       bool   `json:"readiness"`
	ManagedRooms    bool   `json:"managedRooms"`
	PersistentRooms bool   `json:"persistentRooms"`
}

// RoomInfo represents the room information.
type RoomInfo map[string]PlayerInfo

// ListResponse represents the list response.
type ListResponse struct {
	List map[string]RoomInfo `json:"List"`
}

// handleListRequest handles the "List" request and returns the response.
func HandleListRequest(connection roomM.Connection) {
	//fmt.Println("List request received")

	// Initialize the list map
	list := make(map[string]RoomInfo)

	// Retrieve user states from the PlaylistManager
	var users, valid = connection.Owner.PlaylistManager.GetUsers()
	if !valid {
		fmt.Println("Error: Failed to retrieve user states from the PlaylistManager")
		return
	}

	// retrive ready states from the ReadyManager
	var readyStates = connection.Owner.ReadyManager.GetReadyStates()

	// global features
	var features = Features.GlobalFeatures

	// Use a mutex to ensure thread-safe access to shared resources
	var mutex sync.RWMutex

	// Iterate over the users and construct the room info
	for _, user := range users {
		// if empty "file": {},
		var fileInfo interface{}
		if user.File != nil {
			fileInfo = FileInfo{
				Duration: user.File.Duration,
				Name:     user.File.Name,
				Size:     user.File.Size,
			}
		} else {
			fileInfo = struct{}{}
		}

		playerInfo := PlayerInfo{
			File: fileInfo,
		}

		playerInfo.Position = &user.Position
		playerInfo.Controller = false
		playerInfo.IsReady = readyStates[user.Username].IsReady
		playerInfo.Features = FeaturesList{
			SharedPlaylists: features.SharedPlaylists,
			Chat:            features.Chat,
			UIMode:          "GUI",
			FeatureList:     true,
			Readiness:       features.Readiness,
			ManagedRooms:    features.ManagedRooms,
			PersistentRooms: features.PersistentRooms,
		}

		// Lock the mutex for writing
		mutex.Lock()
		// Check if the room already exists in the list
		if _, exists := list[connection.Owner.Name]; !exists {
			list[connection.Owner.Name] = make(RoomInfo)
		}

		// Add the player info to the room info
		list[connection.Owner.Name][user.Username] = playerInfo
		mutex.Unlock()
	}

	// Construct the list response
	response := ListResponse{
		List: list,
	}

	// prety print the response
	// jsonResponse, err := json.MarshalIndent(response, "", "  ")
	// if err != nil {
	// 	fmt.Println("Error: Failed to marshal list response:", err)
	// 	return
	// }
	// fmt.Println("List response:", string(jsonResponse))

	err1 := utils.SendJSONMessage(connection.Conn, response)
	if err1 != nil {
		fmt.Println("Error: Failed to send list response to", connection.Username, ":", err1)
		return
	}

}
