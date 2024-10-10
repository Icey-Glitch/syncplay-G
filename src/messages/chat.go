package messages

import (
	connM "github.com/Icey-Glitch/Syncplay-G/mngr/conn"
	"github.com/Icey-Glitch/Syncplay-G/utils"
)

//client {"Chat": "sample chat message"}
//server {"Chat": {"message": "sample chat message", "username": "sample user"}}

type ChatMessage struct {
	Chat struct {
		Message  string `json:"message"`
		Username string `json:"username"`
	} `json:"Chat"`
}

func SendChatMessage(message, username string) {
	room := connM.GetConnectionManager().GetRoomByUsername(username)
	chatMessage := ChatMessage{}
	chatMessage.Chat.Message = message
	chatMessage.Chat.Username = username

	utils.SendJSONMessageMultiCast(chatMessage, room.Name)
}
