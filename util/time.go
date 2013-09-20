//Copyright (c) 2013, The Impendulo Authors
//All rights reserved.
//
//Redistribution and use in source and binary forms, with or without modification,
//are permitted provided that the following conditions are met:
//
//  Redistributions of source code must retain the above copyright notice, this
//  list of conditions and the following disclaimer.
//
//  Redistributions in binary form must reproduce the above copyright notice, this
//  list of conditions and the following disclaimer in the documentation and/or
//  other materials provided with the distribution.
//
//THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
//ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
//WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
//DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR
//ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
//(INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
//LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON
//ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
//(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
//SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

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
