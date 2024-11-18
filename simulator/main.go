package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/goccy/go-json"
)

const (
	serverAddr             = "localhost:8080"
	maxClients             = 500000                // Maximum number of clients to simulate
	numofRooms             = 100                   // Number of rooms to simulate
	clientDuration         = 30 * time.Second      // Duration to keep the client connection open
	stateInterval          = 5 * time.Second       // Interval between state messages
	responseDeadline       = 10 * time.Millisecond // Response deadline for AB testing
	responseWindowSize     = 1000                  // Number of responses to consider
	maxSlowResponsePercent = 5.0                   // Maximum acceptable percentage of slow responses
	poolSize               = 1000                  // Size of the connection pool
	workerPoolSize         = 100                   // Size of the worker pool
)

var (
	helloMessages = make([]string, maxClients)
	stateMessages = make([]string, 3)
	fileMessages  = make([]string, 10)
	connPool      = make(chan net.Conn, poolSize)
)

func init() {
	var wg sync.WaitGroup

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

	// Initialize connection pool
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < poolSize; i++ {
			conn, err := net.Dial("tcp", serverAddr)
			if err != nil {
				continue
			}
			connPool <- conn
		}
	}()

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

	conn := <-connPool
	defer func() {
		connPool <- conn
	}()

	_, err := conn.Write([]byte(helloMessages[id]))
	if err != nil {
		return
	}

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
			var fileMsg FileMessage
			if err := json.Unmarshal(line, &fileMsg); err == nil {
				// Process file message
			}
		}
	}()

	ticker := time.NewTicker(stateInterval)
	defer ticker.Stop()

	clientDone := time.After(clientDuration)

	for {
		select {
		case <-ticker.C:
			action := rand.Intn(3)
			_, err = conn.Write([]byte(stateMessages[action]))
			if err != nil {
				return
			}
			if rand.Intn(10) < 2 {
				_, err = conn.Write([]byte(fileMessages[rand.Intn(10)]))
				if err != nil {
					return
				}
			}
		case <-clientDone:
			return
		}
	}
}

func worker(id int, jobs <-chan int, wg *sync.WaitGroup, responseTimes chan<- time.Duration) {
	defer wg.Done()
	for job := range jobs {
		simulateClient(job, wg, responseTimes)
	}
}

func testMaxConcurrentConnections() int {
	var wg sync.WaitGroup
	responseTimes := make(chan time.Duration, 100000)

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
				atomic.AddUint64(&slowResponses, ^uint64(0))
			}
		}
		close(done)
	}()

	jobs := make(chan int, maxClients)
	for w := 0; w < workerPoolSize; w++ {
		wg.Add(1)
		go worker(w, jobs, &wg, responseTimes)
	}

	concurrentClients := 0
	maxConcurrentConnections := 0
	for i := 0; i < maxClients; i++ {
		jobs <- i
		concurrentClients++

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
	close(jobs)

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
