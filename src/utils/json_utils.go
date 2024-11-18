package utils

import (
	"fmt"
	"net"
	"os"
	"sync"
	"syscall"

	RoomM "github.com/Icey-Glitch/Syncplay-G/mngr/room"
	"github.com/goccy/go-json"
)

// Map to store mutexes per connection
var connMutexes sync.Map

// SendData writes data to the connection ensuring thread safety
func SendData(conn net.Conn, data []byte) error {
	if conn == nil {
		return fmt.Errorf("connection is nil")
	}

	// Get or create mutex for the connection
	mutexInterface, _ := connMutexes.LoadOrStore(conn, &sync.Mutex{})
	mutex := mutexInterface.(*sync.Mutex)

	mutex.Lock()
	defer mutex.Unlock()

	// Check if the connection is still open
	if err := checkConnection(conn); err != nil {
		return fmt.Errorf("connection is closed: %w", err)
	}

	_, err := conn.Write(data)
	if err != nil {
		if isBrokenPipe(err) {
			//return fmt.Errorf("broken pipe: %w", err)
			return fmt.Errorf("broken pipe")
		}
		return fmt.Errorf("error writing data to connection: %w", err)
	}

	return nil
}

// checkConnection checks if the connection is still open
func checkConnection(conn net.Conn) error {
	// Use syscall to check the connection status
	var sysErr error
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		file, err := tcpConn.File()
		if err != nil {
			return err
		}
		defer file.Close()
		_, sysErr = syscall.GetsockoptInt(int(file.Fd()), syscall.SOL_SOCKET, syscall.SO_ERROR)
	}
	return sysErr
}

// isBrokenPipe checks if the error is a broken pipe error
func isBrokenPipe(err error) bool {
	if opErr, ok := err.(*net.OpError); ok {
		if sysErr, ok := opErr.Err.(*os.SyscallError); ok {
			return sysErr.Err == syscall.EPIPE
		}
	}
	return false
}

// SendJSONMessageMultiCast sends a JSON message to all users in a room
func SendJSONMessageMultiCast(message interface{}, room *RoomM.Room) {

	// Marshal the message once
	data, err := json.Marshal(message)
	if err != nil {
		fmt.Printf("Error marshalling JSON message: %v\n", err)
		return
	}

	// Append CRLF
	data = append(data, '\r', '\n')

	if room.Users == nil {
		fmt.Println("Room users is nil")
		return
	}

	for _, user := range room.Users {
		// check if range is nil

		// prevent nil pointer dereference
		if user == nil {
			// Skip users without a connection
			continue
		}
		// prevent broken pipe error
		if user.Conn == nil {
			continue
		}

		err := SendData(user.Conn, data)
		if err != nil {
			if err.Error() == "broken pipe" {
				//room.RemoveConnection(user.Conn)
				//fmt.Printf("User %s disconnected\n", user.Username)
				continue
			}
			//fmt.Printf("Error sending data to user %s: %v\n", user.Username, err)
			// Handle error as needed
		}
	}
}

// SendJSONMessage marshals the message and sends it to the connection
func SendJSONMessage(conn net.Conn, message interface{}) error {
	if conn == nil {
		return fmt.Errorf("connection is nil")
	}

	// Marshal the message
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("error marshalling JSON message: %w", err)
	}

	// Append CRLF
	data = append(data, '\r', '\n')

	return SendData(conn, data)
}
