package schedule

import (
	"fmt"
	"io"
	"strings"
	"time"
)

type Schedule struct {
	Items []Item
}

func (s *Schedule) String() string {
	ret := []string{}
	for _, item := range s.Items {
		ret = append(ret, item.String())
	}
	return strings.Join(ret, "\n")
}

// Next returns the next schedule transition time
func (s *Schedule) Next(now time.Time) (time.Time, bool) {
	var nextTime time.Time
	var state bool
	// fmt.Println("======================================", now.String())

	setIfEarlier := func(x time.Time, newState bool) {
		// fmt.Println("- SetIfEarlier", x.String())
		if nextTime.IsZero() || x.Before(nextTime) {
			nextTime = x
			state = newState
		}
	}

	for i := 0; i < 31; i++ { // Check the next 31 days to match a DoW or DoM
		for _, item := range s.Items {
			// Check DOW
			// fmt.Println("Next:", now.String())
			if len(item.DoW) == 0 || item.DoW.Includes(now.Weekday()) {
				// fmt.Println(" - DOW OK", nextTime.Weekday().String(), int(nextTime.Weekday()))
				hr, min, secs := now.Clock()
				nowSecs := TokenTime{hr, min, secs}.Secs() + 1

				// fmt.Println(" - Check before start", nowSecs, item.Start.Secs())
				if nowSecs < item.Start.Secs() {
					setIfEarlier(time.Date(now.Year(), now.Month(), now.Day(), item.Start.Hour, item.Start.Minute, 0, 0, now.Location()), true)
				}
				switch v := item.End.(type) {
				case TokenTime:
					if nowSecs <= v.Secs() {
						setIfEarlier(time.Date(now.Year(), now.Month(), now.Day(), v.Hour, v.Minute, 0, 0, now.Location()), false)
					}
				}
			}
		}
		if nextTime.IsZero() {
			now = time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
		} else {
			break
		}
	}

	return nextTime, state
}

// Checktime returns true if the given time is within the schedule
func (s *Schedule) CheckTime(t time.Time) bool {
	for _, item := range s.Items {
		// Check DOW
		// fmt.Println("Checktime:", t.String())
		if len(item.DoW) == 0 || item.DoW.Includes(t.Weekday()) {
			// fmt.Println("Checktime:", "DOW OK", t.Weekday().String(), int(t.Weekday()))
			hr, min, secs := t.Clock()
			tSecs := TokenTime{hr, min, secs}.Secs()
			// Check is after start time
			if tSecs >= item.Start.Secs() {
				// fmt.Println(" - Checktime:", "Start OK")
				switch v := item.End.(type) {
				case TokenTime:
					if tSecs <= v.Secs() {
						// fmt.Println(" - Checktime:", "End OK")
						return true
						// } else {
						// 	fmt.Println(" - Checktime:", "End FAILED")
					}
				}
				// } else {
				// 	fmt.Println(" - Checktime:", "Start FAILED")
			}
		}
	}
	return false
}

type Item struct {
	Start TokenTime
	DoW   TokenDoW
	// DoM    []int
	// Month  []int
	// Year   []int
	End End
}

func (item *Item) String() string {
	ret := []string{}

	if len(item.DoW) > 0 {
		ret = append(ret, "On "+item.DoW.String())
	}

	ret = append(ret, "From "+item.Start.String())

	switch v := item.End.(type) {
	case TokenTime:
		ret = append(ret, "until "+v.String())
	}

	return strings.Join(ret, " ")
}

type End interface {
}

type EndDuration struct {
}

func New() *Schedule {
	return &Schedule{}
}

const (
	STAGE_FROM = iota
	STAGE_TO
	STAGE_FOR
)

func Parse(lines string) (*Schedule, error) {
	s := New()
	for _, line := range strings.Split(lines, "\n") {
		tokens := getTokens(line)
		// fmt.Println("Line ", i+1, ":", tokens)

		if len(tokens.Tokens) == 0 {
			continue
		}

		item := Item{
			Start: TokenTime{0, 0, 0}, // Default start
		}
		stage := STAGE_FROM

		for {
			tok, strs, err := tokens.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, err
			}
			// fmt.Printf("Parsed token: %#v %s\n", tok, tok)

			switch stage {
			case STAGE_FROM:
				switch v := tok.(type) {
				case TokenTime:
					// fmt.Printf(" | TIME %s", v.String())
					item.Start = v

				case TokenDoW:
					// fmt.Printf(" | DOW %s", v.String())
					item.DoW = v

				case TokenTo:
					// fmt.Print(" | TO")
					stage = STAGE_TO

				case TokenFor:
					// fmt.Print(" | FOR")
					stage = STAGE_FOR
				}

			case STAGE_TO:
				switch v := tok.(type) {
				case TokenTime:
					// fmt.Printf(" | TIME %s", v.String())
					end := TokenTime{
						Hour:   v.Hour,
						Minute: v.Minute,
					}
					if end.Hour < 13 && end.Secs() < item.Start.Secs() {
						fmt.Println(strs)
						if !strings.Contains(strs[0], ":") || len(strs[0]) > 2 || (len(strs) > 1 && strs[1] == "pm") {
							end.Hour += 12
						}
					}
					item.End = end

				}
			}
		}

		if item.End == nil {
			item.End = TokenTime{24, 0, 0}
		}

		s.Items = append(s.Items, item)
	}
	return s, nil
}
