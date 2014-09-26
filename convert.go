package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
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
		// <-time.After(time.Second)
	}
}

func main() {
	// id := "97542154"
	// url := fmt.Sprintf("http://api.soundcloud.com/tracks/%s.json?client_id=1182e08b0415d770cfb0219e80c839e8", id)
	// fmt.Println(url)
	// resp, err := http.Get(url)
	// if err != nil {
	// 	log.Printf("Failed to GET %s", id)
	// 	return
	// }
	// defer resp.Body.Close()
	// t := Track{
	// 	USState: "Minnesota",
	// }
	// err = json.NewDecoder(resp.Body).Decode(&t)
	// fmt.Println(t)
	// return
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
	// state, track_id, playcount, title, link, artwork
	header = append(header, []string{"playcount", "title", "link", "artwork"}...)

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
		case <-time.After(time.Second * 10):
			log.Println("Timeout, continuing")
			quit = true
		}
		if quit {
			break
		}
	}

	data := make([][]string, len(tracks))
	for i := 0; i < len(tracks); i++ {
		t := tracks[i]
		data[i] = []string{t.USState, strconv.Itoa(t.Id), strconv.Itoa(t.PlaybackCount), t.Title, t.PermalinkUrl, t.ArtworkUrl}
	}
	newCSV, err := os.Create("./random_state_data.csv")
	if err != nil {
		log.Println("Error creating new CSV file: ", err)
		return
	}

	w := csv.NewWriter(newCSV)
	err = w.Write(header)
	if err != nil {
		log.Println("Error writing header to new CSV file: ", err)
		return
	}
	err = w.WriteAll(data)
	if err != nil {
		log.Println("Error writing body to new CSV file: ", err)
		return
	}
}

type Track struct {
	Id            int    `json:"id"`
	PlaybackCount int    `json:"playback_count"`
	Title         string `json:"title"`
	PermalinkUrl  string `json:"permalink_url"`
	ArtworkUrl    string `json:"artwork_url"`
	USState       string
}
