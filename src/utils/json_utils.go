package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
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
func SendJSONMessage(conn net.Conn, message interface{}) {
	jsonData, err := json.Marshal(message)
	if err != nil {
		fmt.Println("Error marshaling message:", err)
		return
	}

	jsonData = InsertSpaceAfterColons(jsonData)
	jsonData = append(jsonData, '\n')

	if _, err := conn.Write(jsonData); err != nil {
		fmt.Println("Error sending message:", err)
	} else {
		fmt.Printf("Sent message: %s\n", jsonData)
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
