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
	for i := 0; i < 5; i++ {
		go apiQuerier(initialRow, out)
	}
	go func() {
		for i, row := range body {
			log.Printf("Row %d", i)
			initialRow <- row
		}
		// close(initialRow)
	}()

	tracks := make([]Track, 0)
	// need to account for failed network requests somehow.
	for {
		var quit bool
		select {
		case t := <-out:
			log.Println("Appending", t.USState)
			tracks = append(tracks, t)
		case <-time.After(time.Second * 3):
			log.Println("Timeout, continuing")
			quit = true
		}
		if quit {
			break
		}
	}

	// filter tracks to state
	stateMap := make(map[string][]Track)
	for _, t := range tracks {
		stateMap[t.USState] = append(stateMap[t.USState], t)
	}
	// now with the states: sum the occurrence of each track
	// trax := make([]Track, 50)
	// plc := 0
	for state, trks := range stateMap {
		songMap := make(map[string]int)
		count := 0
		var track Track
		for _, t := range trks {
			_, prs := songMap[t.Title]
			if prs {
				songMap[t.Title] = songMap[t.Title] + 1
			} else {
				songMap[t.Title] = 1
			}
			if songMap[t.Title] > count {
				count = songMap[t.Title]
				track = t
			}
		}
		track.Count = count
		stateMap[state] = []Track{track}
		// trax[plc] = track
		// plc++
	}

	stateJson := make([]State, len(stateMap))
	var i int
	for name, tracks := range stateMap {
		var sum int
		for _, t := range tracks {
			sum += t.Count
		}
		stateJson[i] = State{
			Name:       name,
			TotalPlays: sum,
			Tracks:     tracks,
		}
		i++
	}

	jf, err := os.Create("./states.json")
	if err != nil {
		log.Println("Error making json file:", err)
		return
	}
	defer jf.Close()
	json.NewEncoder(jf).Encode(stateJson)

	// data := make([][]string, len(trax))
	// for i, _ := range data {
	// 	t := trax[i]
	// 	data[i] = []string{
	// 		t.USState,
	// 		strconv.Itoa(t.Id),
	// 		strconv.Itoa(t.PlaybackCount),
	// 		t.Title,
	// 		t.PermalinkUrl,
	// 		t.ArtworkUrl,
	// 		strconv.Itoa(t.Count),
	// 	}
	// }
	// newCSV, err := os.Create("./random_state_data.csv")
	// if err != nil {
	// 	log.Println("Error creating new CSV file: ", err)
	// 	return
	// }
	//
	// w := csv.NewWriter(newCSV)
	// err = w.Write(header)
	// if err != nil {
	// 	log.Println("Error writing header to new CSV file: ", err)
	// 	return
	// }
	// err = w.WriteAll(data)
	// if err != nil {
	// 	log.Println("Error writing body to new CSV file: ", err)
	// 	return
	// }
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
