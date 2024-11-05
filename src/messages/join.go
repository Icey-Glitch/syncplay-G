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
		_ = cm.CreateRoom(roomName)
	}

	connection, err := cm.AddConnection(username, roomName, nil, conn)
	if err != nil {
		fmt.Println("Error adding connection to room:", err)
		return
	}

	err = BroadcastJoinAnnouncement(*connection)
	if err != nil {
		fmt.Printf("Failed to send Join Anouncement" + err.Error())
		return
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
	utils.SendJSONMessageMultiCast(announcement, connection.Owner.Name)

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
	utils.SendJSONMessageMultiCast(announcement, connection.Owner.Name)

	return nil
}
