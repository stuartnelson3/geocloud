package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func querySoundcloud(id, state string, out chan<- Track) {
	resp, err := http.Get(fmt.Sprintf("http://api.soundcloud.com/tracks/%s.json?client_id=1182e08b0415d770cfb0219e80c839e8", id))
	if err != nil {
		log.Printf("Failed to GET %s", id)
		return
	}
	defer resp.Body.Close()
	t := Track{
		USState: state,
	}
	err = json.NewDecoder(resp.Body).Decode(&t)
	if err != nil {
		log.Println("Error decoding json: ", err)
		log.Println("Id:", id)
	}
	out <- t
}

func apiQuerier(initialRow chan []string, out chan Track) {
	for r := range initialRow {
		querySoundcloud(r[1], r[0], out)
	}
}

func readOut(out <-chan Track) []Track {
	tracks := make([]Track, 0)
	timeout := time.After(time.Second * 10)
	for {
		select {
		case t := <-out:
			log.Println("Appending", t.USState)
			tracks = append(tracks, t)
		case <-timeout:
			log.Println("Timeout, continuing")
			return tracks
		}
	}
}

func tracksByState(tracks []Track) map[string][]Track {
	stateMap := make(map[string][]Track)
	for _, t := range tracks {
		stateMap[t.USState] = append(stateMap[t.USState], t)
	}
	// now with the states: sum the occurrence of each track
	for state, trks := range stateMap {
		songMap := make(map[string]Track)
		for _, t := range trks {
			track, prs := songMap[t.Title]
			if prs {
				track.Count++
				songMap[t.Title] = track
			} else {
				t.Count = 1
				songMap[t.Title] = t
			}
		}
		stateMap[state] = make([]Track, 0)
		for _, t := range songMap {
			stateMap[state] = append(stateMap[state], t)
		}
	}
	return stateMap
}

func stateJson(stateMap map[string][]Track) []State {
	data := make([]State, len(stateMap))
	var i int
	for name, tracks := range stateMap {
		var sum int
		for _, t := range tracks {
			sum += t.Count
		}
		data[i] = State{
			Name:       name,
			TotalPlays: sum,
			Tracks:     tracks,
		}
		i++
	}
	return data
}

func main() {
	// http://ip-api.com/json/[ip]
	// obj["zip"]

	// http://zip.getziptastic.com/v2/US/48867
	// obj["state"]

	f, err := os.Open("./state_seed.csv")
	if err != nil {
		log.Println("Error opening state_seed: ", err)
		return
	}
	defer f.Close()
	csvF := csv.NewReader(f)
	header, err := csvF.Read()
	if err != nil {
		log.Println("Error reading header: ", err)
		return
	}
	// state, track_id, playcount, title, link, artwork, count
	header = append(header, []string{"playcount", "title", "link", "artwork", "count"}...)

	body, err := csvF.ReadAll()
	if err != nil {
		log.Println("Error reading body: ", err)
		return
	}

	initialRow := make(chan []string)
	out := make(chan Track)
	for i := 0; i < 100; i++ {
		go apiQuerier(initialRow, out)
	}
	go func() {
		for _, row := range body {
			initialRow <- row
		}
	}()

	tracks := readOut(out)
	stateMap := tracksByState(tracks)

	jf, err := os.Create("./states.json")
	if err != nil {
		log.Println("Error making json file:", err)
		return
	}
	defer jf.Close()
	json.NewEncoder(jf).Encode(stateJson(stateMap))
}

type State struct {
	Name       string  `json:"name"`
	TotalPlays int     `json:"total_plays"`
	Tracks     []Track `json:"tracks"`
}

type Track struct {
	Id            int    `json:"id"`
	PlaybackCount int    `json:"playback_count"`
	Title         string `json:"title"`
	PermalinkUrl  string `json:"permalink_url"`
	ArtworkUrl    string `json:"artwork_url"`
	USState       string `json:"us_state"`
	Count         int    `json:"count"`
}
