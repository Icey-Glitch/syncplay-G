package messages

import (
	Features "github.com/Icey-Glitch/Syncplay-G/features"
)

type HelloResponseMessage struct {
	Hello struct {
		Username string `json:"username"`
		Room     struct {
			Name string `json:"name"`
		} `json:"room"`
		Version     string            `json:"version"`
		RealVersion string            `json:"realversion"`
		Features    Features.Features `json:"features"`
		MOTD        string            `json:"motd"`
	} `json:"Hello"`
}

func CreateHelloResponse(username, version, roomName string) HelloResponseMessage {
	return HelloResponseMessage{
		Hello: struct {
			Username string `json:"username"`
			Room     struct {
				Name string `json:"name"`
			} `json:"room"`
			Version     string            `json:"version"`
			RealVersion string            `json:"realversion"`
			Features    Features.Features `json:"features"`
			MOTD        string            `json:"motd"`
		}{
			Username: username,
			Room: struct {
				Name string `json:"name"`
			}{
				Name: roomName,
			},
			Version:     version,
			RealVersion: "1.7.3",
			Features:    Features.GlobalFeatures,
			MOTD:        "",
		},
	}
}
