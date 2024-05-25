package main

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