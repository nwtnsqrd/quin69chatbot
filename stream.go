package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func isQuinOnline() bool {
	// Twitch API endpoint and query parameters
	endpoint := "https://api.twitch.tv/helix/streams"
	params := "?user_login=quin69"

	// Twitch Client-ID
	clientID := os.Getenv("QUINBOT_CLIENTID")

	// Create HTTP request
	req, err := http.NewRequest("GET", endpoint+params, nil)
	if err != nil {
		fmt.Println("Error creating HTTP request:", err)
		return false
	}

	// Set Twitch Client-ID header
	req.Header.Set("Client-ID", clientID)

	// Send HTTP request
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending HTTP request:", err)
		return false
	}
	defer resp.Body.Close()

	// Read HTTP response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading HTTP response body:", err)
		return false
	}

	// Parse JSON response
	var data struct {
		Data []struct {
			Type    string `json:"type"`
			UserID  string `json:"user_id"`
			Started string `json:"started_at"`
		} `json:"data"`
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		fmt.Println("Error parsing JSON response:", err)
		return false
	}

	// Check if stream is live
	if len(data.Data) > 0 && data.Data[0].Type == "live" {
		return true
	}

	return false
}
