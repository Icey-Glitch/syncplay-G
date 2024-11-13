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
	serverAddr             = "localhost:8080"
	maxClients             = 50000                 // Maximum number of clients to simulate
	numofRooms             = 100                   // Number of rooms to simulate
	connectInterval        = 1 * time.Microsecond  // Interval between client connections
	clientDuration         = 30 * time.Second      // Duration to keep the client connection open
	stateInterval          = 5 * time.Second       // Interval between state messages
	maxFiles               = 5                     // Maximum number of files per room
	responseDeadline       = 10 * time.Millisecond // Response deadline for AB testing
	responseWindowSize     = 1000                  // Number of responses to consider
	maxSlowResponsePercent = 5.0                   // Maximum acceptable percentage of slow responses
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

type FileMessage struct {
	Set struct {
		File struct {
			Duration float64     `json:"duration"`
			Name     string      `json:"name"`
			Size     interface{} `json:"size"`
		} `json:"file"`
	} `json:"Set"`
}

func generateRandomFileName() string {
	return fmt.Sprintf("file_%d", rand.Intn(100000))
}

func simulateClient(id int, wg *sync.WaitGroup, responseTimes chan<- time.Duration) {
	defer wg.Done()

	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		fmt.Printf("Client %d: Error connecting to server: %v\n", id, err)
		return
	}
	defer conn.Close()

	// Send Hello message
	username := fmt.Sprintf("user%d", id)
	room := fmt.Sprintf("room%d", rand.Intn(numofRooms))
	helloMessage := fmt.Sprintf(`{"Hello": {"username": "%s", "version": "1.2.7", "room": {"name": "%s"}}}`+"\r\n", username, room)
	_, err = conn.Write([]byte(helloMessage))
	if err != nil {
		fmt.Printf("Client %d: Error sending Hello message: %v\n", id, err)
		return
	}

	// Start reading server responses
	go func() {
		reader := bufio.NewReader(conn)
		for {
			start := time.Now()
			line, err := reader.ReadBytes('\n')
			if err != nil {
				return
			}
			duration := time.Since(start)
			responseTimes <- duration
			// Handle server messages if necessary
			var fileMsg FileMessage
			if err := json.Unmarshal(line, &fileMsg); err == nil {
				// Process file message
			}
		}
	}()

	// Simulate client actions
	startTime := time.Now()
	paused := false
	position := 0.0

	ticker := time.NewTicker(stateInterval)
	defer ticker.Stop()

	for time.Since(startTime) < clientDuration {
		select {
		case <-ticker.C:
			// Randomly decide to play, pause, or seek
			action := rand.Intn(3)
			switch action {
			case 0:
				paused = !paused
			case 1:
				position = rand.Float64() * 100
			case 2:
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

			// Randomly add a file to the room
			if rand.Intn(10) < 2 {
				fileMsg := FileMessage{}
				fileMsg.Set.File.Duration = rand.Float64() * 1000
				fileMsg.Set.File.Name = generateRandomFileName()
				fileMsg.Set.File.Size = rand.Intn(1000000)

				fileData, _ := json.Marshal(fileMsg)
				_, err = conn.Write(append(fileData, '\r', '\n'))
				if err != nil {
					fmt.Printf("Client %d: Error sending File message: %v\n", id, err)
					return
				}
			}
		}
	}
}

func testMaxConcurrentConnections() int {
	var wg sync.WaitGroup
	responseTimes := make(chan time.Duration, 100000)
	defer close(responseTimes)

	var totalResponses int
	var slowResponses int
	var responseTimeWindow []time.Duration
	var maxConcurrentConnections int
	concurrentClients := 0
	var mutex sync.Mutex

	done := make(chan struct{})
	go func() {
		for responseTime := range responseTimes {
			mutex.Lock()
			totalResponses++
			if responseTime > responseDeadline {
				slowResponses++
			}
			responseTimeWindow = append(responseTimeWindow, responseTime)
			if len(responseTimeWindow) > responseWindowSize {
				oldest := responseTimeWindow[0]
				responseTimeWindow = responseTimeWindow[1:]
				if oldest > responseDeadline {
					slowResponses--
				}
				totalResponses--
			}
			mutex.Unlock()
		}
		close(done)
	}()

	for i := 0; i < maxClients; i++ {
		wg.Add(1)
		go simulateClient(i, &wg, responseTimes)
		concurrentClients++
		time.Sleep(connectInterval)

		mutex.Lock()
		if len(responseTimeWindow) >= responseWindowSize {
			slowResponsePercent := float64(slowResponses) / float64(len(responseTimeWindow)) * 100.0
			if slowResponsePercent > maxSlowResponsePercent {
				fmt.Printf("Exceeded response time threshold at %d concurrent connections\n", concurrentClients)
				maxConcurrentConnections = concurrentClients - 1
				mutex.Unlock()
				break
			}
		}
		mutex.Unlock()
	}

	wg.Wait()
	close(responseTimes)
	<-done

	if maxConcurrentConnections == 0 {
		maxConcurrentConnections = concurrentClients
	}

	return maxConcurrentConnections
}

func main() {
	maxConcurrentConnections := testMaxConcurrentConnections()
	fmt.Printf("Maximum concurrent connections before exceeding response threshold: %d\n", maxConcurrentConnections)
}
