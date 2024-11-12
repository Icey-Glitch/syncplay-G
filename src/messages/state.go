package messages

import (
	"fmt"
	"net"
	"time"

	roomM "github.com/Icey-Glitch/Syncplay-G/mngr/room"
	"github.com/Icey-Glitch/Syncplay-G/utils"
)

type ClientStateMessage struct {
	State struct {
		// optional elements
		IgnoringOnTheFly *IgnoringOnTheFly `json:"ignoringOnTheFly,omitempty"`

		Playstate struct {
			Position float64     `json:"position"`
			Paused   bool        `json:"paused"`
			DoSeek   bool        `json:"doSeek"`
			SetBy    interface{} `json:"setBy"`
		} `json:"playstate"`
		Ping struct {
			LatencyCalculation       float64 `json:"latencyCalculation"`
			ClientLatencyCalculation float64 `json:"clientLatencyCalculation"`
			ClientRtt                float64 `json:"clientRtt"`
		} `json:"ping"`
	} `json:"State"`
}

type ServerStateMessage struct {
	State struct {
		// optional elements
		IgnoringOnTheFly *IgnoringOnTheFly `json:"ignoringOnTheFly,omitempty"`

		Playstate struct {
			Position float64     `json:"position"`
			Paused   bool        `json:"paused"`
			DoSeek   bool        `json:"doSeek"`
			SetBy    interface{} `json:"setBy"`
		} `json:"playstate"`
		Ping struct {
			LatencyCalculation       float64 `json:"latencyCalculation"`
			ClientLatencyCalculation float64 `json:"clientLatencyCalculation"`
			ServerRtt                float64 `json:"serverRtt"`
		} `json:"ping"`
	} `json:"State"`
}

type UserState struct {
	Position float64
	Paused   bool
	DoSeek   bool
	SetBy    interface{}
}

type IgnoringOnTheFly struct {
	Client float64 `json:"client,omitempty"`
	Server float64 `json:"server,omitempty"`
}

func SendInitialState(connection roomM.Connection) {
	if connection.Owner == nil {
		return
	}

	if len(connection.Owner.Users) <= 1 {
		stateMessage := ServerStateMessage{}
		stateMessage.State.Ping.LatencyCalculation = float64(time.Now().UnixNano()) / 1e9
		stateMessage.State.Ping.ServerRtt = 0
		stateMessage.State.Playstate.DoSeek = false
		stateMessage.State.Playstate.Position = 0
		stateMessage.State.Playstate.Paused = true
		stateMessage.State.Playstate.SetBy = "Nobody"

		err := utils.SendJSONMessage(connection.Conn, stateMessage)
		if err != nil {
			fmt.Println("Error sending initial state message:", err)
			return
		}
		return
	}

	roomState := connection.Owner.PlaylistManager.Playlist
	stateMessage := ServerStateMessage{}
	stateMessage.State.Ping.LatencyCalculation = float64(time.Now().UnixNano()) / 1e9
	stateMessage.State.Ping.ServerRtt = 0

	stateMessage.State.Playstate.DoSeek = false
	stateMessage.State.Playstate.Position = roomState.Position
	stateMessage.State.Playstate.Paused = roomState.Paused

	err := utils.SendJSONMessage(connection.Conn, stateMessage)
	if err != nil {
		fmt.Println("Error sending initial state message:", err)
		return
	}
}

func SendUserState(connection roomM.Connection) bool {
	//fmt.Println("Sending user state")

	latencyCalculation, err := connection.Owner.GetUsersLatencyCalculation(&connection)
	if err != nil {
		fmt.Println("Error getting user latency calculation:", err)
		return true
	}

	processingTime := float64(time.Now().UnixNano())/1e9 - latencyCalculation.ArivalTime

	yourLatency := latencyCalculation.ClientTime - connection.Owner.PlaylistManager.GetLastMessageAge(connection.Username)

	Ignore := connection.Owner.PlaylistManager.Playlist.Ignore

	err = sendStateMessage(connection.Owner, connection.Conn, connection.Owner.PlaylistManager.Playlist.Position, connection.Owner.PlaylistManager.Playlist.Paused, connection.Owner.PlaylistManager.Playlist.DoSeek, processingTime, connection.Owner.PlaylistManager.Playlist.SetBy, latencyCalculation.ClientTime, connection.Username, yourLatency, Ignore)
	if err != nil {
		fmt.Println("Error sending state message:", err)
		return true
	}

	// SetLastMessageAge
	connection.Owner.PlaylistManager.SetLastMessageAge(connection.Username, float64(time.Now().UnixNano())/1e9)

	return false
}

// SendGlobalState sends the global state to all users
func SendGlobalState(connection roomM.Connection) {
	if connection.Owner == nil {
		return
	}

	for _, user := range connection.Owner.Users {
		if user.Conn != nil {
			SendUserState(*user)
		}
	}

	connection.Owner.PlaylistManager.SetIgnoreInt(0)
}

func sendStateMessage(room *roomM.Room, conn net.Conn, position float64, paused bool, doSeek bool, processingTime float64, stateChange string, clientTime float64, usr string, LastMsgAge float64, Ignore float64) error {
	if room == nil {
		return fmt.Errorf("room cannot be nil")
	}

	if conn == nil {
		return fmt.Errorf("connection cannot be nil")
	}

	stateMessage := ServerStateMessage{}
	stateMessage.State.Ping.LatencyCalculation = float64(time.Now().UnixNano()) / 1e9
	// if clientTime != 0 {
	// 	stateMessage.State.Ping.SenderLatency = clientTime + processingTime
	// }
	// stateMessage.State.Ping.YourLatency = LastMsgAge

	stateMessage.State.Playstate.Position = position
	stateMessage.State.Playstate.Paused = paused
	stateMessage.State.Playstate.DoSeek = doSeek
	if stateChange == "" {
		stateChange = "Nobody"
	}
	stateMessage.State.Playstate.SetBy = stateChange

	if Ignore != 0 {
		stateMessage.State.IgnoringOnTheFly = &IgnoringOnTheFly{}
		stateMessage.State.IgnoringOnTheFly.Client = Ignore
		stateMessage.State.IgnoringOnTheFly.Server = Ignore
	}

	err := utils.SendJSONMessage(conn, stateMessage)
	if err != nil {
		return fmt.Errorf("error sending JSON message: %w", err)
	}

	return nil
}

func HandleStatePing(ping map[string]interface{}) (clientRTT float64, latencyCalculation float64, clientLatencyCalculation float64, err error) {
	/*
				General client ping (no file open / paused at beginning)

				Client >> {"State": {"ping": {"clientRtt": 0, "clientLatencyCalculation": 1394590473.585, "latencyCalculation": 1394590688.962084}, "playstate": {"paused": true, "position": 0.0}}}

				General client ping (file playing)

				Client >> {"State": {"ping": {"clientRtt": 0, "clientLatencyCalculation": 1394590473.585, "latencyCalculation": 1394590688.962084}, "playstate": {"paused": false, "position": 2.236}}}

				calculate the time the message was sent using the client's latency calculation and the server's time

				Look at the code to see how it works. ‘[client]LatencyCalculation’  is a timestamp based on the time in seconds since the epoch as a floating point number. ‘clientRtt’ is round-trip time (i.e. ping time). In older versions Syncplay used ‘yourLatency’ and ‘senderLatency’ but not ‘serverRtt’.

			{
		  "State": {
		    "ping": {
		      "clientRtt": 0,
		      "clientLatencyCalculation": 1394590473.585,
		      "latencyCalculation": 1394590688.962084
		    },
		    "playstate": {
		      "paused": true,
		      "position": 0
		    }
		  }
		}
	*/

	latencyCalculation, ok := ping["latencyCalculation"].(float64)
	if !ok {
		return 0, 0, 0, fmt.Errorf("invalid latencyCalculation")
	}

	clientRTT, ok = ping["clientRtt"].(float64)
	if !ok {
		return 0, 0, 0, fmt.Errorf("invalid clientRtt")
	}

	clientLatencyCalculation, ok = ping["clientLatencyCalculation"].(float64)
	if !ok {
		return 0, 0, 0, fmt.Errorf("invalid clientLatencyCalculation")
	}

	// playstate logic

	return clientRTT, latencyCalculation, clientLatencyCalculation, nil
}

var globalState = struct {
	position float64
	paused   bool
	doSeek   bool
	setBy    interface{}
}{}

func UpdateGlobalState(connection roomM.Connection, position, paused, doSeek, setBy interface{}, messageAge float64, latencyCalculation float64, Ignore float64) error {

	room := connection.Owner

	globalState.position = position.(float64)
	globalState.paused = paused.(bool)
	globalState.doSeek = doSeek.(bool)
	globalState.setBy = setBy

	// store IgnoreOnTheFly
	if Ignore != 0 {
		room.PlaylistManager.SetIgnoreInt(Ignore)
		setBy = connection.Username
		err := room.PlaylistManager.SetUserPlaystate(connection.Username, position.(float64), paused.(bool), doSeek.(bool), setBy.(string), messageAge, true)
		if err != nil {
			return fmt.Errorf("error storing user playstate: %w", err)
		}
	} else {
		err := room.PlaylistManager.SetUserPlaystate(connection.Username, position.(float64), paused.(bool), doSeek.(bool), setBy.(string), messageAge, false)
		if err != nil {
			return fmt.Errorf("error storing user playstate: %w", err)
		}
	}

	return nil
}

func GetLocalState() (interface{}, interface{}, interface{}, interface{}) {
	return globalState.position, globalState.paused, globalState.doSeek, globalState.setBy
}

type UserMessage struct {
	Set struct {
		User struct {
			Username string `json:"username"`
			IsOwner  bool   `json:"isOwner"`
		} `json:"user"`
	} `json:"Set"`
}

func HandleUserMessage(value interface{}, conn net.Conn) {
	user, ok := value.(map[string]interface{})
	if !ok || user == nil {
		fmt.Println("Error: user is nil or not a map")
		return
	}
	HandleJoinMessage(conn, user)
}
