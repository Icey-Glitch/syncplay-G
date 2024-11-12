package ready

import (
	"sync"

	"github.com/Icey-Glitch/Syncplay-G/mngr/event"
)

/*
TODO:

- Implement a way to automatically set all users to ready when the room is paused / unpaused more efficiently

*/

type ReadyState struct {
	Username          string
	IsReady           bool
	ManuallyInitiated bool
}

type ReadyManager struct {
	readyStates map[string]ReadyState
	mutex       sync.RWMutex
	stateEvent  *event.Event
}

func NewReadyManager() *ReadyManager {
	return &ReadyManager{
		readyStates: make(map[string]ReadyState),
		stateEvent:  event.NewEvent(),
	}
}

func (rm *ReadyManager) SetUserReadyState(username string, isReady bool, manuallyInitiated bool) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	rm.readyStates[username] = ReadyState{
		Username:          username,
		IsReady:           isReady,
		ManuallyInitiated: manuallyInitiated,
	}

	rm.stateEvent.Publish(rm.readyStates[username])
}

func (rm *ReadyManager) GetUserReadyState(username string) (ReadyState, bool) {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	state, exists := rm.readyStates[username]
	return state, exists
}

func (rm *ReadyManager) GetReadyStates() map[string]ReadyState {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	return rm.readyStates
}

func (rm *ReadyManager) RemoveUserReadyState(username string) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	delete(rm.readyStates, username)
	rm.stateEvent.Publish(username)
}

func (rm *ReadyManager) SubscribeToStateChanges() chan interface{} {
	return rm.stateEvent.Subscribe()
}

func (rm *ReadyManager) UnsubscribeFromStateChanges(ch chan interface{}) {
	rm.stateEvent.Unsubscribe(ch)
}
