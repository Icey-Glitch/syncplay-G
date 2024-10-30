package playlists

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetUsers(t *testing.T) {
	pm := NewPlaylistManager()

	// Test case 1: Empty playlist
	users, exists := pm.GetUsers()
	assert.False(t, exists)
	assert.Empty(t, users)

	// Test case 2: Non-empty playlist
	username1 := "user1"
	username2 := "user2"

	err := pm.CreateUserPlaystate(username1)
	assert.NoError(t, err)

	err = pm.CreateUserPlaystate(username2)
	assert.NoError(t, err)

	users, exists = pm.GetUsers()
	assert.True(t, exists)
	assert.Len(t, users, 2)
	assert.Contains(t, users, username1)
	assert.Contains(t, users, username2)
}

func TestCreateUserPlaystate(t *testing.T) {
	pm := NewPlaylistManager()

	// Test case 1: Empty username
	err := pm.CreateUserPlaystate("")
	assert.Error(t, err)

	// Test case 2: Non-empty username
	username := "testUser"

	err = pm.CreateUserPlaystate(username)
	assert.NoError(t, err)

	users, exists := pm.GetUsers()
	assert.True(t, exists)
	assert.Len(t, users, 1)
	assert.Contains(t, users, username)
}

func TestRemoveUserPlaystate(t *testing.T) {
	pm := NewPlaylistManager()

	// Test case 1: Non-existent user
	err := pm.RemoveUserPlaystate("nonExistentUser")
	assert.Error(t, err)

	// Test case 2: Existing user
	username := "testUser"

	err = pm.CreateUserPlaystate(username)
	assert.NoError(t, err)

	err = pm.RemoveUserPlaystate(username)
	assert.NoError(t, err)

	users, exists := pm.GetUsers()
	assert.False(t, exists)
	assert.Empty(t, users)
}

func TestSetUserPlaystate(t *testing.T) {
	pm := NewPlaylistManager()

	// Test case 1: Empty username
	err := pm.SetUserPlaystate("", 0, false, false, "", 0)
	assert.Error(t, err)

	// Test case 2: Non-existent user
	username := "testUser"

	err = pm.SetUserPlaystate(username, 0, false, false, "", 0)
	assert.Error(t, err)

	// Test case 3: Existing user
	err = pm.CreateUserPlaystate(username)
	assert.NoError(t, err)

	err = pm.SetUserPlaystate(username, 0, false, false, "", 0)
	assert.NoError(t, err)
}

func TestSetUsersDoSeek(t *testing.T) {
	pm := NewPlaylistManager()

	// Test case 1: Set doSeek to true
	err := pm.SetUsersDoSeek(true, 0)
	assert.NoError(t, err)

	// Test case 2: Set doSeek to false
	err = pm.SetUsersDoSeek(false, 0)
	assert.NoError(t, err)
}

func TestSubscribeToStateChanges(t *testing.T) {
	pm := NewPlaylistManager()
	ch := pm.SubscribeToStateChanges()

	assert.NotNil(t, ch)
}
