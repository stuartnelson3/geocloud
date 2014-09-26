package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
)

func querySoundcloud(id int, out chan<- Track) {
	resp, err := http.Get(fmt.Sprintf("http://api.soundcloud.com/tracks/%d.json?client_id=1182e08b0415d770cfb0219e80c839e8", id))
	if err != nil {
		log.Printf("Failed to GET %d", id)
		return
	}
	defer resp.Body.Close()
	t := Track{}
	err = json.NewDecoder(resp.Body).Decode(&t)
	if err != nil {
		log.Println(err)
	}
	out <- t
}

func main() {
	// http://ip-api.com/json/[ip]
	// obj["zip"]

	// http://zip.getziptastic.com/v2/US/48867
	// obj["state"]

	// ips := make([]byte, 10)
	// for i := 0; i < len(ips); i++ {
	// 	resp, err := http.Get(fmt.Sprintf("http://ip-api.com/json/%s", ips[i]))
	// 	if err != nil {
	// 		return
	// 	}
	// 	fmt.Println(resp)
	// }
	ids := []int{151370911, 97542154, 104449478, 7910262}
	out := make(chan Track)
	for i := 0; i < len(ids); i++ {
		go querySoundcloud(ids[i], out)
	}

	tracks := make([]Track, len(ids))
	// need to account for failed network requests somehow.
	for i := 0; i < len(ids); i++ {
		tracks[i] = <-out
	}
	f, err := os.Open("./us-ag-productivity-2004.csv")
	if err != nil {
		log.Println(err)
		return
	}
	defer f.Close()
	csvF := csv.NewReader(f)
	header, err := csvF.Read()
	if err != nil {
		log.Println("Error reading header: ", err)
		return
	}
	header = append(header, []string{"Playback count", "Title", "Link", "Artwork"}...)

	body, err := csvF.ReadAll()
	if err != nil {
		log.Println("Error reading body: ", err)
		return
	}
	for i := 0; i < len(body); i++ {
		idx := rand.Int() % len(tracks)
		t := tracks[idx]
		data := []string{strconv.Itoa(t.PlaybackCount), t.Title, t.PermalinkUrl, t.ArtworkUrl}
		body[i] = append(body[i], data...)
	}
	newCSV, err := os.Create("./golang_music_data.csv")
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
	err = w.WriteAll(body)
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
}
