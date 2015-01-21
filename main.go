package main

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"

	_ "github.com/lib/pq"

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
	if t.ID != 0 {
		// weird behaviour where tracks with id 0 are being returned
		outc <- t
	} else {
		log.Println(t.Title, "has no ID")
	}
	<-time.After(50 * time.Millisecond)
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

func readOut(outc <-chan *Track) []*Track {
	tracks := make([]*Track, 0)
	var i int
	for {
		select {
		case t := <-outc:
			tracks = append(tracks, t)
			log.Printf("wrote %d tracks", i)
			i++
		case <-time.After(5 * time.Second):
			log.Printf("wrote %d tracks", len(tracks))
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
	// need to pull from just country == US
	// figure by 2 letter state code
	// have listener_id to do gender/age filtering (maybe)
	//

	// Dump from db into this, csv of state,track_id
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
	// TODO: Aggregate the ids first to eliminate duplicate API calls.
	for i := 0; i < 100; i++ {
		go apiQuerier(rowc, clientID, outc)
	}
	go func() {
		for _, row := range body {
			rowc <- row
		}
	}()

	tracks := readOut(outc)
	stateMap := tracksByState(tracks)

	jf, err := os.Create("./public/states.json")
	if err != nil {
		log.Println("Error making json file:", err)
		return
	}
	defer jf.Close()
	json.NewEncoder(jf).Encode(makeStates(stateMap))
}

func generateSeed(dbs string) {
	db, err := sql.Open("postgres", dbs)
	if err != nil {
		log.Fatal(err)
	}

	var (
		n = 10000000
		i = 1
	)

	rows, err := db.Query(`select region, track_id from playduration where country = 'US' and region <> '' limit $1`, n)
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Create("./state-temp.csv")
	if err != nil {
		log.Fatal(err)
	}
	w := csv.NewWriter(f)
	w.Write([]string{"state", "track_id"})

	for rows.Next() {
		d := make([]string, 2)
		rows.Scan(&d[0], &d[1])
		w.Write(d)
		fmt.Printf("Writing record: %d", i)
		i++
	}
	w.Flush()
	if err = w.Error(); err != nil {
		log.Fatal(err)
	}

	f.Close()
	os.Rename("./state-temp.csv", "./state_seed.csv")
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
		port     = flag.String("port", "8080", "port to listen on")
		seed     = flag.Bool("seed", false, "generate seed file and bail")
		regen    = flag.Bool("regen", false, "regenerate song file hourly")
		clientID = flag.String("clientID", "", "your clientID")
		dbs      = flag.String("db", "", "db connection string")
	)
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()

	if *regen {
		// ticker := time.NewTicker(time.Hour)
		// go func() {
		// 	for range ticker.C {
		if *clientID == "" {
			log.Println("need to set sc client id")
			os.Exit(1)
		}
		convert(*clientID)
		return
		// 	}
		// }()
	}

	if *seed {
		if *dbs == "" {
			log.Println("need to set database connection string")
			os.Exit(1)
		}
		generateSeed(*dbs)
		return
	}

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
