package messages

import (
	"encoding/json"
	"net"

	connM "github.com/Icey-Glitch/Syncplay-G/mngr/conn"
	roomM "github.com/Icey-Glitch/Syncplay-G/mngr/room"
	"github.com/Icey-Glitch/Syncplay-G/utils"
)

func HandleJoinMessage(conn net.Conn, msg map[string]interface{}) {
	// {"Set": {"user": {"Bob": {"room": {"name": "SyncRoom"}, "event": {"joined": true}}}}}

	// print the incoming message
	chatBytes, _ := json.Marshal(msg)
	utils.PrettyPrintJSON(utils.InsertSpaceAfterColons(chatBytes))

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

	// pritty print the message
	chatBytes, _ := json.Marshal(announcement)
	utils.PrettyPrintJSON(utils.InsertSpaceAfterColons(chatBytes))

	return nil
}
