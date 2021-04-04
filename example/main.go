package main

import (
	"log"
	"time"

	"github.com/kgolding/go-schedule"
)

func main() {
	s, err := schedule.Parse(`
		0:30 to 0:35
		1 to 2
		3 to 4
		5 to 6
		7 to 8
		10 to 12
	`)
	if err != nil {
		panic(err)
	}
	log.Printf("Schedule: %s\n", s.String())

	CurrentState := s.CheckTime(time.Now())
	for {
		log.Println("State is ", CurrentState)
		nextChange, state := s.Next(time.Now())
		dur := time.Until(nextChange)
		log.Printf("Waiting until %s (%s)", nextChange.String(), dur.String())
		select {
		case <-time.After(dur):
			log.Println("Schedule transitioned to", state)
		}
	}
}
