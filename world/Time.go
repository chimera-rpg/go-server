package world

/* This file provides functionality related to keeping track of the total time and date. It is _very_ experimental and _will_ change.
 */

import (
	"math"
	"time"
)

// Season represents what you would expect. All 4 in-game seasons occur within one  real world month.
type Season int

// These be our expected seasons.
const (
	Spring Season = iota
	Summer
	Fall
	Winter
)

// String returns the season as a nicely formatted string.
func (s Season) String() string {
	switch s {
	case Spring:
		return "Spring"
	case Summer:
		return "Summer"
	case Fall:
		return "Fall"
	case Winter:
		return "Winter"
	}
	return "Death"
}

// Cycle represents one or more day/night cycle contained with a real world "day".
type Cycle float64

// These consts provide friendly names for a range of 1 to 6 in-game hours to real-life days.
const (
	First Cycle = iota
	Second
	Third
	Fourth
	Fifth
	Sixth
)

// String returns the period of the cycle as a string.
func (c Cycle) String() string {
	switch Cycle(math.Floor(float64(c))) {
	case First:
		return "first"
	case Second:
		return "second"
	case Third:
		return "third"
	case Fourth:
		return "fourth"
	case Fifth:
		return "fifth"
	case Sixth:
		return "sixth"
	}
	return "zeroth"
}

// Period returns the rounded cycle(period) and its remainder(diel).
func (c Cycle) Period() (Cycle, Diel) {
	cyc, diel := math.Modf(float64(c))
	return Cycle(cyc), Diel(diel)
}

// Diel represents the day/night part of the cycle.
type Diel float64

// String returns the human-friendly "shorthand" strings for dawn, day, dusk, and night.
func (d Diel) String() string {
	// Assume "6"(0.25) as AM and "18"(0.75) as PM. Presume dawn/dusk to remain for roughly 1 hour(0.5).
	if d >= 0.25 && d <= 0.70 {
		if d <= 0.30 {
			return "dawn"
		}
		return "light"
	} else if d < 0.25 || d > 0.70 {
		if d >= 0.70 {
			return "dusk"
		}
		return "night"
	}
	return "drugs o'clock"
}

// Diel returns the diel portion of the cycle.
func (c Cycle) Diel() Diel {
	_, diel := c.Period()
	return diel
}

// Time represents the world's current time.
type Time struct {
	realTime             time.Time
	cacheTime            time.Time
	lastUpdate           time.Time // lastUpdate is used to limit heavy updates to a periodic time stored here.
	season               Season
	cycle                Cycle
	year                 int
	month                time.Month
	week                 time.Weekday
	day                  int
	hour, minute, second int
	// Cached properties for Set calls
	lastSeason Season
	lastCycle  Cycle
}

// Set sets the world time to the given "real world" time. It calculates and caches necessary values for the current time, date, cycle, and season.
func (w *Time) Set(t time.Time) (updates []Update) {
	w.realTime = t
	// FIXME: Make the time update check specifiable.
	if t.Sub(w.lastUpdate) >= 30*time.Second {
		w.Ensure()
		// Okay, let's generate our events.
		if w.lastSeason != w.season {
			updates = append(updates, w.season)
			w.lastSeason = w.season
		}
		if w.lastCycle != w.cycle {
			updates = append(updates, w.cycle)
			w.lastCycle = w.cycle
		}
		w.lastUpdate = t
	}
	return
}

// Ensure ensures the current time-related properties have been updated to match the current real time.
func (w *Time) Ensure() {
	if w.realTime == w.cacheTime {
		return
	}
	w.year, w.month, w.day = w.realTime.Date()
	w.week = w.realTime.Weekday()
	w.hour, w.minute, w.second = w.realTime.Clock()

	daysIn := time.Date(w.year, w.month+1, 0, 0, 0, 0, 0, time.UTC).Day()
	w.season = Season(w.day / (daysIn / 4))
	w.cycle = (Cycle(w.hour) + Cycle(w.minute)/60) / 6

	w.year -= 1070

	w.cacheTime = w.realTime
}

// Cycle returns the current cycle.
func (w *Time) Cycle() Cycle {
	w.Ensure()
	return w.cycle
}

// Season returns the current season.
func (w *Time) Season() Season {
	w.Ensure()
	return w.season
}

// Date returns the year, month, and day.
func (w *Time) Date() (year int, month time.Month, day int) {
	w.Ensure()
	return w.year, w.month, w.day
}

// Clock returns the hour, minute, and second.
func (w *Time) Clock() (hour, minute, second int) {
	w.Ensure()
	return w.hour, w.minute, w.second
}
