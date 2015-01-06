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

func querySoundcloud(id, state string, outc chan<- *Track) {
	resp, err := http.Get(fmt.Sprintf("http://api.soundcloud.com/tracks/%s.json?client_id=1182e08b0415d770cfb0219e80c839e8", id))
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
		log.Println("Error decoding json: ", err)
		log.Println("Id:", id)
	}
	outc <- t
}

func apiQuerier(rowc chan []string, outc chan *Track) {
	for r := range rowc {
		querySoundcloud(r[1], r[0], outc)
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
	for _, t := range tracks {
		stateMap[t.USState] = append(stateMap[t.USState], t)
	}
	// now with the states: sum the occurrence of each track
	for state, trks := range stateMap {
		songMap := make(map[string]*Track)
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
		stateMap[state] = make([]*Track, 0)
		for _, t := range songMap {
			stateMap[state] = append(stateMap[state], t)
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

func convert() {
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
		go apiQuerier(rowc, outc)
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
	Id            int    `json:"id"`
	PlaybackCount int    `json:"playback_count"`
	Title         string `json:"title"`
	PermalinkUrl  string `json:"permalink_url"`
	ArtworkUrl    string `json:"artwork_url"`
	USState       string `json:"us_state"`
	Count         int    `json:"count"`
}

func main() {
	port := flag.String("port", "8080", "port to listen on")
	flag.Parse()
	// ticker := time.NewTicker(time.hour * 24)
	// go func() {
	//     for t := range ticker.C {
	// 		convert()
	//     }
	// }()
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
