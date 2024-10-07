package ready

import (
	"sync"
)

type ReadyState struct {
	Username          string `json:"username"`
	IsReady           bool   `json:"isReady"`
	ManuallyInitiated bool   `json:"manuallyInitiated"`
}

type ReadyManager struct {
	readyStates map[string]ReadyState
	mutex       sync.RWMutex
}

func NewReadyManager() *ReadyManager {
	return &ReadyManager{
		readyStates: make(map[string]ReadyState),
	}
}

func (rm *ReadyManager) SetUserReadyState(username string, isReady, manuallyInitiated bool) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	rm.readyStates[username] = ReadyState{
		Username:          username,
		IsReady:           isReady,
		ManuallyInitiated: manuallyInitiated,
	}
}

func (rm *ReadyManager) GetUserReadyState(username string) (ReadyState, bool) {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	state, exists := rm.readyStates[username]
	return state, exists
}

func (rm *ReadyManager) RemoveUserReadyState(username string) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	delete(rm.readyStates, username)
}
