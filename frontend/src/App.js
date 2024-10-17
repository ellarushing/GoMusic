import React, { useEffect, useState } from "react"
import axios from "axios"
import './App.css'



function App() {
  const [playlists, setPlaylists] = useState([]); // stores playlist data
  const [loading, setLoading] = useState(true);


  // useState Hook -> takes initial state value as arg and returns an
  // array w/two elements: curr state value (playlists), func to update state value (setPlaylists)

  useEffect(() => {
    // get data from API
    const getData = async () => {
      try {
        // GET request to API endpoint
        const result = await axios.get('http://localhost:8888/playlists');
        console.log('Fetched data:', result.data); // Log the fetched data
        if(result.data.items) {
          // use setPlaylists to get data
          setPlaylists(result.data.playlists.playlist);
          console.log('success')
        } else {
          console.warn('No items found in data: ', result.data);
        }
      } catch(error) {
        console.error('Error fetching the playlists', error.message);
        if(error.response) {
          console.error('Error data:', error.response.data);
        } else if (error.request) {
          // The request was made but no response was received
          console.error('No response received:', error.request);
        } else {
          // Something happened in setting up the request that triggered an Error
          console.error('Error setting up request:', error.message);
        }
      }
    };
    getData();
  }, []);

  return (
    <div className="App">
      <header className="App-header">
        <h1> Spotify Playlists </h1>
        {playlists.map(playlist => (
          <div key={playlist.id} className="playlist">
            <h2>{playlist.name}</h2>
            <p>Total Tracks: {playlist.tracks.total}</p>
            <ul>
              {playlist.tracks.items.map(track => {
                <li key={track.id}>
                  {track.name} by {track.artists.map(artist => artist.name).join(', ')}
                </li>
              })}
            </ul>
          </div>
        ))}
      </header>
    </div>
  );
}

export default App;
