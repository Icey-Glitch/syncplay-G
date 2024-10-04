package ConMngr

import (
	"net"
	"sync"
)

// Connection represents a connection with username, state object, and conn.
type Connection struct {
	Username string
	State    interface{}
	Conn     net.Conn
}

// ConnectionManager manages a list of connections.
type ConnectionManager struct {
	connections []*Connection
	mutex       sync.RWMutex
}

// AddConnection adds a new connection to the manager.
func (cm *ConnectionManager) AddConnection(username string, state interface{}, conn net.Conn) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	connection := &Connection{
		Username: username,
		State:    state,
		Conn:     conn,
	}

	cm.connections = append(cm.connections, connection)
}

// ModifyConnection modifies an existing connection in the manager.
func (cm *ConnectionManager) ModifyConnection(username string, state interface{}, conn net.Conn) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	for _, connection := range cm.connections {
		if connection.Username == username {
			connection.State = state
			connection.Conn = conn
			break
		}
	}
}

// RemoveConnection removes a connection from the manager.
func (cm *ConnectionManager) RemoveConnection(conn net.Conn) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	for i, connection := range cm.connections {
		if connection.Conn == conn {
			cm.connections = append(cm.connections[:i], cm.connections[i+1:]...)
			break
		}
	}
}

// GetConnections returns a copy of the current connections.
func (cm *ConnectionManager) GetConnections() []*Connection {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	connections := make([]*Connection, len(cm.connections))
	copy(connections, cm.connections)

	return connections
}

// GetConnection returns a connection by username.
func (cm *ConnectionManager) GetConnection(username string) *Connection {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	for _, connection := range cm.connections {
		if connection.Username == username {
			return connection
		}
	}

	return nil
}

// GetConnectionByConn returns a connection by net.Conn.
func (cm *ConnectionManager) GetConnectionByConn(conn net.Conn) *Connection {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	for _, connection := range cm.connections {
		if connection.Conn == conn {
			return connection
		}
	}

	return nil
}

// GetUsername returns the username associated with a net.Conn.
func (cm *ConnectionManager) GetUsername(conn net.Conn) string {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	for _, connection := range cm.connections {
		if connection.Conn == conn {
			return connection.Username
		}
	}

	return ""
}

// GetConnectionManager returns the singleton instance of ConnectionManager.
func GetConnectionManager() *ConnectionManager {
	if connectionManager == nil {
		connectionManager = &ConnectionManager{}
	}
	return connectionManager
}

// connectionManager is a global variable that holds the singleton instance of ConnectionManager.
var connectionManager *ConnectionManager
