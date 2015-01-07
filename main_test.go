package main

import (
	"testing"
	"time"
)

func TestReadOut(t *testing.T) {
	var (
		outc   = make(chan *Track)
		states = []string{"MN", "CA", "NY", "WI"}
	)
	go func(c chan *Track) {
		for _, state := range states {
			c <- &Track{USState: state}
		}
	}(outc)
	tracks := readOut(outc, time.After(5*time.Millisecond))
	if len(tracks) != len(states) {
		t.Fatalf("got %d, expected %d", len(tracks), len(states))
	}
}

func TestTracksByState(t *testing.T) {
	var (
		stateStructs = []struct {
			id           int
			title, state string
		}{
			{id: 1, title: "Herp", state: "MN"},
			{id: 1, title: "Herp", state: "MN"},
			{id: 1, title: "Herp", state: "MN"},
			{id: 1, title: "Herp", state: "MN"},
			{id: 1, title: "Herp", state: "MN"},
			{id: 2, title: "Other", state: "MN"},
			{id: 2, title: "Other", state: "MN"},
			{id: 2, title: "Other", state: "MN"},
			{id: 1, title: "Herp", state: "WI"},
			{id: 1, title: "Herp", state: "WI"},
			{id: 1, title: "Herp", state: "WI"},
			{id: 1, title: "Herp", state: "WI"},
			{id: 2, title: "Other", state: "WI"},
			{id: 2, title: "Other", state: "WI"},
			{id: 3, title: "Something", state: "WI"},
			{id: 3, title: "Something", state: "WI"},
			{id: 3, title: "Something", state: "WI"},
		}
		tracks = make([]*Track, len(stateStructs))
	)
	for i, t := range stateStructs {
		tracks[i] = createTrack(t.id, t.title, t.state)
	}
	trackMap := tracksByState(tracks)
	if len(trackMap["MN"]) != 2 {
		t.Fatalf("MN: got %d, expected 2", len(trackMap["MN"]))
	}
	if len(trackMap["WI"]) != 3 {
		t.Fatalf("WI: got %d, expected 3", len(trackMap["WI"]))
	}

	stateStructs = []struct {
		id           int
		title, state string
	}{
		{id: 1, title: "Herp", state: "MN"},
		{id: 1, title: "Herp", state: "MN"},
		{id: 1, title: "Herp", state: "MN"},
		{id: 1, title: "Herp", state: "MN"},
		{id: 1, title: "Herp", state: "MN"},
		{id: 2, title: "Something", state: "WI"},
		{id: 2, title: "Something", state: "WI"},
		{id: 2, title: "Something", state: "WI"},
	}
	tracks = make([]*Track, len(stateStructs))
	for i, t := range stateStructs {
		tracks[i] = createTrack(t.id, t.title, t.state)
	}

	expected := map[string]int{"MN": 5, "WI": 3}

	trackMap = tracksByState(tracks)
	for state, trks := range trackMap {
		if trks[0].Count != expected[state] {
			t.Fatalf("%s: got %d, expected %d", state, trks[0].Count, expected[state])
		}
		if trks[0].USState != state {
			t.Fatalf("got %s, expected %s", trks[0].USState, state)
		}
	}
}

func TestMakeStates(t *testing.T) {
	var (
		stateStructs = []struct {
			id           int
			title, state string
		}{
			{id: 1, title: "Herp", state: "MN"},
			{id: 1, title: "Herp", state: "MN"},
			{id: 1, title: "Herp", state: "MN"},
			{id: 1, title: "Herp", state: "MN"},
			{id: 1, title: "Herp", state: "MN"},
			{id: 2, title: "Something", state: "WI"},
			{id: 2, title: "Something", state: "WI"},
			{id: 2, title: "Something", state: "WI"},
		}
		expected = map[string]int{"MN": 5, "WI": 3}
		tracks   = make([]*Track, len(stateStructs))
	)
	for i, t := range stateStructs {
		tracks[i] = createTrack(t.id, t.title, t.state)
	}
	trackMap := tracksByState(tracks)
	states := makeStates(trackMap)
	if len(states) != 2 {
		t.Fatalf("got %d, expected 2", len(states))
	}

	for _, state := range states {
		if expected[state.Name] != state.TotalPlays {
			t.Fatalf("%s: got %d, expected %d", state.Name, state.TotalPlays, expected[state.Name])
		}
	}
}

func createTrack(id int, title, state string) *Track {
	return &Track{ID: id, Title: title, USState: state}
}
