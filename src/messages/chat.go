package messages

import (
	"encoding/json"

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
	chatMessage := ChatMessage{}
	chatMessage.Chat.Message = message
	chatMessage.Chat.Username = username

	chatBytes, _ := json.Marshal(chatMessage)
	utils.PrettyPrintJSON(utils.InsertSpaceAfterColons(chatBytes))
	utils.SendJSONMessageMultiCast(chatMessage)
}
