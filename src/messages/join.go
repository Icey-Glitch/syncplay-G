package messages

import (
	"net"

	connM "github.com/Icey-Glitch/Syncplay-G/mngr/conn"
	roomM "github.com/Icey-Glitch/Syncplay-G/mngr/room"
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
		cm.CreateRoom(roomName)
	}

	room.AddConnection(&roomM.Connection{
		Username: username,
		Conn:     conn,
		RoomName: roomName,
	})

	broadcastJoinAnnouncement(username, roomName, cm)
}

func HandleUserLeftMessage(conn net.Conn) {
	// {"Set": {"user": {"Bob": {"room": {"name": "SyncRoom"}, "event": {"left": true}}}}}

	// print the incoming message

	cm := connM.GetConnectionManager()
	room := cm.GetRoomByConnection(conn)

	username := room.GetUsernameByConnection(conn)

	if room == nil {
		return
	}

	room.RemoveConnection(conn)

	broadcastLeaveAnnouncement(username, room.Name, cm)
}

func broadcastLeaveAnnouncement(username, roomName string, cm *connM.ConnectionManager) error {
	announcement := map[string]interface{}{
		"Set": map[string]interface{}{
			"user": map[string]interface{}{
				username: map[string]interface{}{
					"room": map[string]interface{}{
						"name": roomName,
					},
					"event": map[string]interface{}{
						"left": true,
					},
				},
			},
		},
	}
	utils.SendJSONMessageMultiCast(announcement, cm.GetRoom(roomName).Name)

	return nil
}

func broadcastJoinAnnouncement(username, roomName string, cm *connM.ConnectionManager) error {
	announcement := map[string]interface{}{
		"Set": map[string]interface{}{
			"user": map[string]interface{}{
				username: map[string]interface{}{
					"room": map[string]interface{}{
						"name": roomName,
					},
					"event": map[string]interface{}{
						"joined": true,
					},
				},
			},
		},
	}
	utils.SendJSONMessageMultiCast(announcement, cm.GetRoom(roomName).Name)

	return nil
}
