package playlists

import (
	"testing"

	Features "github.com/Icey-Glitch/Syncplay-G/features"
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

	// Test case 3: Empty username
	err = pm.RemoveUserPlaystate("")
	assert.Error(t, err)

}

func TestSetUserPlaystate(t *testing.T) {
	pm := NewPlaylistManager()

	// Test case 1: Empty username
	err := pm.SetUserPlaystate("", 0, false, false, "", 0, false)
	assert.Error(t, err)

	// Test case 2: Non-existent user
	username := "testUser"

	err = pm.SetUserPlaystate(username, 0, false, false, "", 0, false)
	assert.Error(t, err)

	// Test case 3: Existing user
	err = pm.CreateUserPlaystate(username)
	assert.NoError(t, err)

	err = pm.SetUserPlaystate(username, 0, false, false, "", 0, false)
	assert.NoError(t, err)

	// Test case 4: Set doSeek to true
	err = pm.SetUserPlaystate(username, 0, false, true, "", 0, false)
	assert.NoError(t, err)

	// Test case 5: Ignore is true
	err = pm.SetUserPlaystate(username, 0, false, false, "", 0, true)
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

// Change doSeek to the specified value for all users in the playlist
func TestSetUserPlaystateDoSeek(t *testing.T) {
	pm := NewPlaylistManager()

	// create a user
	err := pm.CreateUserPlaystate("TestUser")

	err = pm.SetUserPlaystate("TestUser", 0, false, false, "", 0, false)
	assert.NoError(t, err)

	err = pm.SetUserPlaystate("TestUser", 0, false, true, "", 0, false)
	assert.NoError(t, err)

}

// Position
func TestSetUsersPosition(t *testing.T) {
	pm := NewPlaylistManager()

	// create a user
	err := pm.CreateUserPlaystate("TestUser")
	assert.NoError(t, err)

	// set position
	err = pm.SetUsersPosition(10, 0)
	assert.NoError(t, err)

	// set position
	err = pm.SetUsersPosition(20, 0)
	assert.NoError(t, err)

	// case 3: Time is less than last message age
	err = pm.SetUsersPosition(30, 20)

	err = pm.SetUsersPosition(10, 10)

	// check if position is set
	assert.Equal(t, pm.Playlist.Position, float64(30), "Position Used old value")

	// case 3: Time is less than last message age
	err = pm.SetUsersPosition(30, 20)

	err = pm.SetUsersPosition(10, -100)

	// check if position is set
	assert.Equal(t, pm.Playlist.Position, float64(30), "Position Used old value")
}

// AddFile
func TestAddFile(t *testing.T) {
	pm := NewPlaylistManager()

	// test case 1: empty file
	_, err := pm.AddFile(0, "", 0, "")
	assert.Error(t, err)

	// test case 2: valid file, but user does not exist
	_, err = pm.AddFile(0, "testFile", 0, "testUser")
	assert.Error(t, err)

	// test case 3: valid file, user exists
	err = pm.CreateUserPlaystate("testUser")
	assert.NoError(t, err)

	_, err = pm.AddFile(0, "testFile", 0, "testUser")
	assert.NoError(t, err)

	// test case 4: valid file, user exists, file already exists
	_, err = pm.AddFile(0, "testFile", 0, "testUser")
	assert.NoError(t, err)

	// test case 5: valid file, user exists, file does not exist
	_, err = pm.AddFile(0, "testFile2", 0, "testUser")
	assert.NoError(t, err)

	// test case 6: valid file, user exists, file exists in playlist
	_, err = pm.AddFile(0, "testFile2", 0, "testUser")
	assert.NoError(t, err)

	// test case 7: valid file, user exists, file does not exist in playlist. Shared playlist is enabled
	// set shared playlist to true
	Features.GlobalFeatures.SharedPlaylists = true
	_, err = pm.AddFile(0, "testFile3", 0, "")
	assert.NoError(t, err)

	// test case 8: valid file, user exists, file exists in playlist. Shared playlist is enabled
	_, err = pm.AddFile(0, "testFile3", 0, "")
	assert.NoError(t, err)

}
