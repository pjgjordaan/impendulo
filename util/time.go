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
func GetTime(m int64) time.Time {
	return time.Unix(0, m*1000000)
}

//Date returns a string representation of the date represented by the miliseconds provided.
func Date(m int64) string {
	return GetTime(m).Format(layout)
}

//CalcTime converts a time string formatted as yyyymmddhhmmssmmm to a time.Time.
func CalcTime(s string) (time.Time, error) {
	t := time.Time{}
	if len(s) != 17 {
		return t, fmt.Errorf("invalid time string length %d for %s", len(s), s)
	}
	y, e := strconv.Atoi(s[:4])
	if e != nil {
		return t, &UtilError{s, "reading year", e}
	}
	if y < 1900 || y > 3000 {
		return t, fmt.Errorf("invalid year %d", y)
	}
	m, e := strconv.Atoi(s[4:6])
	if e != nil {
		return t, &UtilError{s, "reading month", e}
	}
	if m < 1 || m > 12 {
		return t, fmt.Errorf("invalid month %d", m)
	}
	d, e := strconv.Atoi(s[6:8])
	if e != nil {
		return t, &UtilError{s, "reading day", e}
	}
	if d < 1 || d > 31 {
		return t, fmt.Errorf("invalid day %d", d)
	}
	h, e := strconv.Atoi(s[8:10])
	if e != nil {
		return t, &UtilError{s, "reading hour", e}
	}
	if h < 0 || h > 24 {
		return t, fmt.Errorf("invalid hour %d", h)
	}
	mi, e := strconv.Atoi(s[10:12])
	if e != nil {
		return t, &UtilError{s, "reading minutes", e}
	}
	if mi < 0 || mi > 60 {
		return t, fmt.Errorf("invalid minutes %d", mi)
	}
	sc, e := strconv.Atoi(s[12:14])
	if e != nil {
		return t, &UtilError{s, "reading seconds", e}
	}
	if sc < 0 || sc > 60 {
		return t, fmt.Errorf("invalid seconds %d", sc)
	}
	ms, e := strconv.Atoi(s[14:17])
	if e != nil {
		return t, &UtilError{s, "reading miliseconds", e}
	}
	if ms < 0 || ms > 1000 {
		return t, fmt.Errorf("invalid miliseconds %d", ms)
	}
	l, e := time.LoadLocation("Local")
	if e != nil {
		return t, &UtilError{"", "loading location", e}
	}
	return time.Date(y, time.Month(m), d, h, mi, sc, ms*1000000, l), nil
}
