package schedule

import (
	"testing"
	"time"
)

func Test_Example(t *testing.T) {
	s, err := Parse("Mon-Fri 9 to 5")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(s.String()) // On Monday, Tuesday, Wednesday, Thursday, Friday From 09:00 until 17:00
	at, state := s.Next(time.Date(2021, 4, 4, 23, 0, 0, 0, time.Local))
	t.Log(at, state) // 021-04-05 09:00:00 +0100 BST true

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
		t.Error("Sun 18:00 should not return true")
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

var tests = []struct {
	Line    string
	Strings []string
	Tokens  []interface{}
}{
	{"09", []string{"09"}, []interface{}{TokenTime{9, 0, 0}}},                      // 0
	{" 09:00", []string{"09:00"}, []interface{}{TokenTime{9, 0, 0}}},               // 1
	{"15:00", []string{"15:00"}, []interface{}{TokenTime{16, 0, 0}}},               // 2
	{"4:00 PM", []string{"4:00", "pm"}, []interface{}{TokenTime{16, 0, 0}}},        // 3
	{"5 PM", []string{"5", "pm"}, []interface{}{TokenTime{16, 0, 0}}},              // 4
	{"23:59", []string{"23:59"}, []interface{}{}},                                  // 5
	{"0905", []string{"0905"}, []interface{}{}},                                    // 6
	{"mon", []string{"mon"}, []interface{}{}},                                      // 7
	{"FrI", []string{"fri"}, []interface{}{}},                                      // 8
	{"mon 9:00 to 16:00", []string{"mon", "9:00", "to", "16:00"}, []interface{}{}}, // 9
	{"mon 6:00   to  18:00 ", []string{"mon", "6:00", "to", "18:00"},
		[]interface{}{}}, // 10
	{"mon 9:00 to 16:00 Friday",
		[]string{"mon", "9:00", "to", "16:00", "friday"},
		[]interface{}{}}, // 11
	{"mon-fri 9:00-16:00", []string{"mon", "-", "fri", "9:00", "-", "16:00"},
		[]interface{}{}}, // 12
	{"mon to 5 pm", []string{"mon", "to", "5", "pm"}, []interface{}{}}, // 13
	{"mon", []string{"mon"}, []interface{}{}},                          // 14
}

func Test_Parse(t *testing.T) {
	for i, test := range tests {
		s, err := Parse(test.Line)
		t.Logf("Test %d: %s [%s]", i, s, test.Line)
		if err != nil {
			t.Error(err)
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
