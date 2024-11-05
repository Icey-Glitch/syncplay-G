package ready

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewReadyManager(t *testing.T) {
	rm := NewReadyManager()

	assert.NotNil(t, rm)
	assert.NotNil(t, rm.readyStates)
	assert.NotNil(t, rm.stateEvent)
}

func TestSetUserReadyState(t *testing.T) {
	rm := NewReadyManager()
	username := "testUser"

	rm.SetUserReadyState(username, true, true)

	state, exists := rm.GetUserReadyState(username)
	assert.True(t, exists)
	assert.Equal(t, username, state.Username)
	assert.Equal(t, true, state.IsReady)
	assert.Equal(t, true, state.ManuallyInitiated)
}

func TestGetUserReadyState(t *testing.T) {
	rm := NewReadyManager()
	username := "testUser"
	rm.SetUserReadyState(username, true, true)

	state, exists := rm.GetUserReadyState(username)
	assert.True(t, exists)
	assert.Equal(t, username, state.Username)
	assert.Equal(t, true, state.IsReady)
	assert.Equal(t, true, state.ManuallyInitiated)

	_, exists = rm.GetUserReadyState("nonExistentUser")
	assert.False(t, exists)
}

func TestRemoveUserReadyState(t *testing.T) {
	rm := NewReadyManager()
	username := "testUser"
	rm.SetUserReadyState(username, true, true)
	rm.RemoveUserReadyState(username)

	_, exists := rm.GetUserReadyState(username)
	assert.False(t, exists)
}

func TestSubscribeToStateChanges(t *testing.T) {
	rm := NewReadyManager()
	ch := rm.SubscribeToStateChanges()

	assert.NotNil(t, ch)
}

func TestUnsubscribeFromStateChanges(t *testing.T) {
	rm := NewReadyManager()
	ch := rm.SubscribeToStateChanges()

	rm.UnsubscribeFromStateChanges(ch)
	// No direct way to test unsubscription, but we can ensure no panic occurs
}
