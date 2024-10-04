package messages

import (
	"encoding/json"
	"net"
	"sync"

	"github.com/Icey-Glitch/Syncplay-G/ConMngr"
	"github.com/Icey-Glitch/Syncplay-G/utils"
)

type ReadyMessage struct {
	Set struct {
		Ready struct {
			Username          string `json:"username"`
			IsReady           bool   `json:"isReady"`
			ManuallyInitiated bool   `json:"manuallyInitiated"`
		} `json:"ready"`
	} `json:"Set"`
}

type ClientReadyMessage struct {
	Ready struct {
		IsReady           bool `json:"isReady"`
		ManuallyInitiated bool `json:"manuallyInitiated"`
	} `json:"ready"`
}

type UserReadyState struct {
	Username          string
	IsReady           bool
	ManuallyInitiated bool
}

type UserReadyMngr struct {
	Users []*UserReadyState
	mutex sync.Mutex
}

func SendReadyMessageInit(conn net.Conn, username string) {
	readyMessage := ReadyMessage{
		Set: struct {
			Ready struct {
				Username          string `json:"username"`
				IsReady           bool   `json:"isReady"`
				ManuallyInitiated bool   `json:"manuallyInitiated"`
			} `json:"ready"`
		}{
			Ready: struct {
				Username          string `json:"username"`
				IsReady           bool   `json:"isReady"`
				ManuallyInitiated bool   `json:"manuallyInitiated"`
			}{
				Username:          username,
				IsReady:           false,
				ManuallyInitiated: false,
			},
		},
	}

	// pretty print

	readyMessageBytes, _ := json.Marshal(readyMessage)
	utils.PrettyPrintJSON(utils.InsertSpaceAfterColons(readyMessageBytes))
	utils.SendJSONMessageMultiCast(readyMessage)
}

func HandleReadyMessage(ready map[string]interface{}, conn net.Conn) {
	readyMngr := UserReadyMngr{}
	/*client {
	      "ready": {
	        "isReady": true,
	        "manuallyInitiated": true
	      }
		}*/
	//server {"Set": {"ready": {"username": "car1", "isReady": true, "manuallyInitiated": true}}}

	// print the incoming message
	cm := ConMngr.GetConnectionManager()
	readyBytes, _ := json.Marshal(ready)

	utils.PrettyPrintJSON(utils.InsertSpaceAfterColons(readyBytes))

	// Unmarshal the incoming JSON data into ClientReadyMessage struct
	var clientReadyMessage ClientReadyMessage
	err := json.Unmarshal(readyBytes, &clientReadyMessage)
	if err != nil {
		utils.PrettyPrintJSON([]byte(`{"error": "Invalid JSON format"}`))
		return
	}

	// extract the isReady and manuallyInitiated
	isReady := clientReadyMessage.Ready.IsReady
	manuallyInitiated := clientReadyMessage.Ready.ManuallyInitiated

	// Assuming username is extracted from the connection or another source
	username := cm.GetUsername(conn)

	readyMngr.mutex.Lock()

	UserReadyState := UserReadyState{
		Username:          username,
		IsReady:           isReady,
		ManuallyInitiated: manuallyInitiated,
	}

	// check if user is already in the list
	userIndex := findUser(username)
	if userIndex != -1 {
		// update user
		readyMngr.Users[userIndex] = &UserReadyState
	} else {
		// add user
		readyMngr.Users = append(readyMngr.Users, &UserReadyState)
	}

	readyMngr.mutex.Unlock()

	// pretty print

	utils.PrettyPrintJSON(utils.InsertSpaceAfterColons(readyBytes))

	// build response
	response := ReadyMessage{
		Set: struct {
			Ready struct {
				Username          string `json:"username"`
				IsReady           bool   `json:"isReady"`
				ManuallyInitiated bool   `json:"manuallyInitiated"`
			} `json:"ready"`
		}{
			Ready: struct {
				Username          string `json:"username"`
				IsReady           bool   `json:"isReady"`
				ManuallyInitiated bool   `json:"manuallyInitiated"`
			}{
				Username:          username,
				IsReady:           isReady,
				ManuallyInitiated: manuallyInitiated,
			},
		},
	}

	utils.SendJSONMessageMultiCast(response)
}

func getReadyState(username string) (bool, bool) {
	readyMngr := UserReadyMngr{}
	readyMngr.mutex.Lock()
	defer readyMngr.mutex.Unlock()

	if readyMngr.Users == nil {
		return false, false
	}

	if len(readyMngr.Users) == 0 {
		return false, false
	}

	for _, user := range readyMngr.Users {
		if user.Username == username {
			return user.IsReady, user.ManuallyInitiated
		}
	}

	return false, false
}

func findUser(username string) int {
	readyMngr := UserReadyMngr{}
	readyMngr.mutex.Lock()
	defer readyMngr.mutex.Unlock()

	if readyMngr.Users == nil {
		return -1
	}

	if len(readyMngr.Users) == 0 {
		return -1
	}

	for i, user := range readyMngr.Users {
		if user.Username == username {
			return i
		}
	}

	return -1
}
