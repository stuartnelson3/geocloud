package main

import (
	"testing"
	"time"
)

func TestReadOut(t *testing.T) {
	var (
		outc   = make(chan Track)
		states = []string{"MN", "CA", "NY", "WI"}
	)
	go func(c chan Track) {
		for _, state := range states {
			c <- Track{USState: state}
		}
	}(outc)
	tracks := readOut(outc, time.After(5*time.Millisecond))
	if len(tracks) != len(states) {
		t.Fatalf("got %d, expected %d", len(tracks), len(states))
	}
}

func TestTracksByState(t *testing.T) {
	var (
		stateStructs = []struct{ name, state string }{
			{name: "Herp", state: "MN"},
			{name: "Herp", state: "MN"},
			{name: "Herp", state: "MN"},
			{name: "Herp", state: "MN"},
			{name: "Herp", state: "MN"},
			{name: "Other", state: "MN"},
			{name: "Other", state: "MN"},
			{name: "Other", state: "MN"},
			{name: "Herp", state: "WI"},
			{name: "Herp", state: "WI"},
			{name: "Herp", state: "WI"},
			{name: "Herp", state: "WI"},
			{name: "Other", state: "WI"},
			{name: "Other", state: "WI"},
			{name: "Something", state: "WI"},
			{name: "Something", state: "WI"},
			{name: "Something", state: "WI"},
		}
		tracks = make([]Track, len(stateStructs))
	)
	for i, t := range stateStructs {
		tracks[i] = createTrack(t.name, t.state)
	}
	trackMap := tracksByState(tracks)
	if len(trackMap["MN"]) != 2 {
		t.Fatalf("MN: got %d, expected 2", len(trackMap["MN"]))
	}
	if len(trackMap["WI"]) != 3 {
		t.Fatalf("WI: got %d, expected 3", len(trackMap["WI"]))
	}

	stateStructs = []struct{ name, state string }{
		{name: "Herp", state: "MN"},
		{name: "Herp", state: "MN"},
		{name: "Herp", state: "MN"},
		{name: "Herp", state: "MN"},
		{name: "Herp", state: "MN"},
		{name: "Something", state: "WI"},
		{name: "Something", state: "WI"},
		{name: "Something", state: "WI"},
	}
	tracks = make([]Track, len(stateStructs))
	for i, t := range stateStructs {
		tracks[i] = createTrack(t.name, t.state)
	}
	trackMap = tracksByState(tracks)
	if trackMap["MN"][0].Count != 5 {
		t.Fatalf("MN: got %d, expected 5", trackMap["MN"][0].Count)
	}
	if trackMap["WI"][0].Count != 3 {
		t.Fatalf("WI: got %d, expected 3", trackMap["WI"][0].Count)
	}
}

func createTrack(title, state string) Track {
	return Track{Title: title, USState: state}
}
