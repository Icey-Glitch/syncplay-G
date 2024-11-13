package utils

import (
	"bytes"
	"fmt"
	"net"
	"sync"

	"github.com/goccy/go-json"

	connM "github.com/Icey-Glitch/Syncplay-G/mngr/conn"
)

// InsertSpaceAfterColons inserts a space after each colon in the JSON byte slice
func InsertSpaceAfterColons(jsonData []byte) []byte {
	modifiedData := make([]byte, 0, len(jsonData))
	for i := 0; i < len(jsonData); i++ {
		modifiedData = append(modifiedData, jsonData[i])
		if jsonData[i] == ':' {
			modifiedData = append(modifiedData, 0x20) // Insert space (0x20) after colon
		}
		if jsonData[i] == ',' {
			modifiedData = append(modifiedData, 0x20) // Insert space (0x20) after colon
		}
	}
	return modifiedData
}

// SendJSONMessage marshals the message to JSON, inserts spaces after colons, and sends it to the connection
var sendMutex sync.Mutex

func SendJSONMessage(conn net.Conn, message interface{}) error {
	if conn == nil {
		return fmt.Errorf("connection is nil")
	}

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("error marshalling JSON message: %w", err)
	}

	// Append CRLF
	data = append(data, '\r', '\n')

	// Write data to connection
	_, err = conn.Write(data)
	if err != nil {
		return fmt.Errorf("error writing JSON message to connection: %w", err)
	}

	return nil
}

func SendJSONMessageMultiCast(message interface{}, roomName string) {
	// send to all connections in the room
	cm := connM.GetConnectionManager()
	room := cm.GetRoom(roomName)
	if room == nil {
		return
	}

	for _, user := range room.Users {
		err := SendJSONMessage(user.Conn, message)
		if err != nil {
			return
		}
	}

}

// PrettyPrintJSON takes a JSON byte slice, formats it with indentation, and prints it to the console
func PrettyPrintJSON(jsonData []byte) {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, jsonData, "", "  "); err != nil {
		fmt.Println("Error formatting JSON:", err)
		return
	}
	fmt.Println(prettyJSON.String())
}
