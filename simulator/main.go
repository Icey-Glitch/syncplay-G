package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"sync"
	"time"
)

const (
	serverAddr      = "localhost:8080"
	numClients      = 10000                // Number of clients to simulate
	numofRooms      = 100                  // Number of rooms to simulate
	connectInterval = 1 * time.Microsecond // Interval between client connections
	clientDuration  = 30 * time.Second     // Duration to keep the client connection open
	stateInterval   = 5 * time.Second      // Interval between state messages
)

type StateMessage struct {
	State struct {
		Ping struct {
			ClientRtt                float64 `json:"clientRtt"`
			ClientLatencyCalculation float64 `json:"clientLatencyCalculation"`
			LatencyCalculation       float64 `json:"latencyCalculation"`
		} `json:"ping"`
		Playstate struct {
			Paused   bool    `json:"paused"`
			Position float64 `json:"position"`
			DoSeek   bool    `json:"doSeek,omitempty"`
			SetBy    string  `json:"setBy,omitempty"`
		} `json:"playstate"`
	} `json:"State"`
}

func simulateClient(id int, wg *sync.WaitGroup) {
	defer wg.Done()

	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		fmt.Printf("Client %d: Error connecting to server: %v\n", id, err)
		return
	}
	defer conn.Close()

	// Send Hello message
	username := fmt.Sprintf("user%d", id)
	// Randomly select a room
	room := fmt.Sprintf("room%d", rand.Intn(numofRooms))
	helloMessage := fmt.Sprintf(`{"Hello": {"username": "%s", "room": {"name": "%s"}}}`+"\r\n", room, username)
	_, err = conn.Write([]byte(helloMessage))
	if err != nil {
		fmt.Printf("Client %d: Error sending Hello message: %v\n", id, err)
		return
	}

	// Start reading server responses
	go func() {
		reader := bufio.NewReader(conn)
		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				return
			}
			// Handle server messages if necessary
			_ = line // Placeholder for processing the server response
		}
	}()

	// Simulate client actions
	startTime := time.Now()
	position := 0.0
	paused := false
	ticker := time.NewTicker(stateInterval)
	defer ticker.Stop()

	for time.Since(startTime) < clientDuration {
		select {
		case <-ticker.C:
			// Randomly decide to play, pause, or seek
			action := rand.Intn(3)
			switch action {
			case 0: // Play/Pause
				paused = !paused
			case 1: // Seek
				position = rand.Float64() * 100
			case 2: // Continue playing
				if !paused {
					position += stateInterval.Seconds()
				}
			}

			stateMsg := StateMessage{}
			stateMsg.State.Ping.ClientRtt = 0
			stateMsg.State.Ping.ClientLatencyCalculation = float64(time.Now().UnixNano()) / 1e9
			stateMsg.State.Ping.LatencyCalculation = stateMsg.State.Ping.ClientLatencyCalculation
			stateMsg.State.Playstate.Paused = paused
			stateMsg.State.Playstate.Position = position
			if action == 1 {
				stateMsg.State.Playstate.DoSeek = true
			}

			stateData, _ := json.Marshal(stateMsg)
			_, err = conn.Write(append(stateData, '\r', '\n'))
			if err != nil {
				fmt.Printf("Client %d: Error sending State message: %v\n", id, err)
				return
			}
		}
	}
}

func main() {
	var wg sync.WaitGroup

	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go simulateClient(i, &wg)
		time.Sleep(connectInterval)
	}

	wg.Wait()
}
