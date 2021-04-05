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
		xAdj := x
		if newState {
			xAdj = x.Add(-time.Second)
		}
		// xAdj := x
		if xAdj.Before(now.Add(time.Second)) {
			return // nextTime must never be before now
		}
		if nextTime.IsZero() || xAdj.Before(nextTime) {
			nextTime = x
			state = newState
			// fmt.Println(" == SetIfEarlier", x.String())
		}
	}

	// Start checking the day before to allow for items that span midnight
	t := now.Add(-time.Hour * 24)

	for i := 0; i < 31; i++ { // Check the next 31 days to match a DoW or DoM
		for _, item := range s.Items {
			// Check DOW
			// fmt.Println("Next:", now.String())
			if len(item.DoW) == 0 || item.DoW.Includes(t.Weekday()) {
				// fmt.Println(" - DOW OK", nextTime.Weekday().String(), int(nextTime.Weekday()))
				hr, min, secs := t.Clock()
				nowSecs := HrMinSec{hr, min, secs, false}.Secs()

				// fmt.Println(" - Check before start", nowSecs, item.Start.Secs())
				tStart := time.Date(now.Year(), t.Month(), t.Day(), item.Start.Hour, item.Start.Minute, 0, 0, now.Location())
				if nowSecs < item.Start.Secs() {
					// fmt.Println(" - Start SetIfEarlier ", tStart.String(), true)
					setIfEarlier(tStart, true)
				}
				tEnd := tStart.Add(item.Duration)
				// fmt.Println(" - End SetIfEarlier ", tEnd.String(), false)
				setIfEarlier(tEnd, false)
				// if nextTime.IsZero() || .Before(nextTime) {
				// 	nextTime = x
				// 	state = newState
				// }
			}
		}
		if nextTime.IsZero() {
			t = time.Date(t.Year(), t.Month(), t.Day()+1, 0, 0, 0, 0, now.Location())
		} else {
			break
		}
	}

	// fmt.Printf("----- Next(%s) = %s, %t\n", now, nextTime, state)
	return nextTime, state
}

// Checktime returns true if the given time is within the schedule
func (s *Schedule) CheckTime(now time.Time) bool {
	_, state := s.Next(now)
	return !state
}

func (s *Schedule) Compare(s2 *Schedule) error {
	if len(s.Items) != len(s2.Items) {
		return fmt.Errorf("different number of items: %d vs %d", len(s.Items), len(s2.Items))
	}
	for i, item := range s.Items {
		item2 := s2.Items[i]
		if item.Start.Hour != item2.Start.Hour ||
			item.Start.Minute != item2.Start.Minute ||
			item.Start.Seconds != item2.Start.Seconds {
			return fmt.Errorf("item %d has different start times: %s != %s", i, item.Start.String(), item2.Start.String())
		}

		if len(item.DoW) != len(item2.DoW) {
			return fmt.Errorf("different number of DoW: %d vs %d", len(item.DoW), len(item2.DoW))
		}
		for j, d := range item.DoW {
			if d != item2.DoW[j] {
				return fmt.Errorf("different DoW: %s vs %s", item.DoW.String(), item2.DoW.String())
			}
		}

		if item.Duration != item2.Duration {
			return fmt.Errorf("different duration: %s vs %s", item.Duration, item2.Duration)
		}
	}
	return nil
}

type Item struct {
	Start HrMinSec
	DoW   TokenDoW
	// DoM    []int
	// Month  []int
	// Year   []int
	Duration time.Duration
}

func (item *Item) String() string {
	ret := []string{}

	if len(item.DoW) > 0 {
		ret = append(ret, "On "+item.DoW.String())
	}

	ret = append(ret, "From "+item.Start.String())

	ret = append(ret, "for "+item.Duration.String())

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
	for lineNum, line := range strings.Split(lines, "\n") {
		tokens := getTokens(line)
		// fmt.Printf("\nPARSE %d '%s' [%s]", lineNum, line, tokens.Tokens)
		// fmt.Println("Line ", i+1, ":", tokens)

		if len(tokens.Tokens) == 0 {
			continue
		}

		item := Item{
			Start: HrMinSec{0, 0, 0, false}, // Default start
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
			// fmt.Printf(" ++%T %s++ ", tok, strs)

			switch stage {
			case STAGE_FROM:
				switch v := tok.(type) {
				case HrMinSec:
					// fmt.Printf(" | START TIME %s", v.String())
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

				default:
					return nil, fmt.Errorf("line %d unknown token in from section '%s'", lineNum+1, strs)
				}

			case STAGE_TO:
				switch v := tok.(type) {
				case HrMinSec:
					end := HrMinSec{
						Hour:   v.Hour,
						Minute: v.Minute,
					}
					item.Duration = time.Second * time.Duration(end.Secs()-item.Start.Secs())
					// fmt.Printf(" | END TIME %s [%s] (%s)\n", v.String(), strs, item.Duration)
					if item.Duration < -time.Hour*12 {
						// fmt.Printf(" Negative duration greater 12 hrs - adding 24 hours\n")
						item.Duration += time.Hour * 24

					}
					if item.Duration <= 0 {
						// hasPM := len(strs) > 1 && strs[1] == "pm"
						if end.Hour < 12 && v.NoAMPM {
							// fmt.Printf(" Negative duration - adding 12 hours\n")
							item.Duration += time.Hour * 12
						}
						if item.Duration <= 0 {
							// fmt.Printf(" Negative duration - adding 24 hours\n")
							item.Duration += time.Hour * 24
						}
					}
					// fmt.Printf(" | DURATION %s", item.Duration.String())

				default:
					return nil, fmt.Errorf("line %d unknown token in to section '%s'", lineNum+1, strs)

				}
			}
			// fmt.Printf(" T[%T] ", tok)
		}

		if item.Duration == 0 { // If no duration then set to end of the day
			end := HrMinSec{24, 0, 0, false}
			item.Duration = time.Second * time.Duration(end.Secs()-item.Start.Secs())
		}

		s.Items = append(s.Items, item)
	}
	// fmt.Println()
	return s, nil
}
