package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/pat"
)

func querySoundcloud(id, state, url string, outc chan<- *Track) {
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Failed to GET %s", id)
		return
	}
	defer resp.Body.Close()
	t := &Track{
		USState: state,
	}
	err = json.NewDecoder(resp.Body).Decode(&t)
	if err != nil {
		log.Printf("Error decoding json for trackID %s: %s", id, err.Error())
		return
	}
	outc <- t
}

func apiQuerier(rowc chan []string, clientID string, outc chan *Track) {
	for r := range rowc {
		var (
			id    = r[1]
			state = r[0]
			url   = fmt.Sprintf("http://api.soundcloud.com/tracks/%s.json?client_id=%s", id, clientID)
		)
		querySoundcloud(id, state, url, outc)
	}
}

func readOut(outc <-chan *Track, timeout <-chan time.Time) []*Track {
	tracks := make([]*Track, 0)
	for {
		select {
		case t := <-outc:
			tracks = append(tracks, t)
		case <-timeout:
			return tracks
		}
	}
}

func tracksByState(tracks []*Track) map[string][]*Track {
	stateMap := make(map[string][]*Track)
	for _, track := range tracks {
		var found bool
		trks := stateMap[track.USState]
		for _, t := range trks {
			if t.ID == track.ID {
				t.Count++
				found = true
				continue
			}
		}
		if !found {
			track.Count = 1
			stateMap[track.USState] = append(stateMap[track.USState], track)
		}
	}
	return stateMap
}

func makeStates(stateMap map[string][]*Track) []*State {
	var (
		i      int
		states = make([]*State, len(stateMap))
	)
	for name, tracks := range stateMap {
		var sum int
		for _, t := range tracks {
			sum += t.Count
		}
		states[i] = &State{
			Name:       name,
			TotalPlays: sum,
			Tracks:     tracks,
		}
		i++
	}
	return states
}

func convert(clientID string) {
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

	rowc := make(chan []string)
	outc := make(chan *Track)
	for i := 0; i < 100; i++ {
		go apiQuerier(rowc, clientID, outc)
	}
	go func() {
		for _, row := range body {
			rowc <- row
		}
	}()

	tracks := readOut(outc, time.After(time.Second*10))
	stateMap := tracksByState(tracks)

	jf, err := os.Create("./public/states.json")
	if err != nil {
		log.Println("Error making json file:", err)
		return
	}
	defer jf.Close()
	json.NewEncoder(jf).Encode(makeStates(stateMap))
}

type State struct {
	Name       string   `json:"name"`
	TotalPlays int      `json:"total_plays"`
	Tracks     []*Track `json:"tracks"`
}

type Track struct {
	ID            int    `json:"id"`
	PlaybackCount int    `json:"playback_count"`
	Title         string `json:"title"`
	PermalinkUrl  string `json:"permalink_url"`
	ArtworkUrl    string `json:"artwork_url"`
	USState       string `json:"us_state"`
	Count         int    `json:"count"`
}

func main() {
	var (
		port = flag.String("port", "8080", "port to listen on")
		// clientID = flag.String("clientID", "1182e08b0415d770cfb0219e80c839e8", "your clientID")
	)
	flag.Parse()
	// ticker := time.NewTicker(time.Hour)
	// go func(cID string) {
	// 	for range ticker.C {
	// 		convert(cID)
	// 	}
	// }(*clientID)
	m := pat.New()
	m.Get("/public/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, r.URL.Path[1:])
	})

	m.Get("/", func(w http.ResponseWriter, r *http.Request) {
		f, err := os.Open("./main.html")
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
		defer f.Close()
		io.Copy(w, f)
	})

	handler := handlers.LoggingHandler(os.Stdout, m)
	handler = handlers.CompressHandler(handler)
	log.Printf("Listening on :%s", *port)
	log.Fatal(http.ListenAndServe(":"+*port, handler))
}
