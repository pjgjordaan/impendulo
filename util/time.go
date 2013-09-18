//Copyright (C) 2013  The Impendulo Authors
//
//This library is free software; you can redistribute it and/or
//modify it under the terms of the GNU Lesser General Public
//License as published by the Free Software Foundation; either
//version 2.1 of the License, or (at your option) any later version.
//
//This library is distributed in the hope that it will be useful,
//but WITHOUT ANY WARRANTY; without even the implied warranty of
//MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
//Lesser General Public License for more details.
//
//You should have received a copy of the GNU Lesser General Public
//License along with this library; if not, write to the Free Software
//Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301  USA

package util

import (
	"fmt"
	"strconv"
	"time"
)

const (
	layout = "2006-01-02 15:04:05"
)

//CurMilis returns the time in miliseconds.
func GetMilis(t time.Time) int64 {
	return t.UnixNano() / 1000000
}

//CurMilis returns the current time in miliseconds.
func CurMilis() int64 {
	return GetMilis(time.Now())
}

//GetTime returns an instance of time.Time for the miliseconds provided.
func GetTime(miliseconds int64) time.Time {
	return time.Unix(0, miliseconds*1000000)
}

//Date returns a string representation of the date
//represented by the miliseconds provided.
func Date(miliseconds int64) string {
	return GetTime(miliseconds).Format(layout)
}

//CalcTime converts a time string formatted as yyyymmddhhmmssmmm
//to a time.Time.
func CalcTime(timeStr string) (t time.Time, err error) {
	if len(timeStr) != 17 {
		err = fmt.Errorf("Invalid time string length %d for %s.",
			len(timeStr), timeStr)
		return
	}
	year, err := strconv.Atoi(timeStr[:4])
	if err != nil {
		err = fmt.Errorf("Error %q reading year from %s.",
			err, timeStr)
		return
	}
	if year < 2000 || year > 3000 {
		err = fmt.Errorf("Invalid year %d.", year)
		return
	}
	m, err := strconv.Atoi(timeStr[4:6])
	if err != nil {
		err = fmt.Errorf("Error %q reading month from %s.",
			err, timeStr)
		return
	}
	if m < 1 || m > 12 {
		err = fmt.Errorf("Invalid month %d.", m)
		return
	}
	month := time.Month(m)
	day, err := strconv.Atoi(timeStr[6:8])
	if err != nil {
		err = fmt.Errorf("Error %q reading day from %s.",
			err, timeStr)
		return
	}
	if day < 1 || day > 31 {
		err = fmt.Errorf("Invalid day %d.", day)
		return
	}
	hour, err := strconv.Atoi(timeStr[8:10])
	if err != nil {
		err = fmt.Errorf("Error %q reading hour from %s.",
			err, timeStr)
		return
	}
	if hour < 0 || hour > 24 {
		err = fmt.Errorf("Invalid hour %d.", hour)
		return
	}
	minutes, err := strconv.Atoi(timeStr[10:12])
	if err != nil {
		err = fmt.Errorf("Error %q reading minutes from %s.",
			err, timeStr)
		return
	}
	if minutes < 0 || minutes > 60 {
		err = fmt.Errorf("Invalid minutes %d.", minutes)
		return
	}
	seconds, err := strconv.Atoi(timeStr[12:14])
	if err != nil {
		err = fmt.Errorf("Error %q reading seconds from %s.",
			err, timeStr)
		return
	}
	if seconds < 0 || seconds > 60 {
		err = fmt.Errorf("Invalid seconds %d.", seconds)
		return
	}
	miliseconds, err := strconv.Atoi(timeStr[14:17])
	if err != nil {
		err = fmt.Errorf("Error %q reading miliseconds from %s.",
			err, timeStr)
		return
	}
	if miliseconds < 0 || miliseconds > 1000 {
		err = fmt.Errorf("Invalid miliseconds %d.", miliseconds)
		return
	}
	nanos := miliseconds * 1000000
	loc, err := time.LoadLocation("Local")
	if err != nil {
		err = fmt.Errorf("Error %q loading location.", err)
		return
	}
	t = time.Date(year, month, day, hour, minutes, seconds, nanos, loc)
	return
}
