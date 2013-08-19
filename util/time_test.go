package util

import(
	"testing"
	"time"
)

func TestGetMilis(t *testing.T){
	t0 := time.Unix(0, 0)
	m0 := GetMilis(t0)
	if m0 != 0{
		t.Errorf("Expected %d but got %d.", 0, m0)
	}
	t1 := time.Unix(10, 0)
	m1 := GetMilis(t1)
	if m1 != 10000{
		t.Errorf("Expected %d but got %d.", 10000, m1)
	}  
	t2 := time.Unix(1000, 121343230)
	m2 := GetMilis(t2)
	if m2 != 1000121{
		t.Errorf("Expected %d but got %d.", 1000121, m2)
	}  
}

func TestGetTime(t *testing.T){
	t0 := time.Unix(0, 0)
	m0 := GetMilis(t0)
	c0 := GetTime(m0)
	if !t0.Equal(c0){
		t.Errorf("Expected %q but got %q.", t0, c0)
	}
	t1 := time.Unix(10, 0)
	m1 := GetMilis(t1)
	c1 := GetTime(m1)
	if !t1.Equal(c1){
		t.Errorf("Expected %q but got %q.", t1, c1)
	}  
	t2 := time.Unix(1000, 121343230)
	t3 := time.Unix(1000, 121000000)
	m2 := GetMilis(t2)
	c2 := GetTime(m2)
	if !t3.Equal(c2){
		t.Errorf("Expected %q but got %q.", t3, c2)
	}  
}

func TestCalcTime(t *testing.T){
	loc, err := time.LoadLocation("Local")
	if err != nil{
		t.Errorf("Error %q loading location.", err)
	}
	tests := map[string] interface{}{
		"hi" : nil, "1111111" : nil,
		"1233333333333333333213123123123" : nil, "1234567891011129h" : nil,
		"20100307123444155" : time.Date(2010, time.Month(03), 07, 12, 34, 44, 155*1000000, loc),
		"20100307123994155" : nil, "30100307123944155" : nil, "20102307123444155" : nil,
		"20100347123444155" : nil, "20100307323444155" : nil, "20100307129444155" : nil,
		"a0100307123944155" : nil, "2010b307123944155" : nil, "201003c7123944155" : nil,
		"201003071d3944155" : nil, "20100307123e44155" : nil, "2010030712394f155" : nil,
		"2010030712394415g" : nil, "20100307123944-50" : nil,  "2010-307123944150" : nil, 
	}
	for str, expected := range tests{
		calc, err := CalcTime(str)
		if expectedTime, ok := expected.(time.Time); ok{
			if err != nil{
				t.Errorf("Unexpected error %q for test %s.", err, str)
			}
			if !expectedTime.Equal(calc){
				t.Errorf("Expected time %q for test %s but got %q.", expectedTime, str, calc)
			}
		} else if err == nil{
			t.Errorf("Expected error for test %s.", str)
		} 
	}
}