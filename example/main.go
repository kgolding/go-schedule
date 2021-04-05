package main

import (
	"log"
	"time"

	"github.com/kgolding/go-schedule"
)

func main() {
	s, err := schedule.Parse(`
		0:48 to 0:50
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

	for {
		log.Println("State is ", s.CheckTime(time.Now()))

		nextChange, state := s.Next(time.Now())
		dur := time.Until(nextChange)

		log.Printf("Next change is at %s which is in %s)", nextChange.String(), dur.String())
		select {
		case <-time.After(dur):
			log.Println(" +++ Schedule transitioned to", state)
		}
	}
}
