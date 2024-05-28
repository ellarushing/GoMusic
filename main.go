package main

import (
	"fmt" //enables formatting I/O funcs
	"log" // logging error messages
	"net/http" //making HTTP requests & handling responses

	"github.com/zmb3/spotify" // Go client library for Spotify Web API
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


type playlist struct {
	Name string `json: "name"`
	Owner string `json: "owner"`
	Tracks []string `json: "owner"`
	Type string `json: "type"`
	Uri string `json: "uri"`

}

var playlists = []playlist {
	{ Name:   "Summer Hits",
	Owner:  "User123",
	Tracks: []string{"spotify:track:1", "spotify:track:2", "spotify:track:3"},
	Type:   "Public",
	Uri:    "spotify:playlist:1",
	},
	{
	Name:   "Workout",
	Owner:  "User456",
	Tracks: []string{"spotify:track:4", "spotify:track:5", "spotify:track:6"},
	Type:   "Private",
	Uri:    "spotify:playlist:2",
	},
	{
	Name:   "Chill Vibes",
	Owner:  "User789",
	Tracks: []string{"spotify:track:7", "spotify:track:8", "spotify:track:9"},
	Type:   "Public",
	Uri:    "spotify:playlist:3",
	},
}