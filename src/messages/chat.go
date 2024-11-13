package messages

import (
	"net"

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

type ClientChatMessage struct {
	Chat string `json:"chat"`
}

func SendChatMessage(message, username string) {
	room := connM.GetConnectionManager().GetRoomByUsername(username)
	chatMessage := ChatMessage{}
	chatMessage.Chat.Message = message
	chatMessage.Chat.Username = username

	utils.SendJSONMessageMultiCast(chatMessage, room)
}

func SendMessageToUser(message string, username string, conn net.Conn) {
	chatMessage := ChatMessage{}
	chatMessage.Chat.Message = message
	chatMessage.Chat.Username = username

	err := utils.SendJSONMessage(conn, chatMessage)
	if err != nil {
		return
	}
}
