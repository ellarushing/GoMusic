package main

import (
	"encoding/json"
	"fmt"      //enables formatting I/O funcs
	"log"      // logging error messages
	"net/http" //making HTTP requests & handling responses
	"strings"
	"sync"
	"github.com/joho/godotenv" // for using .env for confidential info
	"golang.org/x/oauth2"
	"github.com/rs/cors"
	"github.com/zmb3/spotify" // Go client library for Spotify Web API
)

const (
	redirectURI = "http://localhost:8888/callback" // where Spotify sends user after authentication
)

var (
	auth = spotify.NewAuthenticator(redirectURI, spotify.ScopePlaylistReadPrivate)
	state = "state-token"
	playlists = &Playlist{}
	userToken *oauth2.Token // this is to connect to React application
)

type Playlists struct {
	sync.Mutex
	Data *spotify.SimplePlaylistPage
}

func main() {
	// for Spotify API Authentication
	auth := spotify.NewAuthenticator(redirectURI, spotify.ScopePlaylistReadPrivate)
	auth.SetAuthInfo("719bcf23f8cd44618e8b510bd4798dd2", "5f1d4527805b41a09fc6cfe8c0dc1113")

	// err := godotenv.Load()
	// if err != nil {
	// 	log.Fatalf("Error loading .env file: %v", err)
	// }


	// get login url
	url := auth.AuthURL("state-token") // gets URL where user logins and authorizes the application
	fmt.Println("Please log in to Spotify by visiting the following page in your browser:" , url)
	// HTTP server to listen on callback URL
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Received request for /callback")
		token, err := auth.Token("state-token", r) // exchange authorization code for access token
		if err != nil { // error message
			http.Error(w, "Couldn't get token", http.StatusForbidden)
			log.Fatal(err)
		}
		log.Println("Token received")

		client := auth.NewClient(token)
		playlists, err := client.CurrentUsersPlaylists() // gets user's playlists
		if err != nil {
			log.Fatalf("Failed to get playlist: %v", err)
		}
		log.Println("Playlists fetched")
		//fmt.Fprintf(w, "Found your playlists: %+v", playlists)

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(playlists); err != nil {
			log.Println("Failed to encode playlists:", err)
		}
		userToken = token;
	})

	http.HandleFunc("/playlists", handlePlaylists)

	handler := cors.Default().Handler(http.DefaultServeMux)

	log.Println("Starting server at :8888")
	if err := http.ListenAndServe(":8888", handler); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}


func handlePlaylists(w http.ResponseWriter, r *http.Request) {
	if userToken == nil {
		http.Error(w, "Not authenitcated", http.StatusUnauthorized);
		return
	}

	client := auth.NewClient((userToken))
	playlists, err := client.CurrentUsersPlaylists()
	if err != nil {
		http.Error(w, "Failed to fetch playlists", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(playlists)
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
			output += fmt.Sprintf("- %s by %s\n", track.Name, strings.Join(artists, ","))
			
		}
	}
	return output, nil
}
