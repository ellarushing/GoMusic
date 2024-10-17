package main

import (
	//"context"
	"encoding/json"
	"fmt"      //enables formatting I/O funcs
	"log"      // logging error messages
	"net/http" //making HTTP requests & handling responses

	//"strings"
	"os"
	"sync"

	"github.com/joho/godotenv" // for using .env for confidential info
	"github.com/rs/cors"
	"github.com/zmb3/spotify/v2" // Go client library for Spotify Web API
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

// global variables
var (
	redirectURI = "http://localhost:8888/callback"
	authenticator *spotifyauth.Authenticator
	userToken *oauth2.Token // this is to connect to React application
	state = "state-token"
	clientID string
	clientSecret string


	//auth = spotify.NewAuthenticator(redirectURI, spotify.ScopePlaylistReadPrivate, spotify.ScopeUserTopRead)
	//playlists = &Playlists{}
	//topArtists = &TopArtists{}
	//topTracks = &TopTracks{}
	//combinedData = &CombinedData{}
	//userToken *oauth2.Token // this is to connect to React application
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
	Playlists *FormattedPlaylists `json:"playlists"`
	TopArtists *FormattedTopArtists `json:"topArtists"`
	TopTracks *FormattedTopTracks `json:"topTracks"`
}


// used for formatting & parsing the json data
type Artist struct {
	Name string `json:"name"`
	Popularity int `json:"popularity"`
	Genres []string `json:"genres"`
}

func formatArtist(spotifyArtist spotify.FullArtist) Artist {
	return Artist {
		Name: spotifyArtist.Name,
	//	Popularity: spotifyArtist.Popularity,
		Genres: spotifyArtist.Genres,
	}
}

type FormattedTopArtists struct {
	Artists []Artist `json:"artist"`
}

func formatTopArtists(spotifyArtists *spotify.FullArtistPage) FormattedTopArtists {
	formattedArtists := make([]Artist, len(spotifyArtists.Artists))
	for i, artist := range spotifyArtists.Artists {
		formattedArtists[i] = formatArtist(artist)
	}
	return FormattedTopArtists{Artists: formattedArtists}
}

type Track struct {
	Name string `json:"name"`
	Popularity int `json:"popularity"`
	Artists []Artist `json:"artists"`
	Album Album `json:"album"`
	// do you want the preview URL?
}

func formatTrack(spotifyTrack spotify.FullTrack) Track {
	artists := make([]Artist, len(spotifyTrack.Artists))
	for i, artist := range spotifyTrack.Artists {
		artists[i] = Artist {
			Name: artist.Name,
		}
	}
	return Track {
		Name: spotifyTrack.Name,
	//	Popularity: spotifyTrack.Popularity,
		Album: Album{
			Name: spotifyTrack.Album.Name,
			ReleaseDate: spotifyTrack.Album.ReleaseDate,
			Artists: artists,
		},
	}
}


func formatTopTracks(spotifyTracks *spotify.FullTrackPage) FormattedTopTracks {
	formattedTracks := make([]Track, len(spotifyTracks.Tracks))
	for i, track := range spotifyTracks.Tracks {
		formattedTracks[i] = formatTrack(track)
	}
	return FormattedTopTracks{Tracks: formattedTracks}
}

type FormattedTopTracks struct {
	Tracks []Track `json:"track"`
}

type Playlist struct {
	Name string `json:"name"`
	Owner []Owner `json:"owner"`
	NoTracks int `json:"track_count"`
}

type FormattedPlaylists struct {
	Playlists []Playlist `json:"playlist"`
}

func formatAllPlaylists(spotifyPlaylist *spotify.SimplePlaylistPage) FormattedPlaylists {
	playlists := make([]Playlist, len(spotifyPlaylist.Playlists))
	for i, playlist := range spotifyPlaylist.Playlists {
		owner := []Owner{
			{
				Display_Name: playlist.Owner.DisplayName,
			},
		}
		playlists[i] = Playlist {
			Name: playlist.Name,
			Owner: owner,
			NoTracks: int(playlist.Tracks.Total),
		}
	}
	return FormattedPlaylists{Playlists: playlists}
}

type Album struct {
	Name string `json:"name"`
	ReleaseDate string `json:"release_date"`
	Artists []Artist `json:"artists"`
}
type TotalData struct {
	TopArtists []Artist `json:"top_artists"`
	TopTracks []Track `json:"top_tracks"`
	Playlists []Playlist `json:"playlists"`
}

type Owner struct {
	Display_Name string `json:"display_name"`
}

type ListeningHistory struct {
}


func loadEnv() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Verify the environment variables
	clientID = os.Getenv("SPOTIFY_CLIENT_ID")
	clientSecret = os.Getenv("SPOTIFY_CLIENT_SECRET")

	if len(clientID) == 0 || len(clientSecret) == 0 {
		log.Fatal("SPOTIFY_CLIENT_ID or SPOTIFY_CLIENT_SECRET is not set")
	}
}

func initAuth() {
	// Initialize the authenticator with client ID and secret
	log.Printf("Client ID before URL generation: %s", clientID)
	authenticator = spotifyauth.New(
		spotifyauth.WithRedirectURL(redirectURI),
		spotifyauth.WithScopes(
			spotifyauth.ScopeUserTopRead,
			spotifyauth.ScopePlaylistReadPrivate,
		),
		spotifyauth.WithClientID(clientID),
		spotifyauth.WithClientSecret(clientSecret),
	)
}

func main() {
	loadEnv()
	initAuth()
	// get login url
	url := authenticator.AuthURL(state) // gets URL where user logins and authorizes the application
	fmt.Println("Please log in to Spotify by visiting the following page in your browser:" , url)
	
	// HTTP server to listen on callback URL
	http.HandleFunc("/callback", handleCallback)
	http.HandleFunc("/playlists", handlePlaylists)
	//http.HandleFunc("/topArtists", handleTopArtists)
	//http.HandleFunc("/topTracks", handleTopTracks)

	handler := cors.Default().Handler(http.DefaultServeMux)

	log.Println("Starting server at :8888")
	if err := http.ListenAndServe(":8888", handler); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}


// refactor to only handle token exchange
func handleCallback(w http.ResponseWriter, r *http.Request) {
	log.Println("Received request for /callback")

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Code not found", http.StatusBadRequest)
		return
	}
	// Exchange auth code for access code
	token, err := authenticator.Exchange(r.Context(), code)
	//token, err := auth.Token(state, r) // exchange authorization code for access token
	if err != nil {
		http.Error(w, "Couldn't exchange code for token", http.StatusForbidden)
		log.Fatal(err)
		return
	}

	httpClient := authenticator.Client(r.Context(), token)
	client := spotify.New(httpClient)
	userToken = token

	user, err := client.CurrentUser(r.Context())

	//userToken = token
	log.Printf("Logged in as: %s", user.DisplayName)
	// Redirect to front end
	http.Redirect(w, r, "http://localhost:3000", http.StatusSeeOther)

	//log.Println("Token received")
	//client := auth.NewClient(token)

	/*


	playlists.Data, err = client.CurrentUsersPlaylists() // gets user's playlists
	if err != nil {
		log.Fatalf("Failed to get playlist: %v", err)
	}
	formattedPlaylists := formatAllPlaylists(playlists.Data)
	log.Println("Playlists Successful")

	// dealing with top artists
	topArtists.Data, err = client.CurrentUsersTopArtists() // get user's top artists
	if err != nil {
		log.Fatalf("Failed to get Top Artists: %v", err)
	}
	formattedTopArtists := formatTopArtists(topArtists.Data)
	log.Println("Top Artists Successful")

	// dealing with top tracks
	topTracks.Data, err = client.CurrentUsersTopTracks() // get user's top tracks
	if err != nil {
		log.Fatalf("Failed to get Top Tracks: %v", err)
	}
	formattedTopTracks := formatTopTracks(topTracks.Data)
	log.Println("Top Tracks Successful")

	*combinedData = CombinedData {
		Playlists: &formattedPlaylists,
		TopArtists: &formattedTopArtists,
		TopTracks: &formattedTopTracks,
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

	*/
}


 func handlePlaylists(w http.ResponseWriter, r *http.Request) {
	// usertoken set in callback

	//cookie, err := r.Cookie("spotify_token")
	if userToken == nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	// create an OAuth2 token from cookie value
	//token := &oauth2.Token{AccessToken: cookie.Value}
	httpClient := authenticator.Client(r.Context(), userToken)
	client := spotify.New(httpClient) // create spotify client
	
	playlists, err := client.CurrentUsersPlaylists(r.Context())

	if err != nil {
		http.Error(w, "Failed to fetch playlists", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(playlists)
}
