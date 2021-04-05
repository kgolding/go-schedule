package schedule

import (
	"testing"
	"time"
)

// func Test_Example(t *testing.T) {
// 	s, err := Parse("Mon-Fri 9 to 5")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	// t.Log(s.String()) // On Monday, Tuesday, Wednesday, Thursday, Friday From 09:00 until 17:00
// 	at, state := s.Next(time.Date(2021, 4, 4, 23, 0, 0, 0, time.Local))
// 	t.Log(at, state) // 021-04-05 09:00:00 +0100 BST true
// }

func Test_CheckTime(t *testing.T) {
	s, err := Parse("Mon-Fri 9 to 5")
	if err != nil {
		t.Fatal(err)
	}
	checks := []struct {
		time   time.Time
		expect bool
	}{
		{time.Date(2021, 4, 5, 8, 0, 0, 0, time.Local), false},  // 8:00
		{time.Date(2021, 4, 5, 8, 59, 0, 0, time.Local), false}, // 8:59
		{time.Date(2021, 4, 5, 9, 0, 0, 0, time.Local), true},   // 9:00
		{time.Date(2021, 4, 5, 9, 1, 0, 0, time.Local), true},   // 9:01
		{time.Date(2021, 4, 5, 16, 59, 0, 0, time.Local), true}, // 16:59
		{time.Date(2021, 4, 5, 17, 0, 0, 0, time.Local), false}, // 17:00
		{time.Date(2021, 4, 5, 17, 1, 0, 0, time.Local), false}, // 17:01
		{time.Date(2021, 4, 5, 20, 0, 0, 0, time.Local), false}, // 20:00
	}

	for _, v := range checks {
		if result := s.CheckTime(v.time); result != v.expect {
			next, state := s.Next(v.time)
			t.Errorf("%s: expected %t got %t: next() = %s, %t", v.time.String(), v.expect, result, next, state)
		}
	}
}

func Test_One(t *testing.T) {
	s, err := Parse("Mon-Fri 9:00-17:00")
	if err != nil {
		t.Fatal(err)
	}
	// t.Log("\n" + s.String())

	if s.CheckTime(time.Date(2021, 4, 5, 8, 30, 0, 0, time.Local)) {
		t.Error("Mon 8:30 should not return true")
	}
	if !s.CheckTime(time.Date(2021, 4, 5, 9, 30, 0, 0, time.Local)) {
		t.Error("Mon 9:30 should not return false")
	}
	if !s.CheckTime(time.Date(2021, 4, 6, 16, 59, 0, 0, time.Local)) {
		t.Error("Tue 16:59 should not return false")
	}
	if s.CheckTime(time.Date(2021, 4, 4, 18, 0, 0, 0, time.Local)) {
		t.Error("Sun 18:00 should not return true")
	}
	if s.CheckTime(time.Date(2021, 4, 4, 10, 0, 0, 0, time.Local)) {
		t.Error("Sun 10:00 should not return true")
	}
}

func Test_Two(t *testing.T) {
	s, err := Parse("Mon-Fri 9:00-17:00")
	if err != nil {
		t.Fatal(err)
	}
	// t.Log("\n" + s.String())

	MonStart := time.Date(2021, 4, 5, 9, 0, 0, 0, time.Local)
	MonEnd := time.Date(2021, 4, 5, 17, 0, 0, 0, time.Local)

	at, state := s.Next(time.Date(2021, 4, 5, 8, 30, 0, 0, time.Local))
	// t.Log("NEXT()", at, state)
	if at.Sub(MonStart) != time.Duration(0) {
		t.Errorf("A. Expected %s got %s", MonStart, at)
	}
	if !state {
		t.Errorf("A. Incorrect state")
	}

	at, state = s.Next(time.Date(2021, 4, 5, 9, 01, 0, 0, time.Local))
	// t.Log("NEXT()", at, state)
	if at.Sub(MonEnd) != time.Duration(0) {
		t.Errorf("B. Expected %s got %s", MonEnd, at)
	}
	if state {
		t.Errorf("B. Incorrect state")
	}

	at, state = s.Next(time.Date(2021, 4, 4, 0, 01, 0, 0, time.Local)) // Sun
	// t.Log("NEXT()", at, state)
	if at.Sub(MonStart) != time.Duration(0) {
		t.Errorf("C. Expected %s got %s", MonEnd, at)
	}
	if !state {
		t.Errorf("C. Incorrect state")
	}
}

func Test_Three(t *testing.T) {
	s, err := Parse("20:00-06:00")
	if err != nil {
		t.Fatal(err)
	}
	// t.Log("\n" + s.String())

	if s.CheckTime(time.Date(2021, 4, 5, 8, 30, 0, 0, time.Local)) {
		t.Error("Mon 8:30 should not return true")
	}
	if !s.CheckTime(time.Date(2021, 4, 5, 5, 30, 0, 0, time.Local)) {
		t.Error("Mon 5:30 should not return false")
	}
	if s.CheckTime(time.Date(2021, 4, 6, 19, 59, 0, 0, time.Local)) {
		t.Error("Tue 19:59 should not return true")
	}
	if !s.CheckTime(time.Date(2021, 4, 6, 20, 01, 0, 0, time.Local)) {
		t.Error("Tue 20:01 should not return false")
	}
	if s.CheckTime(time.Date(2021, 4, 4, 18, 0, 0, 0, time.Local)) {
		t.Error("Sun 18:00 should not return true")
	}
	if s.CheckTime(time.Date(2021, 4, 4, 10, 0, 0, 0, time.Local)) {
		t.Error("Sun 10:00 should not return true")
	}
}

func createSchedule(startHr int, startMin int, dow TokenDoW, durHrs int) Schedule {
	return Schedule{
		[]Item{
			Item{
				HrMinSec{startHr, startMin, 0, false},
				dow,
				time.Hour * time.Duration(durHrs),
			},
		},
	}
}

var tests = []struct {
	Line    string
	Strings []string
	Expect  Schedule
}{
	{"09", []string{"09"}, createSchedule(9, 0, nil, 15)},                                                                               // 0
	{" 09:00", []string{"09:00"}, createSchedule(9, 0, nil, 15)},                                                                        // 1
	{"15:00", []string{"15:00"}, createSchedule(15, 0, nil, 9)},                                                                         // 2
	{"12:00", []string{"12:00"}, createSchedule(12, 0, nil, 12)},                                                                        // 3
	{"0:00", []string{"0:00"}, createSchedule(0, 0, nil, 24)},                                                                           // 4
	{"23:59 to 00:59", []string{"23:59", "to", "00:59"}, createSchedule(23, 59, nil, 1)},                                                // 5
	{"0905 to 10:05", []string{"0905", "to", "10:05"}, createSchedule(9, 5, nil, 1)},                                                    // 6
	{"mon", []string{"mon"}, createSchedule(0, 0, TokenDoW{1}, 24)},                                                                     // 7
	{"FrI", []string{"fri"}, createSchedule(0, 0, TokenDoW{5}, 24)},                                                                     // 8
	{"mon 9:00 to 16:00", []string{"mon", "9:00", "to", "16:00"}, createSchedule(9, 0, TokenDoW{1}, 7)},                                 // 9
	{"mon 6:00   to  18:00 ", []string{"mon", "6:00", "to", "18:00"}, createSchedule(6, 0, TokenDoW{1}, 12)},                            // 10
	{"mon 9:00 to 16:00", []string{"mon", "9:00", "to", "16:00"}, createSchedule(9, 0, TokenDoW{1}, 7)},                                 // 11
	{"mon-fri 9:00-16:00", []string{"mon", "-", "fri", "9:00", "-", "16:00"}, createSchedule(9, 0, TokenDoW{1, 2, 3, 4, 5}, 7)},         // 12
	{"fri to Sat", []string{"fri", "to", "sat"}, createSchedule(0, 0, TokenDoW{5, 6}, 24)},                                              // 13
	{"mon", []string{"mon"}, createSchedule(0, 0, TokenDoW{1}, 24)},                                                                     // 14
	{"mon 12:00-12:00", []string{"mon", "12:00", "-", "12:00"}, createSchedule(12, 0, TokenDoW{1}, 24)},                                 // 15
	{"Sat 22:00 - 02:00", []string{"sat", "22:00", "-", "02:00"}, createSchedule(22, 0, TokenDoW{6}, 4)},                                // 16
	{"Mon-Fri 9 to 5", []string{"mon", "-", "fri", "9", "to", "5"}, createSchedule(9, 0, TokenDoW{1, 2, 3, 4, 5}, 8)},                   // 17
	{"4:00 PM", []string{"4:00", "pm"}, createSchedule(16, 0, nil, 8)},                                                                  // 18
	{"5 PM", []string{"5", "pm"}, createSchedule(17, 0, nil, 7)},                                                                        // 19
	{"Mon-Fri 9 am to 5 pm", []string{"mon", "-", "fri", "9", "am", "to", "5", "pm"}, createSchedule(9, 0, TokenDoW{1, 2, 3, 4, 5}, 8)}, // 20
}

func Test_Parse(t *testing.T) {
	for i, test := range tests {
		s, err := Parse(test.Line)
		// t.Logf("Test %d: %s [%s]", i, s, test.Line)
		if err != nil {
			t.Errorf("Test %d: %s", i, err)
		} else {
			if err = test.Expect.Compare(s); err != nil {
				t.Errorf("Test %d: %s", i, err)
			}
		}
	}
}

func Test_TokenParse(t *testing.T) {
	for i, test := range tests {
		ret := getTokens(test.Line)
		if len(ret.Tokens) != len(test.Strings) {
			t.Errorf("Test %d: Expected: %v\n     Got: %v", i, test.Strings, ret.Tokens)
		} else {
			for j, e := range ret.Tokens {
				if test.Strings[j] != e {
					t.Errorf("Test %d: Expected: %v\n     Got: %v", i, test.Strings[j], e)
					break
				}
			}
		}
	}
}
