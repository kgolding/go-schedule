package schedule

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Tokens struct {
	Index  int
	Tokens []string
}

// HrMinSec stores a time of day
type HrMinSec struct {
	Hour    int
	Minute  int
	Seconds int
}

func (t HrMinSec) Secs() int {
	return (t.Hour*60+t.Minute)*60 + t.Seconds
}

func (t HrMinSec) String() string {
	return fmt.Sprintf("%02d:%02d", t.Hour, t.Minute)
}

type TokenDoW []time.Weekday

func (t TokenDoW) Includes(d time.Weekday) bool {
	for _, v := range t {
		if v == d {
			return true
		}
	}
	return false
}

func (t TokenDoW) String() string {
	s := []string{}
	for _, d := range t {
		s = append(s, d.String())
	}
	return strings.Join(s, ", ")
}

type TokenTo string
type TokenFor string
type TokenUntil string

var regTime = regexp.MustCompile(`^(\d{1,2}):{0,1}(\d{2}){0,1}$`)
var dows = []string{"sunday", "monday", "tuesday", "wednesday", "thursday", "friday", "saturday"}

func (t *Tokens) Next() (interface{}, []string, error) {
	startIndex := t.Index
	tok, err := t.TokenAt(t.Index)
	if err != nil {
		return nil, nil, err
	}

	defer func() {
		t.Index++
	}()

	switch v := tok.(type) {
	case TokenDoW: // Check for DOW range
		if n1, err := t.TokenAt(t.Index + 1); err == nil {
			if _, ok := n1.(TokenTo); ok {
				// fmt.Print(" GOT TO ")
				if a, err := t.TokenAt(t.Index + 2); err == nil {
					// fmt.Printf(" GOT %T ", a)
					if n2, ok := a.(TokenDoW); ok {
						// fmt.Print(" GOT DOW ")
						dFrom := time.Weekday(0) // Get last day from v
						if len(v) > 0 {
							dFrom = v[len(v)-1]
						}
						dTo := time.Weekday(6) // Get first day from n2
						if len(n2) > 0 {
							dTo = n2[0]
						}

						// Append days in range between n1 last and n2 first
						for d := dFrom + 1; d < dTo; d++ {
							v = append(v, d)
						}

						// Append to days
						v = append(v, n2...)

						t.Index += 2
						return v, t.Tokens[startIndex : t.Index+1], nil
					}
				}
			}
		}
	}

	return tok, t.Tokens[startIndex : t.Index+1], err
}

func (t *Tokens) TokenAt(index int) (interface{}, error) {
	if index >= len(t.Tokens) {
		return nil, io.EOF
	}
	str := t.Tokens[index]

	switch str {
	case "to", "-":
		return TokenTo("to"), nil
	case "for":
		return TokenFor(str), nil
	case "until":
		return TokenUntil(str), nil
	}

	// Time
	if m := regTime.FindStringSubmatch(str); m != nil {
		item := HrMinSec{}
		item.Hour, _ = strconv.Atoi(m[1])
		item.Minute, _ = strconv.Atoi("0" + m[2])
		if index < len(t.Tokens)-1 {
			switch t.Tokens[index+1] {
			case "am":
				if item.Hour > 12 {
					return nil, errors.New("24 hour clock given can not have AM")
				}
				t.Index++ // Skip token
			case "pm":
				if item.Hour > 12 {
					return nil, errors.New("24 hour clock given can not have AM")
				}
				t.Index++ // Skip token
				item.Hour += 12
			}
		}
		return item, nil
	}

	// DOW
	var dow TokenDoW
	for i, s := range dows {
		if strings.HasPrefix(s, str) {
			dow = append(dow, time.Weekday(i))
		}
	}
	if len(dow) > 0 {
		sort.Slice(dow, func(i, j int) bool {
			return dow[i] < dow[j]
		})
		return dow, nil
	}

	return nil, fmt.Errorf("unknown '%s'", str)
}

func getTokens(line string) Tokens {
	line = strings.ReplaceAll(line, "-", " - ")
	t := Tokens{}
	for _, e := range strings.Split(line, " ") {
		e = strings.TrimSpace(e)
		if e != "" {
			t.Tokens = append(t.Tokens, strings.ToLower(e))
		}
	}
	return t
}
