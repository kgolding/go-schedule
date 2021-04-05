# Go-schedule

** WORK IN PROGRESS **

To-do:

* Allow optional seconds in times e.g. 09:30:12

## Intro

A user friendly definable schedule (think home heating on/off time).

Schedules are defined using near natural language, for example:

* `Mon 9:00 to 17:00`
* `Mon 8 to 5 pm`
* `Mon-Fri 9:00 to 17:30`
* `Sat 22:00 - 02:00`

Typically a schedule will be created from a call to `Parse(string)` with the human reable definition, and a call will 


## Example usage

```
	s, err := Parse("Mon-Fri 9 to 5")
	if err != nil {
		panic(err)
	}
	t.Log(s.String()) // On Monday, Tuesday, Wednesday, Thursday, Friday From 09:00 until 17:00
	
	at, state := s.Next(time.Date(2021, 4, 4, 23, 0, 0, 0, time.Local))
	t.Log(at, state) // 2021-04-05 09:00:00 +0100 BST true

```

