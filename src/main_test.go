package main

import (
	"bytes"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type MockConn struct {
	net.Conn
	readBuffer  *bytes.Buffer
	writeBuffer *bytes.Buffer
}

func (m *MockConn) Read(b []byte) (n int, err error) {
	return m.readBuffer.Read(b)
}

func (m *MockConn) Write(b []byte) (n int, err error) {
	return m.writeBuffer.Write(b)
}

func (m *MockConn) Close() error {
	return nil
}

func (m *MockConn) SetDeadline(t time.Time) error {
	return nil
}

func (m *MockConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (m *MockConn) SetWriteDeadline(t time.Time) error {
	return nil
}

func TestHandleClient(t *testing.T) {
	// Test Case 1: Handle a valid client connection
	t.Run("Valid Client Connection", func(t *testing.T) {
		mockConn := &MockConn{
			readBuffer:  bytes.NewBuffer([]byte(`{"Hello": {"username": "Bob", "password": "", "room": {"name": "SyncRoom"}, "version": "1.2.7"}}`)),
			writeBuffer: &bytes.Buffer{},
		}

		go handleClient(mockConn)

		time.Sleep(1 * time.Second) // Give some time for the goroutine to process

		assert.Contains(t, mockConn.writeBuffer.String(), `"Hello"`)
	})

	// Test Case 2: Handle a client connection that sends an EOF (disconnect)
	t.Run("Client Disconnect", func(t *testing.T) {
		mockConn := &MockConn{
			readBuffer:  bytes.NewBuffer(nil),
			writeBuffer: &bytes.Buffer{},
		}

		go handleClient(mockConn)

		time.Sleep(1 * time.Second) // Give some time for the goroutine to process

		assert.Contains(t, mockConn.writeBuffer.String(), "")
	})

	// Test Case 3: Handle a client connection that sends an invalid message
	t.Run("Invalid Message", func(t *testing.T) {
		mockConn := &MockConn{
			readBuffer:  bytes.NewBuffer([]byte(`{"Invalid": "message"}`)),
			writeBuffer: &bytes.Buffer{},
		}

		go handleClient(mockConn)

		time.Sleep(1 * time.Second) // Give some time for the goroutine to process

		assert.Contains(t, mockConn.writeBuffer.String(), "")
	})

	// Test Case 4: Handle a client connection that sends a valid message
	t.Run("Valid Message", func(t *testing.T) {
		mockConn := &MockConn{
			readBuffer:  bytes.NewBuffer([]byte(`{"State": {"playstate": {"position": 10, "paused": false}}}`)),
			writeBuffer: &bytes.Buffer{},
		}

		go handleClient(mockConn)

		time.Sleep(1 * time.Second) // Give some time for the goroutine to process

		assert.Contains(t, mockConn.writeBuffer.String(), "")
	})
}
