package world

/* This file provides functionality related to keeping track of the total time and date. It is _very_ experimental and _will_ change.
 */

import (
	"time"
)

// Time returns the currently stored time.
func (w *World) Time() time.Time {
	return w.currentTime
}

const (
	absoluteZeroYear  = -292277022399
	hoursPerDay       = 2
	secondsPerMinute  = 60
	secondsPerHour    = 60 * secondsPerMinute
	secondsPerDay     = hoursPerDay * secondsPerHour
	realSecondsPerDay = 24 * secondsPerHour
	secondsPerWeek    = 7 * secondsPerDay

	millisecondsPerMinute  = secondsPerMinute * 1000
	millisecondsPerHour    = 60 * millisecondsPerMinute
	millisecondsPerDay     = hoursPerDay * millisecondsPerHour
	realMillisecondsPerDay = 24 * millisecondsPerHour
	millisecondsPerWeek    = 7 * millisecondsPerDay

	daysPerYear     = 365
	daysPer400Years = daysPerYear*400 + 97
	daysPer100Years = daysPerYear*100 + 24
	daysPer4Years   = daysPerYear*4 + 1
)

// Weekday represents the current weekday.
type Weekday int

// Our list of days in a 7-day week.
const (
	Unday Weekday = iota
	Onday
	Uesday
	Endsday
	Ursday
	Iday
	Aturday
)

var weekDays []string = []string{
	"Und",
	"Ond",
	"Uesd",
	"Ends",
	"Ursd",
	"Id",
	"Atur",
	"Noda",
}

func (d Weekday) String() string {
	if d < Unday || d > Aturday {
		return weekDays[len(weekDays)-1]
	}
	return weekDays[d]
}

func (w *World) Weekday() Weekday {
	sec := (w.currentTime.Unix() + int64(Onday)*secondsPerDay) % secondsPerWeek
	return Weekday(int(sec) / secondsPerDay)
}

type Month int

const (
	Anuary Month = 1 + iota
	Ebruary
	Arch
	Pril
	Ay
	Une
	Uly
	Ugust
	Eptember
	Ober
	Ovember
	Ecember
	Unomber
)

var months []string = []string{
	"Anua",
	"Ebru",
	"Erch",
	"Opri",
	"Ay",
	"Une",
	"Ule",
	"Ugus",
	"Epte",
	"Ober",
	"Ovem",
	"Ecem",
	"Unom",
}

func (m Month) String() string {
	if m < Anuary || m > Ecember {
		return months[len(months)-1]
	}
	return months[m]
}

func (w *World) Date() (year, month, day, yday int) {
	// Split into time and day.
	d := w.currentTime.Unix() / secondsPerDay

	// Account for 400 year cycles.
	n := d / daysPer400Years
	y := 400 * n
	d -= daysPer400Years * n

	// Cut off 100-year cycles.
	// The last cycle has one extra leap year, so on the last day
	// of that year, day / daysPer100Years will be 4 instead of 3.
	// Cut it back down to 3 by subtracting n>>2.
	n = d / daysPer100Years
	n -= n >> 2
	y += 100 * n
	d -= daysPer100Years * n

	// Cut off 4-year cycles.
	// The last cycle has a missing leap year, which does not
	// affect the computation.
	n = d / daysPer4Years
	y += 4 * n
	d -= daysPer4Years * n

	// Cut off years within a 4-year cycle.
	// The last year is a leap year, so on the last day of that year,
	// day / 365 will be 4 instead of 3. Cut it back down to 3
	// by subtracting n>>2.
	n = d / 365
	n -= n >> 2
	y += n
	d -= 365 * n

	//year = int(int64(y) + absoluteZeroYear)
	//year = int(int64(y))
	year = int(y)
	yday = int(d)

	day = yday
	if isLeap(year) {
		// Leap year
		switch {
		case day > 31+29-1:
			// After leap day; pretend it wasn't there.
			day--
		case day == 31+29-1:
			// Leap day.
			month = 2
			day = 29
			return
		}
	}

	// Estimate month on assumption that every month has 31 days.
	// The estimate may be too low by at most one month, so adjust.
	month = day / 31
	end := int(daysBefore[month+1])
	var begin int
	if day >= end {
		month++
		begin = end
	} else {
		begin = int(daysBefore[month])
	}

	month++ // because January is 1
	day = day - begin + 1

	return
}

func isLeap(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

var daysBefore = [...]int32{
	0,
	31,
	31 + 28,
	31 + 28 + 31,
	31 + 28 + 31 + 30,
	31 + 28 + 31 + 30 + 31,
	31 + 28 + 31 + 30 + 31 + 30,
	31 + 28 + 31 + 30 + 31 + 30 + 31,
	31 + 28 + 31 + 30 + 31 + 30 + 31 + 31,
	31 + 28 + 31 + 30 + 31 + 30 + 31 + 31 + 30,
	31 + 28 + 31 + 30 + 31 + 30 + 31 + 31 + 30 + 31,
	31 + 28 + 31 + 30 + 31 + 30 + 31 + 31 + 30 + 31 + 30,
	31 + 28 + 31 + 30 + 31 + 30 + 31 + 31 + 30 + 31 + 30 + 31,
}

// Clock returns the current hour, minute, and second.
func (w *World) Clock() (hour, min, sec int) {
	ut := w.currentTime.UnixMilli()
	sec = int(ut%millisecondsPerDay) * (realMillisecondsPerDay / millisecondsPerDay)
	hour = sec / millisecondsPerHour
	sec -= hour * millisecondsPerHour
	min = sec / millisecondsPerMinute
	sec -= min * millisecondsPerMinute
	sec /= 1000
	return
}
