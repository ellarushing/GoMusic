package main

import (
	"encoding/json"
	"fmt"      //enables formatting I/O funcs
	"log"      // logging error messages
	"net/http" //making HTTP requests & handling responses
	"os"
	"strings"
	"sync"
	"github.com/joho/godotenv" // for using .env for confidential info
	"github.com/rs/cors"
	"github.com/zmb3/spotify" // Go client library for Spotify Web API
	"golang.org/x/oauth2"
)

const (
	redirectURI = "http://localhost:8888/callback" // where Spotify sends user after authentication
)

// global variables
var (
	auth = spotify.NewAuthenticator(redirectURI, spotify.ScopePlaylistReadPrivate, spotify.ScopeUserTopRead)
	state = "state-token"
	playlists = &Playlists{}
	topArtists = &TopArtists{}
	topTracks = &TopTracks{}
	combinedData = &CombinedData{}
	userToken *oauth2.Token // this is to connect to React application
)

// used for managing concurrent access & store playlist data
type Playlists struct {
	sync.Mutex // lock to protect shared data from being accessed by multiple go-routines
	Data *spotify.SimplePlaylistPage // pointer to spotify playlist obj to hold playlist data
}

type TopArtists struct {
	sync.Mutex
	Data *spotify.FullArtistPage // multiple top artists data
}

type TopTracks struct {
	sync.Mutex
	Data *spotify.FullTrackPage // holds multiple top tracks data
}

type CombinedData struct {
	Playlists *spotify.SimplePlaylistPage `json:"playlists"`
	TopArtists *spotify.FullArtistPage `json:"topArtists"`
	TopTracks *spotify.FullTrackPage `json:"topTracks"`
}

func main() {
	// get info from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	
	// for Spotify API Authentication
	auth.SetAuthInfo(os.Getenv("SPOTIFY_CLIENT_ID"), os.Getenv("SPOTIFY_CLIENT_SECRET"))

	// get login url
	url := auth.AuthURL(state) // gets URL where user logins and authorizes the application
	fmt.Println("Please log in to Spotify by visiting the following page in your browser:" , url)
	
	// HTTP server to listen on callback URL
	http.HandleFunc("/callback", handleCallback)
	http.HandleFunc("/playlists", handlePlaylists)
	http.HandleFunc("/topArtists", handleTopArtists)
	http.HandleFunc("/topTracks", handleTopTracks)

	handler := cors.Default().Handler(http.DefaultServeMux)

	log.Println("Starting server at :8888")
	if err := http.ListenAndServe(":8888", handler); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

func handleCallback(w http.ResponseWriter, r *http.Request) {
	log.Println("Received request for /callback")
	token, err := auth.Token(state, r) // exchange authorization code for access token
	if err != nil { // error message
		http.Error(w, "Couldn't get token", http.StatusForbidden)
		log.Fatal(err)
	}
	log.Println("Token received")

	// dealing with playlists
	client := auth.NewClient(token)
	playlists.Data, err = client.CurrentUsersPlaylists() // gets user's playlists
	if err != nil {
		log.Fatalf("Failed to get playlist: %v", err)
	}
	log.Println("Playlists Successful")

	// dealing with top artists
	topArtists.Data, err = client.CurrentUsersTopArtists() // get user's top artists
	if err != nil {
		log.Fatalf("Failed to get Top Artists: %v", err)
	}
	log.Println("Top Artists Successful")

	// dealing with top tracks
	topTracks.Data, err = client.CurrentUsersTopTracks() // get user's top tracks
	if err != nil {
		log.Fatalf("Failed to get Top Tracks: %v", err)
	}
	log.Println("Top Tracks Successful")

	*combinedData = CombinedData {
		Playlists: playlists.Data,
		TopArtists: topArtists.Data,
		TopTracks: topTracks.Data,
	}

	if combinedData.Playlists == nil {
		log.Fatal("Failed to save Playlist data into combined struct")
	}
	if combinedData.TopArtists == nil {
		log.Fatal("Failed to save TopArtists data into combined struct")
	}
	if combinedData.TopTracks == nil {
		log.Fatal("Failed to save TopTracks data into combined struct")
	}
	log.Println("Combined Data fetched")


	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(combinedData); err != nil {
		log.Println("Failed to encode combinedData:", err)
	}
	userToken = token;

	if userToken == nil {
		http.Error(w, "Not authenticated", http.StatusUnauthorized);
		return
	}
}

func handlePlaylists(w http.ResponseWriter, r *http.Request) {
	client := auth.NewClient(userToken)
	playlists, err := client.CurrentUsersPlaylists()
	if err != nil {
		http.Error(w, "Failed to fetch playlists", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(playlists)
}

// need this to somehow check every week and regather the data
// to actively recheck and reevaluate
func handleTopArtists(w http.ResponseWriter, r *http.Request) {
	client := auth.NewClient(userToken)
	topArtists, err := client.CurrentUsersTopArtists()
	if err != nil {
		http.Error(w, "Failed to get Top Artists", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(topArtists)
}

func handleTopTracks(w http.ResponseWriter, r *http.Request) {
	client := auth.NewClient(userToken)
	topTracks, err := client.CurrentUsersTopTracks()
	if err != nil {
		http.Error(w, "Failed to get Top Tracks", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(topTracks)

}

func handleListeningHistory(w http.ResponseWriter, r *http.Request) {

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

