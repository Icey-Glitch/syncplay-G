package messages

import (
	"fmt"
	"net"

	roomM "github.com/Icey-Glitch/Syncplay-G/mngr/room"

	connM "github.com/Icey-Glitch/Syncplay-G/mngr/conn"
	"github.com/Icey-Glitch/Syncplay-G/utils"
)

func HandleJoinMessage(conn net.Conn, msg map[string]interface{}) {
	// {"Set": {"user": {"Bob": {"room": {"name": "SyncRoom"}, "event": {"joined": true}}}}}

	// print the incoming message

	cm := connM.GetConnectionManager()
	roomName := msg["room"].(string)
	username := msg["username"].(string)

	room := cm.GetRoom(roomName)
	if room == nil {
		room = cm.CreateRoom(roomName)
	}

	// check if the connection exist / what room they are in and move them into the new room if they are in a different room
	// if they are in the same room do nothing

	if room.GetConnectionByUsername(username) != nil {
		// check if the user is in the same room
		user := room.GetConnectionByUsername(username)
		if user.Owner.Name == roomName {
			// do nothing
			return
		} else {
			// remove the user from the room
			HandleUserLeftMessage(*user)
			// send a leave message
			connection, err := cm.MoveConnection(username, roomName, user.Owner.Name, conn)
			if err != nil {
				fmt.Println("Error moving connection to room:", err)
				return
			}

			fmt.Println("Moved user to new room")

			// send a join message
			err = BroadcastJoinAnnouncement(*connection)
			if err != nil {
				fmt.Printf("Failed to send Join Anouncement" + err.Error())
				return
			}
			return
		}
	} else {
		connection, err := cm.AddConnection(username, roomName, nil, conn)
		if err != nil {
			fmt.Println("Error adding connection to room:", err)
			return
		}
		BroadcastUserRoomChangeMessage(*connection, roomName)
		return
	}

}

type RoomMessage struct {
	Name string `json:"name"`
}

// User Move room message
func HandleUserMoveRoomMessage(connection roomM.Connection, msg *RoomMessage) {
	// {"Set": {"room": {"name": "room"}}}

	roomName := msg.Name

	// Move the user to the new room
	cm := connM.GetConnectionManager()
	oldRoom := connection.Owner
	newRoom := cm.GetRoom(roomName)
	if newRoom == nil {
		newRoom = cm.CreateRoom(roomName)
	}

	if oldRoom.Name != newRoom.Name {

		HandleUserLeftMessage(connection)
		_, err := cm.MoveConnection(connection.Username, newRoom.Name, oldRoom.Name, connection.Conn)
		if err != nil {
			fmt.Println("Error moving connection to new room:", err)
			return
		}

		// err = BroadcastJoinAnnouncement(connection)
		// if err != nil {
		// 	fmt.Printf("Failed to send Join Announcement: " + err.Error())
		// 	return
		// }
	}
}

func HandleUserLeftMessage(connection roomM.Connection) {
	// {"Set": {"user": {"Bob": {"room": {"name": "SyncRoom"}, "event": {"left": true}}}}}

	// print the incoming message

	room := connection.Owner

	if room == nil {
		return
	}

	err := broadcastLeaveAnnouncement(connection)
	if err != nil {
		fmt.Printf("Failed to send Leave Anouncement" + err.Error())
		return
	}
}

type RoomChangeMessage struct {
	Set struct {
		Room struct {
			Name string `json:"name"`
		} `json:"room"`
	} `json:"Set"`
}

// Broadcast User Room change message
func BroadcastUserRoomChangeMessage(connection roomM.Connection, roomName string) {
	announcement := RoomChangeMessage{}
	announcement.Set.Room.Name = roomName
	utils.SendJSONMessageMultiCast(announcement, connection.Owner)
}

func broadcastLeaveAnnouncement(connection roomM.Connection) error {
	announcement := map[string]interface{}{
		"Set": map[string]interface{}{
			"user": map[string]interface{}{
				connection.Username: map[string]interface{}{
					"room": map[string]interface{}{
						"name": connection.Owner.Name,
					},
					"event": map[string]interface{}{
						"left": true,
					},
				},
			},
		},
	}
	utils.SendJSONMessageMultiCast(announcement, connection.Owner)

	return nil
}

func BroadcastJoinAnnouncement(connection roomM.Connection) error {
	announcement := map[string]interface{}{
		"Set": map[string]interface{}{
			"user": map[string]interface{}{
				connection.Username: map[string]interface{}{
					"room": map[string]interface{}{
						"name": connection.Owner.Name,
					},
					"event": map[string]interface{}{
						"joined": true,
					},
				},
			},
		},
	}
	utils.SendJSONMessageMultiCast(announcement, connection.Owner)

	return nil
}
