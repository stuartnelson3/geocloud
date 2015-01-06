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
