package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"sync"
	"sync/atomic"
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

var (
	helloMessages = make([]string, maxClients)
	stateMessages = make([]string, 3)
	fileMessages  = make([]string, 10)
)

func init() {
	var wg sync.WaitGroup

	rand.Seed(time.Now().UnixNano())

	// Precompute hello messages
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < maxClients; i++ {
			username := fmt.Sprintf("user%d", i)
			room := fmt.Sprintf("room%d", rand.Intn(numofRooms))
			helloMessages[i] = fmt.Sprintf(`{"Hello": {"username": "%s", "version": "1.2.7", "room": {"name": "%s"}}}`+"\r\n", username, room)
		}
	}()

	// Precompute state messages
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 3; i++ {
			stateMsg := StateMessage{}
			stateMsg.State.Ping.ClientRtt = 0
			stateMsg.State.Ping.ClientLatencyCalculation = float64(time.Now().UnixNano()) / 1e9
			stateMsg.State.Ping.LatencyCalculation = stateMsg.State.Ping.ClientLatencyCalculation
			stateMsg.State.Playstate.Paused = i == 0
			stateMsg.State.Playstate.Position = float64(i * 100)
			if i == 1 {
				stateMsg.State.Playstate.DoSeek = true
			}
			stateData, _ := json.Marshal(stateMsg)
			stateMessages[i] = string(stateData) + "\r\n"
		}
	}()

	// Precompute file messages
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			fileMsg := FileMessage{}
			fileMsg.Set.File.Duration = rand.Float64() * 1000
			fileMsg.Set.File.Name = generateRandomFileName()
			fileMsg.Set.File.Size = rand.Intn(1000000)
			fileData, _ := json.Marshal(fileMsg)
			fileMessages[i] = string(fileData) + "\r\n"
		}
	}()

	// Wait for all precomputations to complete
	wg.Wait()
}

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
		// fmt.Printf("Client %d: Error connecting to server: %v\n", id, err)
		return
	}
	defer conn.Close()

	// Send Hello message
	_, err = conn.Write([]byte(helloMessages[id]))
	if err != nil {
		// fmt.Printf("Client %d: Error sending Hello message: %v\n", id, err)
		return
	}

	// Start reading server responses in the same goroutine
	reader := bufio.NewReader(conn)
	go func() {
		for {
			start := time.Now()
			line, err := reader.ReadBytes('\n')
			if err != nil {
				return
			}
			duration := time.Since(start)
			select {
			case responseTimes <- duration:
			default:
			}
			// Handle server messages if necessary
			var fileMsg FileMessage
			if err := json.Unmarshal(line, &fileMsg); err == nil {
				// Process file message
			}
		}
	}()

	// Simulate client actions
	startTime := time.Now()
	ticker := time.NewTicker(stateInterval)
	defer ticker.Stop()

	for time.Since(startTime) < clientDuration {
		select {
		case <-ticker.C:
			// Randomly decide to play, pause, or seek
			action := rand.Intn(3)
			_, err = conn.Write([]byte(stateMessages[action]))
			if err != nil {
				return
			}

			// Randomly add a file to the room
			if rand.Intn(10) < 2 {
				_, err = conn.Write([]byte(fileMessages[rand.Intn(10)]))
				if err != nil {
					return
				}
			}
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func testMaxConcurrentConnections() int {
	var wg sync.WaitGroup
	responseTimes := make(chan time.Duration, 100000)
	defer close(responseTimes)

	var totalResponses uint64
	var slowResponses uint64
	responseTimeWindow := make([]time.Duration, responseWindowSize)
	var index uint64

	done := make(chan struct{})
	go func() {
		for responseTime := range responseTimes {
			atomic.AddUint64(&totalResponses, 1)
			if responseTime > responseDeadline {
				atomic.AddUint64(&slowResponses, 1)
			}

			i := atomic.AddUint64(&index, 1) % uint64(responseWindowSize)
			oldResponseTime := responseTimeWindow[i]
			responseTimeWindow[i] = responseTime

			if oldResponseTime > responseDeadline {
				atomic.AddUint64(&slowResponses, ^uint64(0)) // Decrement
			}
		}
		close(done)
	}()

	concurrentClients := 0
	maxConcurrentConnections := 0
	for i := 0; i < maxClients; i++ {
		wg.Add(1)
		go simulateClient(i, &wg, responseTimes)
		concurrentClients++
		time.Sleep(connectInterval)

		if atomic.LoadUint64(&totalResponses) >= uint64(responseWindowSize) {
			slowResp := atomic.LoadUint64(&slowResponses)
			slowResponsePercent := float64(slowResp) / float64(responseWindowSize) * 100.0
			if slowResponsePercent > maxSlowResponsePercent {
				fmt.Printf("Exceeded response time threshold at %d concurrent connections\n", concurrentClients)
				maxConcurrentConnections = concurrentClients - 1
				break
			}
		}
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
