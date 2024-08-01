package main

import (
	"fmt" //enables formatting I/O funcs
	"log" // logging error messages
	"net/http" //making HTTP requests & handling responses
	"github.com/zmb3/spotify" // Go client library for Spotify Web API
	"encoding/json"
	"strings"
)

const (
	redirectURI = "http://localhost:8888/callback" // where Spotify sends user after authentication
)

func main() {
	auth := spotify.NewAuthenticator(redirectURI, spotify.ScopePlaylistReadPrivate)
	auth.SetAuthInfo("719bcf23f8cd44618e8b510bd4798dd2", "5f1d4527805b41a09fc6cfe8c0dc1113")
	// get login url
	url := auth.AuthURL("state-token")
	fmt.Println("Please log in to Spotify by visiting the following page in your browser:" , url)
	// HTTP server to listen on callback URL
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		token, err := auth.Token("state-token", r)
		if err != nil {
			http.Error(w, "Couldn't get token", http.StatusForbidden)
			log.Fatal(err)
		}
		client := auth.NewClient(token)
		playlists, err := client.CurrentUsersPlaylists()
		if err != nil {
			log.Fatalf("Failed to get playlist: %v", err)
		}
		fmt.Fprintf(w, "Found your playlists: %+v", playlists)
	})
	log.Fatal(http.ListenAndServe(":8888", nil))
}


type Playlist struct {
	Name string `json:"name"`
	Tracks struct {
		Href string `json:"href"`
		Total int `json:"total"`
		Items []Track `json:"items"`
	} `json:"tracks"`
	Type string `json:"type"`
	URI string `json:"uri"`
}

type Track struct {
	Name string `json:"name"`
	Artists []struct {
		Name string `json:"name"`
	} `json:"artists"`
}

type PlaylistItems struct {
	Items []Playlist `json:"items"`
}

func formatUserPlaylists(jsonData []byte) (string, error) {
	var playlists PlaylistItems
	err := json.Unmarshal(jsonData, &playlists)
	if err != nil {
		return "", err
	}
	var output string
	for _, playlist := range playlists.Items {
		output += fmt.Sprintf("Playlist: %s\n", playlist.Name)
		output += "Tracks\n"
		for _, track := range playlist.Tracks.Items {
			artists := make([]string, len(track.Artists))
			for i, artist := range track.Artists {
				artists[i] = artist.Name
			}
			output += fmt.Sprintf("- %s by %s\n", track.Name, strings.Join(artists, ", "))
			
		}
	}
	return output, nil
}
