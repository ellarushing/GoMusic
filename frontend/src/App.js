import React, { useEffect, useState } from "react"
import axios from "axios"
import './App.css'



function App() {
  const [playlists, setPlaylists] = useState([]);
  const [error, setError] = useState(null);
  const [loading, setLoading] = useState(true); // diff loading
  // useState Hook -> takes initial state value as arg and returns an
  // array w/two elements: curr state value (playlists), func to update state value (setPlaylists)

  useEffect(() => {
    // get data from API
    const getData = async () => {
      try {
        // GET request to API endpoint
        const result = await axios.get('http://localhost:8888/callback');
        console.log('Fetched data:', result.data); // Log the fetched data
       // if(result.data.items) {
        if(result.data?.playlists?.playlists) {
        // use setPlaylists to get data
        // setPlaylists(result.data.items);
          setPlaylists(result.data.playlists.playlists)
          console.log('success')
        } else {
          setError('No playlist data found in response');
          console.warn('No items found in data: ', result.data);
        }
      } catch(error) {
        setError(error.message);
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
      } finally {
        setLoading(false);
      }
    };
    getData();
  }, []);

  if(loading) {
    return (
      <div className="App">
        <header className="App-header">
          <h1>Loading...</h1>
        </header>
      </div>
    )
  }



  return (
    <div className="App">
      <header className="App-header">
        <h1> Spotify Playlists</h1>
        <div className="playlists-container">
          {playlists && playlists.map((playlist, index) => (
              <div key={index} className="playlist-card">
                <h2>{playlist.name}</h2>
                <p>Tracks: {playlist.track_count}</p>
                <p>Owner: {playlist.owner[0].display_name}</p>
                </div>
          ))}
        </div>
      </header>
    </div>
    
  );
}

export default App;



